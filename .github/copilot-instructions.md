# Copilot Instructions for financialsx

Use these repo-specific notes to work effectively with this Wails (Go + React) desktop app that modernizes a legacy Visual FoxPro system.

## Big picture
- Desktop app lives in `desktop/` (Go backend + React/Vite frontend). Key entry: `desktop/main.go` (Wails `App` methods are exported to the frontend).
- Data sources:
  - Legacy DBF files (COA.dbf, GLMASTER.dbf, CHECKS.dbf, etc.) read directly via `internal/company` using `github.com/Valentin-Kaiser/go-dbase`.
  - Optional Visual FoxPro COM server for DB access on Windows via `internal/ole` (COM: `Pivoten.DbApi`, defined in `desktop/dbapi.prg`). This is deferred/Windows-only; the default path uses native DBF reads.
  - SQLite (via `internal/database`) for app data, balance caching, users/auth, and reconciliation.
- Frontend bindings are generated in `desktop/frontend/wailsjs/` (don’t edit generated files).
- Deep architecture and feature docs: `desktop/CLAUDE.md`.

## Dev workflows
- Run (live dev): `cd desktop && wails dev` (frontend hot reload; restart dev on backend changes).
- Build app: `cd desktop && wails build`.
- Frontend deps: `cd desktop/frontend && npm install`.
- Tests/Lint: `go test ./...`, `go fmt ./...` (project uses Go 1.24 per `go.mod`).

## Key modules and patterns
- DBF access: `internal/company`
  - Primary function: `ReadDBFFile(companyNameOrPath, fileName, search, offset, limit, sortCol, sortDir)`.
  - Returns map with `rows` (not `data`). When consuming, use `result["rows"]`.
  - Column name variations are handled. Example for account numbers: `CACCTNO`/`ACCOUNT`/`ACCTNO`. For amounts: `AMOUNT`/`NAMOUNT`/`BALANCE`.
- Balance cache: `internal/database`
  - Initialize on company DB open: `database.InitializeBalanceCache(db)`.
  - Cached view of bank balances combines GL totals + outstanding checks. Use fast getters (e.g., “get cached balances” functions) and explicit refreshers for heavy rescans.
- Reconciliation: `internal/reconciliation`
  - SQLite-backed drafts/history with JSON fields. Frontend uses CIDCHEC (unique check IDs) for reliable selection.
  - Public methods are surfaced from `App` in `main.go` (e.g., `SaveReconciliationDraft`, `GetReconciliationDraft`, `CommitReconciliation`).
- Auth/permissions: `internal/auth`
  - Check with `a.currentUser.HasPermission(...)`. Gate admin-only endpoints (users, settings, maintenance).
- Logging: `internal/debug` (`SimpleLog`, `LogInfo`, `LogError`). Prefer these over bare `fmt.Printf` for app-visible traces.

## Important conventions
- Company identifier can be a human name or a full data path; server-side functions commonly pass-through and autodetect. When you have a path (e.g., from `compmast.dbf`), prefer passing the path for precision.
- Outstanding checks are defined as `LCLEARED = false` AND `LVOID = false` (filter additionally by `CACCTNO` for account-specific views).
- Frontend should never import from `wailsjs/` by path you edit; bindings are generated. Call exported `App` methods (example: `GetDBFTableData(company, file)` or `GetDBFTableDataPaged(...)`).
- Avoid preloading the OLE COM connection. The app intentionally initializes OLE lazily and calls `ole.CloseOLEConnection()` on logout/switch.

## Typical flows (examples)
- Load bank accounts (COA.dbf) → filter `LBANKACCT = true` → call `GetAccountBalance(company, accountNumber)` to compute GL total (falls back to column variants) → optionally persist/calc via balance cache.
- Fetch DBF page: `GetDBFTableDataPaged(company, "CHECKS.dbf", offset, limit, sortCol, sortDir)` → consume `rows` and `columns` in frontend table.
- Reconciliation: use `SaveReconciliationDraft`/`GetReconciliationDraft` to persist UI state; commit with `CommitReconciliation` (see schema notes in `desktop/CLAUDE.md`).

## Where to add things
- New backend API: add exported method on `App` in `desktop/main.go`. Keep return types JSON-serializable. Use services in `internal/*` for logic.
- New DBF readers/parsers: extend `internal/company` and keep field-name normalization consistent.
- New cached computations: extend `internal/database` and initialize in `InitializeCompanyDatabase`.
- Frontend lists: follow the reusable DataTable pattern described in `desktop/CLAUDE.md` (`frontend/src/components/ui/data-table.jsx`).

## Gotchas
- macOS/Linux won’t support the Visual FoxPro COM server; DBF reads still work via `go-dbase`.
- Don’t edit generated `wailsjs` bindings or embedded assets block in `main.go`.
- Heavy GL scans are slow; prefer cached reads, and expose explicit “refresh” actions for recompute.

For deeper details and field mappings, consult `desktop/CLAUDE.md` and `desktop/README.md`. If anything here seems off or incomplete, ask for the specific flow and we’ll refine this file.
