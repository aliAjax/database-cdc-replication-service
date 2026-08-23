package txapply

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

type Change struct {
	Table string
	Key   string
	Value string
}

type Sink interface {
	Begin(context.Context, string) error
	Apply(context.Context, string, Change) error
	Commit(context.Context, string) error
	Rollback(context.Context, string) error
}

type Transaction struct {
	ID      string
	Changes []Change
}

type Applier struct {
	mu      sync.Mutex
	states  map[string]State
	sink    Sink
	journal Journal
}

func New(sink Sink, journal Journal) *Applier {
	return &Applier{states: make(map[string]State), sink: sink, journal: journal}
}

func (a *Applier) State(txID string) State {
	a.mu.Lock()
	defer a.mu.Unlock()
	state, ok := a.states[txID]
	if !ok {
		return StatePending
	}
	return state
}

func (a *Applier) move(ctx context.Context, txID string, next State, count int) error {
	a.mu.Lock()
	current, ok := a.states[txID]
	if !ok {
		current = StatePending
	}
	if isTerminal(current) {
		a.mu.Unlock()
		return fmt.Errorf("%w: %s is %s", ErrTransactionClosed, txID, current)
	}
	if !CanTransition(current, next) {
		a.mu.Unlock()
		return transitionError(current, next)
	}
	a.mu.Unlock()

	if err := a.journal.Append(ctx, JournalEntry{TxID: txID, State: next, Count: count}); err != nil {
		return fmt.Errorf("journal %s: %w", next, err)
	}

	a.mu.Lock()
	defer a.mu.Unlock()
	latest, ok := a.states[txID]
	if !ok {
		latest = StatePending
	}
	if latest != current {
		return fmt.Errorf("%w: state changed from %s to %s", ErrInvalidTransition, current, latest)
	}
	a.states[txID] = next
	return nil
}

func (a *Applier) Apply(ctx context.Context, tx Transaction) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := a.sink.Begin(ctx, tx.ID); err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	if err := a.move(ctx, tx.ID, StateApplying, 0); err != nil {
		_ = a.sink.Rollback(context.WithoutCancel(ctx), tx.ID)
		return err
	}

	for i, change := range tx.Changes {
		if err := ctx.Err(); err != nil {
			return a.fail(ctx, tx.ID, i, err)
		}
		if err := a.sink.Apply(ctx, tx.ID, change); err != nil {
			return a.fail(ctx, tx.ID, i, fmt.Errorf("apply change %d: %w", i, err))
		}
	}
	if err := a.move(ctx, tx.ID, StatePrepared, len(tx.Changes)); err != nil {
		return a.fail(ctx, tx.ID, len(tx.Changes), err)
	}
	if err := a.sink.Commit(ctx, tx.ID); err != nil {
		return a.fail(ctx, tx.ID, len(tx.Changes), fmt.Errorf("commit transaction: %w", err))
	}
	if err := a.move(ctx, tx.ID, StateCommitted, len(tx.Changes)); err != nil {
		return fmt.Errorf("record commit: %w", err)
	}
	return nil
}

func (a *Applier) fail(ctx context.Context, txID string, count int, cause error) error {
	cleanupCtx := context.WithoutCancel(ctx)
	joined := cause
	if err := a.move(cleanupCtx, txID, StateFailed, count); err != nil && !errors.Is(err, ErrInvalidTransition) {
		joined = errors.Join(joined, err)
	}
	if err := a.sink.Rollback(cleanupCtx, txID); err != nil {
		joined = errors.Join(joined, fmt.Errorf("rollback transaction: %w", err))
	}
	if err := a.move(cleanupCtx, txID, StateRolledBack, count); err != nil {
		joined = errors.Join(joined, err)
	}
	return joined
}
