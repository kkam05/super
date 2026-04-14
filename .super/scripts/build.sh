#!/bin/bash

# Get the directory name of the current working directory
DIR_NAME=$(basename "$(pwd)")

# Create build directory if it doesn't exist
mkdir -p build

# Build the Go binary
go build -o "build/${DIR_NAME}" src/main.go

echo "[SUPER]: Binary built successfully"