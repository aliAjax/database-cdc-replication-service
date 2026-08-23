package transactionwindow

func Merge(left, right []Change) []Change {
	return append(left, right...)
}
