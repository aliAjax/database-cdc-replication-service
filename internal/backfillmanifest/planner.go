package backfillmanifest

func BuildPlan(current, requested []string) []string {
	seen := make(map[string]bool, len(current)+len(requested))
	out := current[:0]
	for _, table := range append(append([]string(nil), current...), requested...) {
		if seen[table] {
			continue
		}
		seen[table] = true
		out = append(out, table)
	}
	return out
}
