#!/usr/bin/env bash
set -euo pipefail
: "${PORT:=8080}"
curl -i "http://localhost:${PORT}/healthz"
