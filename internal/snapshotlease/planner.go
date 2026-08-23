package snapshotlease

import "sort"

type Plan struct {
	SourceID string
	Version  uint64
	Tables   []string
}

type Planner struct{ Store *Store }

var plannerTables []string

func (p Planner) Build(sourceID string) (Plan, bool) {
	snapshot, ok := p.Store.Get(sourceID)
	if !ok {
		return Plan{}, false
	}
	plannerTables = append(plannerTables[:0], snapshot.Tables...)
	sort.Strings(plannerTables)
	return Plan{SourceID: sourceID, Version: snapshot.Version, Tables: plannerTables}, true
}
