# Bug Reproduction

## Bug

Batch resource cleanup is delayed or skipped, and deferred transaction and audit operations overwrite or omit earlier failures. This leaks open files and leases while hiding the original operation or flush error.

## How to trigger

Run these checks from the repository root:

```bash
go test ./internal/batchresources -run '^TestFilesClosePerIteration$' -count=1
go test ./internal/batchresources -run '^TestTransactionKeepsOperationError$' -count=1
go test ./internal/batchresources -run '^TestLeaseAlwaysReleases$' -count=1
go test ./internal/batchresources -run '^TestAuditFlushErrorIsJoined$' -count=1
```

## Error output

```text
grading_test.go:30: too many open files
grading_test.go:51: err=commit failed rolled=true committed=true
grading_test.go:67: released=false err=apply failed
grading_test.go:80: err=append failed
```
