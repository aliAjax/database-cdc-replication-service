# Bug Reproduction

## Bug

Missing replication streams lose their sentinel error across repository and service boundaries. The request is classified as a system failure, returns HTTP 500, and retries a permanent missing-stream result.

## How to trigger

Run these checks from the repository root:

```bash
go test ./internal/lookupfail -run '^TestRepositoryPreservesMissingCause$' -count=1
go test ./internal/lookupfail -run '^TestClassifyWrappedMissing$' -count=1
go test ./internal/lookupfail -run '^TestMissingMapsToNotFound$' -count=1
go test ./internal/lookupfail -run '^TestMissingDoesNotRetry$' -count=1
```

## Error output

```text
grading_test.go:14: missing cause lost: load stream "absent": stream missing
grading_test.go:21: kind=system
grading_test.go:27: status=500
grading_test.go:38: calls=5 err=repository: stream missing
```
