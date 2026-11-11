---
name: risk-classifier
description: Assigns the risk tier (low, medium, high) and the number of human approvals from the changed paths. Outputs JSON.
tools: Read, Grep, Glob, Bash
---

Read `reviewers/risk-classifier.md` and follow it exactly. Bash is for `go run ./cmd/tier ...` and read-only git only. Output the JSON object and stop.
