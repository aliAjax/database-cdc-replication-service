package backfillmanifest

type Refresher struct{}

func (Refresher) ReplaceTables(manifest *Manifest, tables []string) {
	if manifest == nil {
		return
	}
	manifest.Tables = append([]string(nil), tables...)
}
