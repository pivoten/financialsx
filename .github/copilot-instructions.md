# Copilot Instructions for financialsx

Concise, repo-specific guidance for working on this Wails (Go + React) desktop companion to a legacy Visual FoxPro system.

## Big picture
- App root: `desktop/` (Go backend + React/Vite frontend). Entry: `desktop/main.go` (Wails `App` exports to frontend).
- Data sources: DBF files via `internal/company`, optional VFP OLE COM (Windows-only) via `internal/ole`, and SQLite via `internal/database`.
- Generated bindings in `desktop/frontend/wailsjs/` (don’t edit). Deeper docs: `desktop/CLAUDE.md`.

## Dev workflow
- Live dev: `cd desktop && wails dev`
- Build: `cd desktop && wails build`
- Frontend deps: `cd desktop/frontend && npm install`
- Go: `go test ./...`, `go fmt ./...` (Go 1.24 per go.mod)

## Core modules & patterns
- DBF access (`internal/company`): use `ReadDBFFile(...)`. Results use `rows` (not `data`). Handle field variants (e.g., account: `CACCTNO|ACCOUNT|ACCTNO`; amount: `AMOUNT|NAMOUNT|BALANCE`).
- Balance cache (`internal/database`): init with `InitializeBalanceCache(db)`; combine GL totals + outstanding checks for fast reads; expose explicit refresh for rescans.
- Reconciliation (`internal/reconciliation`): SQLite-backed drafts/history (JSON fields). Frontend tracks checks by CIDCHEC. App methods in `main.go` (e.g., `SaveReconciliationDraft`, `CommitReconciliation`).
- Auth (`internal/auth`): gate endpoints with `a.currentUser.HasPermission(...)` (admin-only for users/settings/maintenance).
- Logging (`internal/debug`): prefer `SimpleLog/LogInfo/LogError` over raw prints for traceability.

## Conventions
- Company identifier may be name or full path; prefer passing the path when available (e.g., from `compmast.dbf`).
- Outstanding checks: `LCLEARED = false` AND `LVOID = false`; filter by `CACCTNO` for account views.
- Frontend calls exported `App` methods (via `wailsjs` bindings); never edit generated bindings.
- OLE COM is lazy-initialized; don’t preload; `ole.CloseOLEConnection()` on logout/switch.

## Common flows
- Bank accounts: read COA.dbf → filter `LBANKACCT = true` → `GetAccountBalance(company, acctNo)` (handles field variants) → optionally cache.
- DBF paging: `GetDBFTableDataPaged(company, file, offset, limit, sortCol, sortDir)` → consume `columns` + `rows`.
- Reconciliation: `SaveReconciliationDraft`/`GetReconciliationDraft` during edit; `CommitReconciliation` to finalize (schema in `desktop/CLAUDE.md`).

## Where to add
- New API: add to `App` in `desktop/main.go`; put logic in `internal/*`; return JSON-serializable types.
- DBF readers: extend `internal/company` with consistent field normalization.
- Cached computations: extend `internal/database`; init in `InitializeCompanyDatabase`.

## Gotchas
- macOS/Linux: COM not available; DBF reads still work via `go-dbase`.
- Don’t edit generated `wailsjs` or the embedded assets block in `main.go`.
- Heavy GL scans are slow; prefer cache + explicit refresh actions.

See `desktop/CLAUDE.md` and `desktop/README.md` for full details and field mappings.

## Docs highlights
- DBF dates (`desktop/DBF_DATE_PARSING.md`): Prefer time.Time from go-dbase; if strings, try multiple formats. Outstanding checks cache currently ignores date; when touching `RefreshOutstandingChecks`, add a configurable date cutoff.
- Supabase auth (`desktop/SUPABASE_AUTH_SETUP.md`): Optional cloud auth. Toggle in `frontend/src/config/supabase.config.js` via `useSupabaseAuth`. Local SQLite auth remains the fallback; backend JWT validation is optional.
- Windows/OLE (`desktop/WINDOWS_DEPLOYMENT.md`): OLE COM ProgID `Pivoten.DbApi` built from `desktop/dbapi.prg`, requires WebView2 + VFP runtime and registration (`/regserver`). Keep OLE lazy; use native DBF reads elsewhere.
- DBF fields (`desktop/docs/DBF_FIELDS.md`): Keys—GL: `CACCTNO`, `NDEBITS`, `NCREDITS` (balance = debits - credits). COA: `LBANKACCT`. CHECKS: `LCLEARED`, `LVOID`, `DCHECKDATE`, `NAMOUNT`.
