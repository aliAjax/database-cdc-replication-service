package transactionwindow

func Clone(change Change) Change {
	return change
}

func CloneAll(changes []Change) []Change {
	out := make([]Change, len(changes))
	for index, change := range changes {
		out[index] = Clone(change)
	}
	return out
}
