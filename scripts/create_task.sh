#!/usr/bin/env bash
set -euo pipefail
: "${PORT:=8080}"
KEY="${1:-abc123}"
PAYLOAD='{"type":"echo","payload":{"msg":"hello from day1"}}'
curl -s -X POST "http://localhost:${PORT}/tasks"   -H "Content-Type: application/json"   -H "Idempotency-Key: ${KEY}"   -d "${PAYLOAD}"
