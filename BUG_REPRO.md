# Bug Reproduction

## Bug

Failover recovery rejects required transitions, mutates state before an invalid transition is rejected, leaves successful recovery in an intermediate state, and omits recovering streams from active results.

## How to trigger

Run these checks from the repository root:

```bash
go test ./internal/failoverstate -run '^TestRecoveryTransitionsAreAllowed$' -count=1
go test ./internal/failoverstate -run '^TestInvalidMoveKeepsState$' -count=1
go test ./internal/failoverstate -run '^TestWorkerPublishesTerminalState$' -count=1
go test ./internal/failoverstate -run '^TestRecoveringRemainsActive$' -count=1
```

## Error output

```text
grading_test.go:10: invalid failover transition degraded -> recovering
grading_test.go:23: state=recovering
grading_test.go:31: invalid failover transition degraded -> recovering
grading_test.go:41: active=[]
```
