package backfillmanifest

type Refresher struct{}

func (Refresher) ReplaceTables(manifest *Manifest, tables []string) {
	manifest.Tables = append(manifest.Tables[:0], tables...)
}
