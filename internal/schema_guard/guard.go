package schema_guard

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

type Column struct {
	Name, Type string
	Nullable   bool
}
type Schema struct {
	Table       string
	Version     int
	Columns     []Column
	Fingerprint string
}
type Drift struct {
	Table      string
	From, To   Schema
	Breaking   bool
	Reasons    []string
	DetectedAt time.Time
}

func Fingerprint(s Schema) string {
	h := sha256.New()
	sort.Slice(s.Columns, func(i, j int) bool { return s.Columns[i].Name < s.Columns[j].Name })
	for _, c := range s.Columns {
		fmt.Fprintf(h, "%s:%s:%t;", c.Name, c.Type, c.Nullable)
	}
	return hex.EncodeToString(h.Sum(nil))
}
func Compare(a, b Schema) Drift {
	d := Drift{Table: b.Table, From: a, To: b, DetectedAt: time.Now().UTC()}
	old := map[string]Column{}
	for _, c := range a.Columns {
		old[c.Name] = c
	}
	for _, c := range b.Columns {
		if x, ok := old[c.Name]; !ok {
			d.Reasons = append(d.Reasons, "column added: "+c.Name)
		} else if !strings.EqualFold(x.Type, c.Type) {
			d.Breaking = true
			d.Reasons = append(d.Reasons, "type changed: "+c.Name)
		}
		delete(old, c.Name)
	}
	for n := range old {
		d.Breaking = true
		d.Reasons = append(d.Reasons, "column removed: "+n)
	}
	return d
}

type Guard struct {
	mu       sync.RWMutex
	schemas  map[string]Schema
	drifts   []Drift
	approved map[string]bool
}

func New() *Guard { return &Guard{schemas: map[string]Schema{}, approved: map[string]bool{}} }
func (g *Guard) Observe(s Schema) Drift {
	g.mu.Lock()
	defer g.mu.Unlock()
	s.Fingerprint = Fingerprint(s)
	old, ok := g.schemas[s.Table]
	g.schemas[s.Table] = s
	if !ok {
		return Drift{Table: s.Table, From: old, To: s, DetectedAt: time.Now().UTC()}
	}
	d := Compare(old, s)
	if len(d.Reasons) > 0 {
		g.drifts = append(g.drifts, d)
	}
	return d
}
func (g *Guard) Approve(table string) { g.mu.Lock(); defer g.mu.Unlock(); g.approved[table] = true }
func (g *Guard) Allowed(table string) bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.approved[table] || len(g.drifts) == 0
}
func (g *Guard) Drifts() []Drift {
	return append([]Drift(nil), g.drifts...)
}
