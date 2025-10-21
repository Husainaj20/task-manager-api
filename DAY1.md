# Day 1 — Quick Guide

## 1) Run server
```bash
make test
make run
```

## 2) Health check
```bash
./scripts/health.sh
```

## 3) Create & fetch a task
```bash
./scripts/create_task.sh           # outputs JSON
./scripts/get_task.sh <TASK_ID>    # paste the id from previous output
```

## 4) Commit suggestions
```bash
git init
git add .
git commit -m "feat: day1 baseline service + health + worker pool"
```
