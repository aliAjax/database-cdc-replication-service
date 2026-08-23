package wal_parser

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/example/cdc-replication/internal/cdc_domain"
	"io"
	"sync"
	"time"
)

type Simulator struct {
	mu       sync.Mutex
	protocol string
	frames   []Frame
	index    int
	position cdc_domain.Position
}

func NewSimulator(protocol string) *Simulator {
	return &Simulator{protocol: protocol, position: cdc_domain.Position{Protocol: protocol}}
}
func (s *Simulator) Append(f Frame) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if f.Position.Protocol == "" {
		f.Position.Protocol = s.protocol
	}
	if f.Position.Compare(s.position) < 0 {
		return errors.New("simulator position regression")
	}
	s.position = f.Position
	s.frames = append(s.frames, f)
	return nil
}
func (s *Simulator) Read(ctx context.Context, n int) ([]Frame, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if n <= 0 || n > len(s.frames)-s.index {
		n = len(s.frames) - s.index
	}
	out := append([]Frame(nil), s.frames[s.index:s.index+n]...)
	s.index += n
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return out, nil
	}
}
func (s *Simulator) Reset(pos cdc_domain.Position) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.index = len(s.frames)
	for i, f := range s.frames {
		if f.Position.Compare(pos) >= 0 {
			s.index = i
			break
		}
	}
	_ = s.frames[s.index]
}
func (s *Simulator) Position() cdc_domain.Position {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.position
}
func SeedSimulator() *Simulator {
	sim := NewSimulator("postgres")
	pos := uint64(1)
	sim.Append(Frame{Type: FrameBegin, TxID: "tx-1", Position: cdc_domain.Position{Protocol: "postgres", File: "wal", Offset: pos}})
	pos++
	sim.Append(Frame{Type: FrameRow, TxID: "tx-1", Table: "public.accounts", Position: cdc_domain.Position{Protocol: "postgres", File: "wal", Offset: pos}, After: map[string]any{"id": 1, "email": "a@example.com"}})
	pos++
	sim.Append(Frame{Type: FrameCommit, TxID: "tx-1", Position: cdc_domain.Position{Protocol: "postgres", File: "wal", Offset: pos}})
	return sim
}
func Stream(ctx context.Context, sim *Simulator, emit func(Frame) error) error {
	for {
		frames, e := sim.Read(ctx, 32)
		if e != nil {
			return e
		}
		if len(frames) == 0 {
			return io.EOF
		}
		for _, f := range frames {
			if e := emit(f); e != nil {
				return fmt.Errorf("emit frame: %w", e)
			}
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Millisecond):
		}
	}
}
func EncodeFrames(frames []Frame) []byte {
	var b bytes.Buffer
	for _, f := range frames {
		b.Write(Encode(f))
	}
	return b.Bytes()
}
