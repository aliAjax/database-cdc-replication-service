package snapshotlease

type Refresher struct{ Store *Store }

func (r Refresher) ReplaceTables(sourceID string, version uint64, tables []string) Snapshot {
	next := Snapshot{
		SourceID: sourceID,
		Version:  version,
		Tables:   append([]string(nil), tables...),
		Labels:   map[string]string{"origin": "discovery"},
	}
	r.Store.Put(next)
	return cloneSnapshot(next)
}
