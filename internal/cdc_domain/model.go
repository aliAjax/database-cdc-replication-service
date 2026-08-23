package cdc_domain

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

type PipelineState string

const (
	StateDraft        PipelineState = "draft"
	StateValidating   PipelineState = "validating"
	StateSnapshotting PipelineState = "snapshotting"
	StateStreaming    PipelineState = "streaming"
	StatePaused       PipelineState = "paused"
	StateDegraded     PipelineState = "degraded"
	StateFailed       PipelineState = "failed"
	StateRetired      PipelineState = "retired"
)

type Operation string

const (
	OpInsert Operation = "insert"
	OpUpdate Operation = "update"
	OpDelete Operation = "delete"
	OpDDL    Operation = "ddl"
)

type Position struct {
	Protocol string `json:"protocol"`
	File     string `json:"file,omitempty"`
	Offset   uint64 `json:"offset"`
	Epoch    uint64 `json:"epoch"`
}

func (p Position) Compare(o Position) int {
	if p.Protocol != o.Protocol {
		return strings.Compare(p.Protocol, o.Protocol)
	}
	if p.Offset != o.Offset {
		if p.Offset < o.Offset {
			return -1
		}
		return 1
	}
	if p.Epoch != o.Epoch {
		if p.Epoch < o.Epoch {
			return -1
		}
		return 1
	}
	if p.File != o.File {
		return strings.Compare(p.File, o.File)
	}
	return 0
}
func (p Position) String() string {
	return fmt.Sprintf("%s/%s/%d/%d", p.Protocol, p.File, p.Offset, p.Epoch)
}
func ParsePosition(s string) Position {
	var p Position
	fmt.Sscanf(s, "%s/%s/%d/%d", &p.Protocol, &p.File, &p.Offset, &p.Epoch)
	return p
}

type ChangeEvent struct {
	ID            string            `json:"id"`
	TxID          string            `json:"tx_id"`
	Position      Position          `json:"position"`
	CommitTime    time.Time         `json:"commit_time"`
	Table         string            `json:"table"`
	PrimaryKey    map[string]any    `json:"primary_key"`
	Before        map[string]any    `json:"before,omitempty"`
	After         map[string]any    `json:"after,omitempty"`
	SchemaVersion int               `json:"schema_version"`
	Operation     Operation         `json:"operation"`
	Metadata      map[string]string `json:"metadata,omitempty"`
}

func NewEvent(tx string, pos Position, table string, op Operation, before, after map[string]any) ChangeEvent {
	e := ChangeEvent{TxID: tx, Position: pos, CommitTime: time.Now().UTC(), Table: table, Before: before, After: after, Operation: op, SchemaVersion: 1}
	e.ID = e.IdempotencyKey()
	return e
}
func (e ChangeEvent) IdempotencyKey() string {
	h := sha256.New()
	fmt.Fprintf(h, "%s|%s|%s|%s", e.TxID, e.Position.String(), e.Table, e.Operation)
	return hex.EncodeToString(h.Sum(nil))
}

type Source struct {
	ID        string           `json:"id"`
	Name      string           `json:"name"`
	Kind      string           `json:"kind"`
	DSN       string           `json:"dsn"`
	Tables    []TableSelection `json:"tables"`
	CreatedAt time.Time        `json:"created_at"`
}
type TableSelection struct {
	Schema  string   `json:"schema"`
	Table   string   `json:"table"`
	Columns []string `json:"columns,omitempty"`
	Masked  []string `json:"masked,omitempty"`
}
type Mapping struct {
	Version int               `json:"version"`
	Rename  map[string]string `json:"rename,omitempty"`
	Cast    map[string]string `json:"cast,omitempty"`
	Mask    []string          `json:"mask,omitempty"`
	Filter  string            `json:"filter,omitempty"`
}
type Pipeline struct {
	ID        string        `json:"id"`
	Name      string        `json:"name"`
	SourceID  string        `json:"source_id"`
	Sink      string        `json:"sink"`
	Mapping   Mapping       `json:"mapping"`
	State     PipelineState `json:"state"`
	UpdatedAt time.Time     `json:"updated_at"`
	CreatedAt time.Time     `json:"created_at"`
	Error     string        `json:"error,omitempty"`
}

