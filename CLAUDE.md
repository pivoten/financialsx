# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a new Go project called "Pivoten Financials X - Enhanced Legacy" that is currently in initial setup phase.

## Commands

Since this is a new project without established build tools yet, here are common Go commands to use:

- **Initialize Go module**: `go mod init github.com/username/financialsx` (replace with your actual module path)
- **Run Go code**: `go run .` or `go run main.go`
- **Build**: `go build -o financialsx`
- **Test**: `go test ./...`
- **Format code**: `go fmt ./...`
- **Lint**: `golangci-lint run` (after installing golangci-lint)

## Project Structure

The project is currently empty. When implementing features, consider following standard Go project layout:

- `/cmd` - Main applications for this project
- `/internal` - Private application and library code
- `/pkg` - Library code that's ok to use by external applications
- `/api` - API definitions (OpenAPI/Swagger specs, Protocol Buffers, etc)
- `/configs` - Configuration file templates or default configs
- `/test` - Additional external test apps and test data

## Development Notes

- This appears to be a financial application based on the project name
- No existing code structure or dependencies are present yet