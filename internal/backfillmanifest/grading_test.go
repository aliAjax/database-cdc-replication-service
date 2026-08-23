package backfillmanifest

import "testing"

func TestBackfillManifestPlanDoesNotMutateCurrent(t *testing.T) {
	current := make([]string, 2, 4); current[0], current[1] = "users", "orders"; plan := BuildPlan(current, []string{"events"}); plan[0] = "changed"
	if current[0] != "users" || current[1] != "orders" { t.Fatalf("plan shares current backing array: %#v", current) }
}
func TestBackfillManifestStoreSnapshotsTables(t *testing.T) {
	tables := []string{"users"}; s := NewStore(); s.Put(Manifest{ID:"m", Tables:tables}); tables[0] = "orders"; if s.Snapshot()["m"].Tables[0] != "users" { t.Fatal("store shared input slice") }
}
func TestBackfillManifestCacheSnapshotsTables(t *testing.T) {
	tables := []string{"users"}; c := NewCache(); c.Set(Manifest{ID:"m", Tables:tables}); got,_ := c.Load("m"); got.Tables[0] = "orders"; reread,_ := c.Load("m"); if reread.Tables[0] != "users" { t.Fatal("cache returned shared slice") }
}
func TestBackfillManifestRefreshDoesNotMutateCaller(t *testing.T) {
	old := make([]string, 1, 2); old[0] = "users"; manifest := Manifest{Tables: old}; caller := []string{"orders"}; Refresher{}.ReplaceTables(&manifest, caller); old[0] = "events"; if manifest.Tables[0] != "orders" { t.Fatal("refresher reused old backing array") }
}
func TestBackfillManifestRefreshNilManifestIsNoop(t *testing.T) {
	Refresher{}.ReplaceTables(nil, []string{"orders"})
}
func TestBackfillManifestPlanDeduplicatesAcrossInputs(t *testing.T) {
	got := BuildPlan([]string{"users"}, []string{"users", "orders"}); if len(got) != 2 || got[1] != "orders" { t.Fatalf("unexpected plan: %#v", got) }
}
