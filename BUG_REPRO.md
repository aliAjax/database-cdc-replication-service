# Bug Reproduction

## Bug

Event-window filtering, projection, merging, and retention keep aliases to caller-owned slices or nested maps. Reusing or mutating a later batch can therefore overwrite data already stored in an older window.

## How to trigger

Run these checks from the repository root:

```bash
go test ./internal/eventwindow -run '^TestFilterKeepsInputUntouched$' -count=1
go test ./internal/eventwindow -run '^TestProjectionOwnsNestedMaps$' -count=1
go test ./internal/eventwindow -run '^TestMergeDoesNotAliasWindows$' -count=1
go test ./internal/eventwindow -run '^TestRetentionSnapshotStaysStable$' -count=1
```

## Error output

```text
grading_test.go:17: input overwritten
grading_test.go:27: nested maps escaped
grading_test.go:36: merge aliased input
grading_test.go:48: retention changed
```
