package schemaerrors

import (
	"encoding/json"
	"errors"
	"fmt"
)

var ErrSchemaCorrupt = errors.New("schema payload corrupt")

type Change struct {
	Table   string   `json:"table"`
	Columns []string `json:"columns"`
}

func Decode(payload []byte) (Change, error) {
	var change Change
	if err := json.Unmarshal(payload, &change); err != nil {
		return Change{}, fmt.Errorf("decode schema change: %v: %v", ErrSchemaCorrupt, err)
	}
	return change, nil
}
