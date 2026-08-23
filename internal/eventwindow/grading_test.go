package eventwindow

import (
	"testing"

	"github.com/example/cdc-replication/internal/cdc_domain"
)

func event(table, value string) cdc_domain.ChangeEvent {
	return cdc_domain.ChangeEvent{Table: table, After: map[string]any{"value": value}, Metadata: map[string]string{"source": value}}
}

func TestFilterKeepsInputUntouched(t *testing.T) {
	input := []cdc_domain.ChangeEvent{event("a", "first"), event("b", "second")}
	_ = KeepTables(input, map[string]bool{"b": true})
	if input[0].Table != "a" {
		t.Fatalf("input overwritten: %#v", input)
	}
}

func TestProjectionOwnsNestedMaps(t *testing.T) {
	input := event("a", "first")
	copy := CloneEvent(input)
	copy.After["value"] = "changed"
	copy.Metadata["source"] = "changed"
	if input.After["value"] != "first" || input.Metadata["source"] != "first" {
		t.Fatalf("nested maps escaped: %#v", input)
	}
}

func TestMergeDoesNotAliasWindows(t *testing.T) {
	window := []cdc_domain.ChangeEvent{event("a", "first")}
	merged := Merge(window)
	merged[0].After["value"] = "changed"
	if window[0].After["value"] != "first" {
		t.Fatalf("merge aliased input: %#v", window)
	}
}

func TestRetentionSnapshotStaysStable(t *testing.T) {
	input := []cdc_domain.ChangeEvent{event("a", "first")}
	var retention Retention
	retention.Append(input)
	input[0].Table = "changed"
	got := retention.Latest()
	got[0].Table = "again"
	if stable := retention.Latest(); stable[0].Table != "a" {
		t.Fatalf("retention changed: %#v", stable)
	}
}
