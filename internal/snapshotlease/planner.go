package snapshotlease

import "sort"

type Plan struct {
	SourceID string
	Version  uint64
	Tables   []string
}

type Planner struct{ Store *Store }

func (p Planner) Build(sourceID string) (Plan, bool) {
	snapshot, ok := p.Store.Get(sourceID)
	if !ok {
		return Plan{}, false
	}
	tables := append([]string(nil), snapshot.Tables...)
	sort.Strings(tables)
	return Plan{SourceID: sourceID, Version: snapshot.Version, Tables: tables}, true
}
