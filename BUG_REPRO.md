# Bug Reproduction

## Bug

Request cancellation and deadlines are replaced while connection probes and retry workers run. A timed-out probe keeps working, retries ignore cancellation, and a later validation does not retain its own request lifetime.

## How to trigger

Run these checks from the repository root:

```bash
go test ./internal/requestscope -run '^TestBudgetReachesProbe$' -count=1
go test ./internal/requestscope -run '^TestClientUsesCurrentContext$' -count=1
go test ./internal/requestscope -run '^TestRetryStopsOnAbortSignal$' -count=1
go test ./internal/requestscope -run '^TestSecondValidationUsesOwnDeadline$' -count=1
```

## Error output

```text
grading_test.go:20: err=budget did not arrive
grading_test.go:29: client ignored current request context
grading_test.go:42: calls=5
grading_test.go:51: manager replaced canceled request context
```
