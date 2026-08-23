package cursorresume

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
)

var ErrInvalidCursor = errors.New("invalid resume cursor")

type Cursor struct {
	Stream string `json:"stream"`
	Epoch  uint64 `json:"epoch"`
	Offset uint64 `json:"offset"`
}

func EncodeCursor(cursor Cursor) (string, error) {
	payload, err := json.Marshal(cursor)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(payload), nil
}

func DecodeCursor(token string) (Cursor, error) {
	payload, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return Cursor{}, fmt.Errorf("%v: decode token: %v", ErrInvalidCursor, err)
	}
	var cursor Cursor
	if err := json.Unmarshal(payload, &cursor); err != nil {
		return Cursor{}, fmt.Errorf("%v: decode payload: %v", ErrInvalidCursor, err)
	}
	if cursor.Stream == "" {
		return Cursor{}, fmt.Errorf("%w: stream is required", ErrInvalidCursor)
	}
	return cursor, nil
}
