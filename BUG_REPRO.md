# Bug Reproduction

## Bug

Dispatch leases are not returned after use or worker failure. The same failure path commits instead of rolling back and drops operation, rollback, and audit errors, so later batches report an unavailable lease while the earliest processing failure disappears.

## How to trigger

Run these checks from the repository root:

```bash
go test ./internal/dispatchlease -run '^TestDispatchLeaseReleaseReturnsLease$' -count=1
go test ./internal/dispatchlease -run '^TestDispatchLeaseWorkerFailureReleasesLease$' -count=1
go test ./internal/dispatchlease -run '^TestDispatchLeaseFinishRollsBackJoinedFailure$' -count=1
go test ./internal/dispatchlease -run '^TestDispatchLeaseAuditPreservesFailures$' -count=1
```

## Error output

```text
internal/dispatchlease/grading_test.go:44:64: a.Flushed undefined (type *Audit has no field or method Flushed, but does have field flushed)
FAIL github.com/example/cdc-replication/internal/dispatchlease [build failed]
```
