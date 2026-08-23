# Bug Reproduction

## Bug

An expired replication-slot lease remains active because the activity check ignores its expiry time. A stale handle from the same owner can renew or authorize writes after a newer generation takes over. When the keeper receives cancellation, the release path uses the canceled context, so the lease can remain in the store.

## Trigger

Acquire a lease, advance the clock beyond its TTL, and attempt takeover. Then cancel a running keeper and inspect the slot. Finally reacquire the same slot and use the old lease handle for renewal and fencing.

## Observed error

The new worker receives `replication slot lease is held`; the old handle renews successfully and the fence authorizes its write. After cancellation, the lease is still present.
