package cursorresume

import "fmt"

func ValidateAdvance(current, next Cursor) error {
	if current.Stream != "" && current.Stream != next.Stream {
		return fmt.Errorf("stream changed from %q to %q", current.Stream, next.Stream)
	}
	if next.Epoch < current.Epoch {
		return fmt.Errorf("epoch regressed from %d to %d", current.Epoch, next.Epoch)
	}
	if next.Epoch == current.Epoch && next.Offset < current.Offset {
		return fmt.Errorf("offset regressed from %d to %d", current.Offset, next.Offset)
	}
	return nil
}
