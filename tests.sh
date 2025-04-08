#!/usr/bin/env bash
# Run all tests with verbose output
go test -v ./pkg/...

# Run specific package tests
go test -v ./pkg/crypto
go test -v ./pkg/storage
go test -v ./pkg/cli

# Or use the test script
go run scripts/run_tests.go