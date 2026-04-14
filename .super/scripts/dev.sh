#!/bin/bash

# Create build directory if it doesn't exist
mkdir -p build

# Build the Go binary
go run src/main.go

echo "[SUPER]: Binary ran successfully"