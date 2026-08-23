package cdc_domain

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"
)

var idPattern = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]{2,63}$`)

func ValidateSource(s Source) error {
	if !idPattern.MatchString(s.Name) {
		return errors.New("source name must be 3-64 characters")
	}
	if s.Kind != "postgres" && s.Kind != "mysql" && s.Kind != "simulator" {
		return fmt.Errorf("unsupported source kind %q", s.Kind)
	}
	if s.DSN == "" {
		return errors.New("dsn is required")
	}
	if u, e := url.Parse(s.DSN); e != nil || u.Scheme == "" {
		return errors.New("dsn must have a scheme")
	}
	for _, t := range s.Tables {
		if t.Table == "" {
			return errors.New("table name is required")
		}
		if len(t.Masked) > 0 {
			set := map[string]bool{}
			for _, c := range t.Columns {
				set[c] = true
			}
			for _, c := range t.Masked {
				if !set[c] {
					return fmt.Errorf("masked column %s not selected", c)
				}
			}
		}
	}
	return nil
}
func ValidatePipeline(p Pipeline) error {
	if !idPattern.MatchString(p.Name) {
		return errors.New("pipeline name must be 3-64 characters")
	}
	if p.SourceID == "" {
		return errors.New("source id is required")
	}
	if p.Sink != "memory" && p.Sink != "file" && p.Sink != "webhook" && p.Sink != "postgres" && p.Sink != "mysql" {
		return fmt.Errorf("unsupported sink %q", p.Sink)
	}
	if p.Mapping.Version < 0 {
		return errors.New("mapping version cannot be negative")
	}
	for from, to := range p.Mapping.Rename {
		if from == "" || to == "" {
			return errors.New("mapping rename names cannot be empty")
		}
	}
	return nil
}
func StableTables(t []TableSelection) []TableSelection {
	out := append([]TableSelection(nil), t...)
	sort.Slice(out, func(i, j int) bool {
		a := out[i].Schema + "." + out[i].Table
		b := out[j].Schema + "." + out[j].Table
		return a < b
	})
	return out
}
func EventValid(e ChangeEvent) error {
	if e.ID == "" {
		return errors.New("event id is required")
	}
	if e.TxID == "" {
		return errors.New("transaction id is required")
	}
	if e.Table == "" {
		return errors.New("table is required")
	}
	switch e.Operation {
	case OpInsert, OpUpdate, OpDelete, OpDDL:
	default:
		return fmt.Errorf("unsupported operation %q", e.Operation)
	}
	if e.CommitTime.IsZero() {
		return errors.New("commit time is required")
	}
	return nil
}
func CheckpointAdvance(old, new Checkpoint) error {
	if old.PipelineID != "" && old.PipelineID != new.PipelineID {
		return errors.New("checkpoint pipeline mismatch")
	}
	if old.Position.Protocol != "" && new.Position.Compare(old.Position) < 0 {
		return errors.New("checkpoint regression")
	}
	if new.CommittedAt.IsZero() {
		return errors.New("checkpoint timestamp is required")
	}
	return nil
}

type AuditEntry struct {
	ID, Actor, Action, Resource string
	At                          time.Time
	Details                     map[string]string
}

func (a AuditEntry) Valid() error {
	if a.ID == "" || a.Actor == "" || a.Action == "" || a.Resource == "" {
		return errors.New("audit fields are required")
	}
	if a.At.IsZero() {
		return errors.New("audit time is required")
	}
	return nil
}
func StateTerminal(s PipelineState) bool { return s == StateRetired }
func StateActive(s PipelineState) bool   { return s == StateStreaming || s == StateSnapshotting }
func StateName(s PipelineState) string   { return strings.ToLower(string(s)) }
