package lookupfail

import "context"

func Retry(ctx context.Context, attempts int, operation func() error) error {
	if attempts < 1 {
		attempts = 1
	}
	var last error
	for i := 0; i < attempts; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		last = operation()
		if last == nil || Classify(last) == KindMissing {
			return last
		}
	}
	return last
}
