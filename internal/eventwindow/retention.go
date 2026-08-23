package eventwindow

import "github.com/example/cdc-replication/internal/cdc_domain"

type Retention struct{ windows [][]cdc_domain.ChangeEvent }

func (r *Retention) Append(events []cdc_domain.ChangeEvent) {
	r.windows = append(r.windows, Merge(events))
}

func (r *Retention) Latest() []cdc_domain.ChangeEvent {
	if len(r.windows) == 0 {
		return nil
	}
	return Merge(r.windows[len(r.windows)-1])
}
