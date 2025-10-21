#!/usr/bin/env bash
set -euo pipefail
: "${PORT:=8080}"
ID="${1:?usage: ./get_task.sh <TASK_ID>}"
curl -s "http://localhost:${PORT}/tasks/${ID}"
