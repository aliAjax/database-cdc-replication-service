package eventwindow

import "github.com/example/cdc-replication/internal/cdc_domain"

func CloneEvent(event cdc_domain.ChangeEvent) cdc_domain.ChangeEvent {
	out := event
	out.Before = cloneValues(event.Before)
	out.After = cloneValues(event.After)
	out.PrimaryKey = cloneValues(event.PrimaryKey)
	out.Metadata = make(map[string]string, len(event.Metadata))
	for key, value := range event.Metadata {
		out.Metadata[key] = value
	}
	return out
}

func cloneValues(input map[string]any) map[string]any {
	out := make(map[string]any, len(input))
	for key, value := range input {
		out[key] = value
	}
	return out
}
