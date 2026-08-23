package transactionwindow

func Merge(left, right []Change) []Change {
	out := make([]Change, 0, len(left)+len(right))
	for _, change := range left {
		out = append(out, Clone(change))
	}
	for _, change := range right {
		out = append(out, Clone(change))
	}
	return out
}
