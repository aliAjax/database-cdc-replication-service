package eventwindow

import "github.com/example/cdc-replication/internal/cdc_domain"

func CloneEvent(event cdc_domain.ChangeEvent) cdc_domain.ChangeEvent {
	return event
}

func cloneValues(input map[string]any) map[string]any {
	out := make(map[string]any, len(input))
	for key, value := range input {
		out[key] = value
	}
	return out
}
