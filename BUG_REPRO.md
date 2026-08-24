# Decoder Registry Nil-State Reproduction

## Bug

The decoder registry accepts several unusable nil states. A default catalog and a zero-value registry contain nil maps, while a disabled provider can be returned as a non-nil interface whose dynamic pointer is nil. The shared validation path does not reject that typed-nil provider before calling it.

## Trigger

Run the decoder registry acceptance checks against the bug snapshot. The failures are triggered by writing to the default catalog, creating a disabled provider, registering a driver through a zero-value registry, and validating a typed-nil provider.

## Observed errors

```text
default catalog has a nil driver map
disabled provider is a non-nil interface: *decoderregistry.StaticProvider
panic: assignment to entry in nil map
panic: runtime error: invalid memory address or nil pointer dereference
```
