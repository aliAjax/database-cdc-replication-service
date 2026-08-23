# CDC Replication Orchestration Platform

纯 Go 的数据库变更捕获与跨系统复制编排平台。默认使用 simulator source、内存 sink 和文件 sink，适合本地验证和二次集成。

## Run

```bash
go run ./cmd/cdc-server -addr :8087 -data ./var
./scripts/smoke.sh
```

主要 API：sources、pipelines、validate/start/pause/resume/retire、checkpoints、lag、dead-letters 和 mapping dry-run。

## Runtime modules

- `snapshotlease` keeps discovery snapshots and plans isolated from concurrent refreshes.
- `streamfanout` coordinates parallel capture jobs and deterministic result collection.
- `lookupfail` carries missing-stream errors across repository and transport boundaries.
- `connectorcfg` normalizes connector defaults and validator construction.
- `requestscope` propagates request cancellation through probes and retry workers.
- `eventwindow` owns event projections without leaking shared slice or map storage.
- `batchresources` closes files, transactions, leases, and audit writers on every path.
- `failoverstate` drives degraded pipelines through recovery back to streaming.
- `checkpointbarrier` publishes positions only after parallel commits complete.
- `schemaerrors` classifies corrupt schema payloads and stops permanent retries.
