package eventwindow

import "github.com/example/cdc-replication/internal/cdc_domain"

func Merge(windows ...[]cdc_domain.ChangeEvent) []cdc_domain.ChangeEvent {
	size := 0
	for _, window := range windows {
		size += len(window)
	}
	out := make([]cdc_domain.ChangeEvent, 0, size)
	for _, window := range windows {
		for _, event := range window {
			out = append(out, CloneEvent(event))
		}
	}
	return out
}
