package observability

import (
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

type Counter struct{ value atomic.Uint64 }

func (c *Counter) Inc()         { c.value.Add(1) }
func (c *Counter) Add(n uint64) { c.value.Add(n) }
func (c *Counter) Get() uint64  { return c.value.Load() }

type Gauge struct{ value atomic.Int64 }

func (g *Gauge) Set(n int64) { g.value.Store(n) }
func (g *Gauge) Get() int64  { return g.value.Load() }

type Histogram struct {
	mu     sync.Mutex
	values []float64
}

func (h *Histogram) Observe(v float64) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.values = append(h.values, v)
}
func (h *Histogram) Snapshot() []float64 {
	h.mu.Lock()
	defer h.mu.Unlock()
	return append([]float64(nil), h.values...)
}

type Registry struct {
	Counters   map[string]*Counter
	Gauges     map[string]*Gauge
	Histograms map[string]*Histogram
	mu         sync.Mutex
}

func New() *Registry {
	return &Registry{Counters: map[string]*Counter{}, Gauges: map[string]*Gauge{}, Histograms: map[string]*Histogram{}}
}
func (r *Registry) Counter(name string) *Counter {
	r.mu.Lock()
	defer r.mu.Unlock()
	if x := r.Counters[name]; x != nil {
		return x
	}
	x := &Counter{}
	r.Counters[name] = x
	return x
}
func (r *Registry) Gauge(name string) *Gauge {
	r.mu.Lock()
	defer r.mu.Unlock()
	if x := r.Gauges[name]; x != nil {
		return x
	}
	x := &Gauge{}
	r.Gauges[name] = x
	return x
}
func (r *Registry) Histogram(name string) *Histogram {
	r.mu.Lock()
	defer r.mu.Unlock()
	if x := r.Histograms[name]; x != nil {
		return x
	}
	x := &Histogram{}
	r.Histograms[name] = x
	return x
}
func (r *Registry) Handler(w http.ResponseWriter, _ *http.Request) {
	r.mu.Lock()
	defer r.mu.Unlock()
	w.Header().Set("content-type", "text/plain; version=0.0.4")
	for n, c := range r.Counters {
		fmt.Fprintf(w, "cdc_%s %d\n", n, c.Get())
	}
	for n, g := range r.Gauges {
		fmt.Fprintf(w, "cdc_%s %d\n", n, g.Get())
	}
	for n, h := range r.Histograms {
		v := h.Snapshot()
		if len(v) > 0 {
			var sum float64
			for _, x := range v {
				sum += x
			}
			fmt.Fprintf(w, "cdc_%s_count %d\ncdc_%s_sum %f\n", n, len(v), n, sum)
		}
	}
}

type Clock interface{ Now() time.Time }
type RealClock struct{}

func (RealClock) Now() time.Time { return time.Now().UTC() }
