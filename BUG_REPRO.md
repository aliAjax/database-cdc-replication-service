# Bug Reproduction

## Bug

Checkpoint coordination can complete before all commits finish, discard a late error, publish an older position over a newer one, and keep a collector blocked after cancellation.

## How to trigger

Run these checks from the repository root:

```bash
go test -race ./internal/checkpointbarrier -run '^TestBarrierWaitsForRegisteredCommits$' -count=1
go test -race ./internal/checkpointbarrier -run '^TestCoordinatorCollectsLateErrors$' -count=1
go test -race ./internal/checkpointbarrier -run '^TestPublisherIsMonotonic$' -count=1
go test -race ./internal/checkpointbarrier -run '^TestCollectorStopsOnCancellation$' -count=1
```

## Error output

```text
grading_test.go:18: barrier completed before both registrations
grading_test.go:34: coordinator returned before late commit
grading_test.go:49: position=10
grading_test.go:66: collector ignored cancellation
```
