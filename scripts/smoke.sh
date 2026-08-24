#!/usr/bin/env bash
set -euo pipefail
base=${BASE_URL:-http://127.0.0.1:8087}
curl -fsS "$base/healthz"
source_json=$(curl -fsS -X POST "$base/api/v1/sources" -H 'content-type: application/json' -d '{"name":"demo","kind":"postgres","dsn":"simulator://source"}')
sid=$(printf '%s' "$source_json" | sed -n 's/.*"id":"\([^"]*\)".*/\1/p')
pipeline_json=$(curl -fsS -X POST "$base/api/v1/pipelines" -H 'content-type: application/json' -d "{\"name\":\"demo-pipeline\",\"source_id\":\"$sid\",\"sink\":\"memory\"}")
pid=$(printf '%s' "$pipeline_json" | sed -n 's/.*"id":"\([^"]*\)".*/\1/p')
curl -fsS -X POST "$base/api/v1/pipelines/$pid/validate"
curl -fsS -X POST "$base/api/v1/pipelines/$pid/start"
curl -fsS "$base/api/v1/pipelines/$pid/lag"
curl -fsS -X POST "$base/api/v1/pipelines/$pid/pause"
printf '\nsmoke ok\n'
