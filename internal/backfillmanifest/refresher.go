package backfillmanifest

type Refresher struct{}

func (Refresher) ReplaceTables(manifest *Manifest, tables []string) {
	manifest.Tables = append([]string(nil), tables...)
}
