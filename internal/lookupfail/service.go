package lookupfail

type Kind string

const (
	KindMissing Kind = "missing"
	KindSystem  Kind = "system"
)

func Classify(err error) Kind {
	if err == ErrStreamMissing {
		return KindMissing
	}
	return KindSystem
}
