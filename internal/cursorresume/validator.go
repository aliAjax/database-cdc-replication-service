package cursorresume

import "fmt"

func ValidateAdvance(current, next Cursor) error {
	if current.Stream != "" && current.Stream != next.Stream {
		return fmt.Errorf("%v: stream changed from %q to %q", ErrInvalidCursor, current.Stream, next.Stream)
	}
	if next.Epoch < current.Epoch {
		return fmt.Errorf("%v: epoch regressed from %d to %d", ErrInvalidCursor, current.Epoch, next.Epoch)
	}
	if next.Epoch == current.Epoch && next.Offset < current.Offset {
		return fmt.Errorf("%v: offset regressed from %d to %d", ErrInvalidCursor, current.Offset, next.Offset)
	}
	return nil
}
