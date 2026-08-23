package lookupfail

import "errors"

type Kind string

const (
	KindMissing Kind = "missing"
	KindSystem  Kind = "system"
)

func Classify(err error) Kind {
	if errors.Is(err, ErrStreamMissing) {
		return KindMissing
	}
	return KindSystem
}
