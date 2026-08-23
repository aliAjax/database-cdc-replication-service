# Backfill Manifest Slice Aliasing

## Symptom

The backfill manifest implementation shares the backing arrays of table-name slices across planning, storage, cache, and refresh boundaries. Mutating an input slice, a planned result, or a value returned by the cache can therefore change another component's state. Calling `Refresher.ReplaceTables` with a nil manifest also panics with a nil pointer dereference.

## Reproduction

Run the focused backfill manifest regression commands from the module root. The failing baseline reports slice state changing after the caller or returned value is mutated, and the nil-manifest case reports:

```text
panic: runtime error: invalid memory address or nil pointer dereference
github.com/example/cdc-replication/internal/backfillmanifest.Refresher.ReplaceTables(...)
```

The fix must allocate independent table slices at each ownership boundary, preserve BuildPlan's stable de-duplication, and return safely for a nil destination manifest.
