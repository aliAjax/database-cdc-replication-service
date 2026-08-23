package eventwindow

import "github.com/example/cdc-replication/internal/cdc_domain"

func KeepTables(events []cdc_domain.ChangeEvent, allowed map[string]bool) []cdc_domain.ChangeEvent {
	out := events[:0]
	for _, event := range events {
		if allowed[event.Table] {
			out = append(out, event)
		}
	}
	return out
}
