# FinancialsX Desktop

## About

FinancialsX Desktop is a comprehensive financial management application for oil & gas operations, built with Wails (Go + React). The application provides banking functionality, DBF file management, audit capabilities, and financial reporting tools.

### Key Features

- **Banking Module**: Bank account management with real-time GL balance integration
- **DBF File Explorer**: Read, view, and edit legacy DBF files with advanced filtering and search
- **Check Batch Audit**: Compare checks.dbf entries against GLMASTER.dbf for discrepancy detection
- **User Management**: Role-based access control (Root, Admin, Read-Only)
- **Data Management**: Complete CRUD operations on company financial data

### Architecture

- **Backend**: Go with Wails framework and DBF file integration
- **Frontend**: React with Vite, TypeScript, and ShadCN UI components
- **Database**: DBF files for legacy data + SQLite for user management
- **DBF Library**: `github.com/Valentin-Kaiser/go-dbase/dbase`

### Recent Updates

- ✅ **GL Balance Integration**: Bank account cards now display actual General Ledger balances from GLMASTER.dbf
- ✅ **Check Audit System**: Comprehensive audit tool for admin/root users to validate check entries
- ✅ **Enhanced Banking UI**: Improved bank account management with real-time data
- ✅ **DBF Explorer**: Advanced DBF file viewer with sorting, filtering, and editing capabilities

You can configure the project by editing `wails.json`. More information about the project settings can be found
here: https://wails.io/docs/reference/project-config

## Live Development

To run in live development mode, run `wails dev` in the project directory. This will run a Vite development
server that will provide very fast hot reload of your frontend changes. If you want to develop in a browser
and have access to your Go methods, there is also a dev server that runs on http://localhost:34115. Connect
to this in your browser, and you can call your Go code from devtools.

## Building

To build a redistributable, production mode package, use `wails build`.

## Documentation

For detailed development guidance, see [CLAUDE.md](./CLAUDE.md) which contains:
- Complete project architecture overview
- Banking section implementation details
- GL balance integration guide
- DBF file handling procedures
- User management and permissions
- Troubleshooting guides
- Common issues and solutions

## Key Files

- `main.go` - Main application and API endpoints
- `internal/company/company.go` - DBF file operations
- `frontend/src/components/BankingSection.jsx` - Banking module
- `frontend/src/components/DBFExplorer.jsx` - DBF file viewer
- `CLAUDE.md` - Comprehensive development documentation
