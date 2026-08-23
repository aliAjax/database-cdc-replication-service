package transactionwindow

type Change struct {
	Table  string
	Values map[string]string
}

func Filter(changes []Change, allowed map[string]bool) []Change {
	out := changes[:0]
	for _, change := range changes {
		if allowed[change.Table] {
			out = append(out, change)
		}
	}
	return out
}