func (p *Pipeline) Transition(next PipelineState) error {
	allowed := map[PipelineState][]PipelineState{StateDraft: {StateValidating, StateRetired}, StateValidating: {StateSnapshotting, StateStreaming, StateFailed, StateDraft}, StateSnapshotting: {StateStreaming, StatePaused, StateFailed}, StateStreaming: {StatePaused, StateDegraded, StateFailed, StateRetired}, StatePaused: {StateStreaming, StateSnapshotting, StateRetired, StateFailed}, StateDegraded: {StateStreaming, StatePaused, StateFailed}, StateFailed: {StateDraft, StateRetired}, StateRetired: {}}
	for _, v := range allowed[p.State] {
		if v == next {
			p.State = next
			p.UpdatedAt = time.Now().UTC()
			return nil
		}
	}
	return fmt.Errorf("invalid transition %s -> %s", p.State, next)
}

type Checkpoint struct {
	PipelineID  string    `json:"pipeline_id"`
	Position    Position  `json:"position"`
	TxID        string    `json:"tx_id"`
	CommittedAt time.Time `json:"committed_at"`
	EventCount  uint64    `json:"event_count"`
}
type DeadLetter struct {
	ID         string      `json:"id"`
	PipelineID string      `json:"pipeline_id"`
	Event      ChangeEvent `json:"event"`
	Reason     string      `json:"reason"`
	Attempts   int         `json:"attempts"`
	CreatedAt  time.Time   `json:"created_at"`
	ReplayedAt *time.Time  `json:"replayed_at,omitempty"`
}

var ErrNotFound = errors.New("not found")

type Registry struct {
	mu          sync.RWMutex
	Sources     map[string]Source
	Pipelines   map[string]Pipeline
	Checkpoints map[string]Checkpoint
	DeadLetters map[string][]DeadLetter
}

func NewRegistry() *Registry {
	return &Registry{Sources: map[string]Source{}, Pipelines: map[string]Pipeline{}, Checkpoints: map[string]Checkpoint{}, DeadLetters: map[string][]DeadLetter{}}
}
func (r *Registry) PutSource(s Source) { r.mu.Lock(); defer r.mu.Unlock(); r.Sources[s.ID] = s }
func (r *Registry) GetSource(id string) (Source, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.Sources[id]
	if !ok {
		return Source{}, ErrNotFound
	}
	return s, nil
}
func (r *Registry) ListSources() []Source {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Source, 0, len(r.Sources))
	for _, s := range r.Sources {
		out = append(out, s)
	}
	return out
}
func (r *Registry) PutPipeline(p Pipeline) { r.mu.Lock(); defer r.mu.Unlock(); r.Pipelines[p.ID] = p }
func (r *Registry) GetPipeline(id string) (Pipeline, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.Pipelines[id]
	if !ok {
		return Pipeline{}, ErrNotFound
	}
	return p, nil
}
func (r *Registry) ListPipelines() []Pipeline {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Pipeline, 0, len(r.Pipelines))
	for _, p := range r.Pipelines {
		out = append(out, p)
	}
	return out
}
func (r *Registry) SetCheckpoint(c Checkpoint) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Checkpoints[c.PipelineID] = c
}
func (r *Registry) GetCheckpoint(id string) (Checkpoint, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.Checkpoints[id]
	if !ok {
		return Checkpoint{}, ErrNotFound
	}
	return c, nil
}
func (r *Registry) AddDeadLetter(d DeadLetter) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.DeadLetters[d.PipelineID] = append(r.DeadLetters[d.PipelineID], d)
}
func (r *Registry) GetDeadLetters(id string) []DeadLetter {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return append([]DeadLetter(nil), r.DeadLetters[id]...)
}
