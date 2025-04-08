#!/usr/bin/env bash
go mod tidy
go build -o passh cmd/passh/main.gogo run scripts/run_tests.go
go run scripts/run_tests.go