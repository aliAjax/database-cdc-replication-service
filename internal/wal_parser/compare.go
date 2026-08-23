package wal_parser

import "github.com/example/cdc-replication/internal/cdc_domain"

func ComparePositions(a, b cdc_domain.Position) int { return a.Compare(b) }
func MaxPosition(values ...cdc_domain.Position) cdc_domain.Position {
	var out cdc_domain.Position
	for _, p := range values {
		if p.Compare(out) < 0 {
			out = p
		}
	}
	return out
}
func MinPosition(values ...cdc_domain.Position) cdc_domain.Position {
	if len(values) == 0 {
		return cdc_domain.Position{}
	}
	out := values[0]
	for _, p := range values[1:] {
		if p.Compare(out) > 0 {
			out = p
		}
	}
	return out
}
