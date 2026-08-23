# Cursor Resume Error Chain Loss

## Symptom

Cursor decoding and advance validation produce messages that mention an invalid cursor, but callers cannot classify them with `errors.Is`. A wrapped invalid-cursor error is retried instead of stopping. Cancellation and deadline errors returned by the checkpoint store also lose their causes at the `Resumer` boundary. A nil checkpoint store panics in `Resumer.Resume`.

## Reproduction

Run the focused cursor-resume regression commands from the module root. The failing baseline reports lost `ErrInvalidCursor`, repeated calls for a wrapped invalid cursor, and lost `context.Canceled` or `context.DeadlineExceeded` causes. The nil-store path reports:

```text
panic: runtime error: invalid memory address or nil pointer dereference
github.com/example/cdc-replication/internal/cursorresume.(*Resumer).Resume(...)
    internal/cursorresume/service.go:29
```

The fix must preserve error chains with wrapping-aware formatting and classification, reject a nil store without panicking, and leave valid cursor advancement unchanged.
