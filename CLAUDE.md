# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Pivoten Financials X - Enhanced Legacy is a modern desktop companion app to the legacy Visual FoxPro Accounting Manager. It uses Wails with a Go backend + React frontend, mines DBF data and persists derived data in SQLite, and supports new reporting features like State Reporting.

## Commands

### Development
- **Run in dev mode**: `cd desktop && wails dev`
- **Build app**: `cd desktop && wails build`
- **Run tests**: `go test ./...`
- **Format code**: `go fmt ./...`
- **Lint**: `golangci-lint run`

### Frontend (from desktop/frontend directory)
- **Install dependencies**: `npm install`
- **Build frontend**: `npm run build`

## Project Structure

```
financialsx/
├── desktop/                  # Wails desktop application
│   ├── main.go              # Entry point with Wails setup
│   ├── go.mod               # Module: github.com/pivoten/financialsx/desktop
│   ├── wails.json           # Wails configuration
│   ├── build/               # Build configuration and assets
│   └── frontend/            # React + Vite frontend
│       ├── src/
│       │   ├── main.js      # Frontend entry point
│       │   ├── app.css      # Application styles
│       │   └── assets/      # Images and fonts
│       └── wailsjs/         # Generated Wails bindings
│           ├── go/          # Go struct bindings
│           └── runtime/     # Wails runtime API
└── go.mod                   # Root module: github.com/pivoten/financialsx
```

## Architecture Notes

- **Wails**: Desktop framework providing Go backend with React frontend
- **DBF Integration**: Will read legacy Visual FoxPro DBF files
- **SQLite**: Local database for persisting derived data
- **Authentication**: JWT-based flow planned
- **State Reporting**: First major feature to implement

## Development Workflow

1. The desktop app runs from the `desktop/` directory
2. Frontend changes auto-reload in dev mode
3. Backend changes require restart of `wails dev`
4. Generated bindings in `wailsjs/` should not be edited manually

## Next Steps

- Add sample Go binding (Greeter) and call from React to verify integration
- Integrate DBF ingestion logic and persist into SQLite
- Implement authentication JWT flow and reflect status in UI
- Define and implement first feature (State Reporting) behind a feature flag