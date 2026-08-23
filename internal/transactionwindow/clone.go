package transactionwindow

func Clone(change Change) Change {
	out := change
	out.Values = make(map[string]string, len(change.Values))
	for key, value := range change.Values {
		out.Values[key] = value
	}
	return out
}

func CloneAll(changes []Change) []Change {
	out := make([]Change, len(changes))
	for index, change := range changes {
		out[index] = Clone(change)
	}
	return out
}
