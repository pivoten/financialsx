# financialsx
Pivoten Financials X - Enhanced Legacy

Modern desktop companion app to the legacy Visual FoxPro Accounting Manager. Uses Wails with a Go backend + React frontend, mines DBF data and persists derived data in SQLite, and supports new reporting features like State Reporting.

## Quick start

### Prerequisites
- Go 1.21+ installed and on your PATH. Verify with \`go version\`.
- Node.js (LTS) and npm or pnpm. Verify with \`node -v\` and \`npm -v\`.
- Wails CLI installed: 
```bash
  go install github.com/wailsapp/wails/v2/cmd/wails@latest
```
  Ensure \`$(go env GOPATH)/bin\` is in your PATH (e.g., add \`export PATH=$PATH:$(go env GOPATH)/bin\` to \`~/.zshrc\`).

### Initialize project
1. Set module path, for example:
 ```bash
   go mod init github.com/pivoten/financialsx
 ```
2. Create and switch to bootstrap branch:
 ```bash
   git checkout -b wails-bootstrap
 ```
3. Initialize Wails app:
 ```bash
   wails init -n claude
 ```
   Choose React + Vite frontend when prompted.
4. Install frontend dependencies:
 ```bash
   cd claude/frontend
   npm install
   cd ../..
 ```
5. Run in development mode:
 ```bash
   wails dev
 ```

## Common commands
- Run: \`go run .\` or \`go run main.go\`
- Build binary: \`go build -o financialsx\`
- Test: \`go test ./...\`
- Format: \`go fmt ./...\`
- Lint: \`golangci-lint run\`
- Commit scaffold:
```bash
  git add .
  git commit -m "Bootstrap Wails + React desktop app"
  git push --set-upstream origin wails-bootstrap
```

## Next steps
- Add sample Go binding (Greeter) and call from React to verify integration.
- Integrate DBF ingestion logic and persist into SQLite.
- Implement authentication JWT flow and reflect status in UI.
- Define and implement first feature (State Reporting) behind a feature flag.

## Notes
The authoritative spec lives in \`claude.md\` in the canvas; that document contains architecture, standards, and detailed bootstrap instructions.
EOF
# financialsx
Pivoten Financials X - Enhanced Legacy

Modern desktop companion app to the legacy Visual FoxPro Accounting Manager. Uses Wails with a Go backend + React frontend, mines DBF data and persists derived data in SQLite, and supports new reporting features like State Reporting.

## Quick start

### Prerequisites
- Go 1.21+ installed and on your PATH. Verify with `go version`.
- Node.js (LTS) and npm or pnpm. Verify with `node -v` and `npm -v`.
- Wails CLI installed:
  ```bash
  go install github.com/wailsapp/wails/v2/cmd/wails@latest
  ```
  Ensure `$(go env GOPATH)/bin` is in your PATH (e.g., add `export PATH=$PATH:$(go env GOPATH)/bin` to `~/.zshrc`).

### Initialize project
1. Create and switch to bootstrap branch:
   ```bash
   git checkout -b wails-bootstrap
   ```
2. Set module path for top-level if desired:
   ```bash
   go mod init github.com/pivoten/financialsx
   ```
3. Initialize Wails desktop project:
   ```bash
   wails init -n desktop
   ```
   Choose React + Vite frontend when prompted.
4. Enter the new Wails project and initialize its module:
   ```bash
   cd desktop
   go mod edit -module=github.com/pivoten/financialsx/desktop
   go mod tidy
   ```
5. Install frontend dependencies:
   ```bash
   cd frontend
   npm install
   cd ..
   ```
6. Run in development mode:
   ```bash
   wails dev
   ```

## Common commands
- Run: `go run .` or `go run main.go`
- Build binary: `go build -o financialsx`
- Test: `go test ./...`
- Format: `go fmt ./...`
- Lint: `golangci-lint run`
- Commit scaffold:
  ```bash
  git add .
  git commit -m "Bootstrap Wails desktop app"
  git push --set-upstream origin wails-bootstrap
  ```

## Test github action locally
* `brew install act`
* `act` or if you are on Apple M-series chi `act --container-architecture linux/amd`
* or, you might have to run
* `act -P macos-latest=nektos/act-environments-ubuntu:18.04`
## Next steps
- Add sample Go binding (Greeter) and call from React to verify integration.
- Integrate DBF ingestion logic and persist into SQLite.
- Implement authentication JWT flow and reflect status in UI.
- Define and implement first feature (State Reporting) behind a feature flag.

## Notes
The authoritative spec lives in `desktop.md` (formerly `claude.md`) in the canvas; that document contains architecture, standards, and detailed bootstrap instructions.