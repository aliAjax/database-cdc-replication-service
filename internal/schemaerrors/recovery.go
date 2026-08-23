package schemaerrors

import "context"

func Recover(ctx context.Context, attempts int, operation func() error) error {
	var last error
	for attempt := 0; attempt < attempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		last = operation()
		if last == nil {
			return last
		}
	}
	return last
}
