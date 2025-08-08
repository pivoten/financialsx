# Windows Deployment Package - FinancialsX Desktop

## Build Information
- **Build Date**: August 7, 2025
- **Architecture**: Windows x64 (64-bit)
- **File**: `desktop.exe` (16.2 MB)
- **Requirements**: Windows 10/11 64-bit with WebView2 Runtime

## Package Contents

### 1. Main Application
- `desktop.exe` - The main FinancialsX application

### 2. OLE COM Server (Required for Database Access)
- `dbapi.prg` - FoxPro source code for COM server
- `dbapi.exe` - Must be built from dbapi.prg in Visual FoxPro (see instructions below)

## Installation Steps

### Step 1: Install Prerequisites

1. **Microsoft Edge WebView2 Runtime** (if not already installed)
   - Download from: https://developer.microsoft.com/en-us/microsoft-edge/webview2/
   - Most Windows 10/11 systems already have this

2. **Visual FoxPro Runtime** (for DBF access)
   - The application needs VFP runtime libraries
   - Usually installed with the dbapi.exe COM server

### Step 2: Build and Register the COM Server

**IMPORTANT**: The dbapi.exe COM server must be built and registered on the target Windows machine.

1. **Build dbapi.exe in Visual FoxPro**:
   ```
   - Open Visual FoxPro 9.0
   - Create new project: dbapi
   - Add dbapi.prg to the project
   - Build as: Win32 executable / COM server (exe)
   - Output: dbapi.exe
   ```

2. **Register the COM Server** (Run as Administrator):
   ```cmd
   cd C:\Path\To\DbApi
   dbapi.exe /regserver
   ```

3. **Verify Registration**:
   - Check Registry: `HKEY_CLASSES_ROOT\Pivoten.DbApi`
   - Should see CLSID and registration info

### Step 3: Configure Application

1. **Create Application Directory**:
   ```cmd
   mkdir "C:\Program Files\FinancialsX"
   ```

2. **Copy Files**:
   - Copy `desktop.exe` to the application directory
   - Copy `dbapi.exe` to the application directory (after building)

3. **Set Up Data Directory**:
   - The app expects data in: `../datafiles/` relative to the exe
   - Create: `C:\Program Files\FinancialsX\datafiles\`
   - Copy your company DBF files to appropriate subdirectories

### Step 4: Create Desktop Shortcut

1. Right-click `desktop.exe` → Send to → Desktop (create shortcut)
2. Rename shortcut to "FinancialsX Desktop"
3. Optional: Set custom icon if available

## Running the Application

### First Run
1. Double-click `desktop.exe` or the desktop shortcut
2. The application will create necessary SQLite databases on first run
3. Log in with your credentials

### Troubleshooting

**Application won't start:**
- Ensure WebView2 Runtime is installed
- Check Windows Event Viewer for errors
- Run as Administrator if permission issues

**"OLE server connection failed" error:**
- Verify dbapi.exe is registered (`dbapi.exe /regserver` as admin)
- Check that Pivoten.DbApi appears in registry
- Ensure Visual FoxPro runtime libraries are installed

**"No bank accounts found" or database errors:**
- Verify datafiles directory structure
- Check that DBF files are in correct locations
- Ensure appdata.dbc exists in company directories

**Blank window or UI not loading:**
- Update WebView2 Runtime to latest version
- Check that JavaScript is not blocked by security software
- Try running with Windows Defender temporarily disabled

## Directory Structure

```
C:\Program Files\FinancialsX\
├── desktop.exe           # Main application
├── dbapi.exe            # COM server (built from dbapi.prg)
├── datafiles\           # Company data
│   ├── Company1\
│   │   ├── appdata.dbc  # Database container
│   │   ├── COA.dbf      # Chart of accounts
│   │   ├── CHECKS.dbf   # Check register
│   │   ├── GLMASTER.dbf # General ledger
│   │   └── ...          # Other DBF files
│   └── Company2\
│       └── ...
└── logs\                # Application logs (created automatically)
```

## Security Notes

1. **Run as Standard User**: The application should run with standard user privileges
2. **COM Server Registration**: Only the initial registration requires admin rights
3. **Data Access**: Users need read/write permissions to the datafiles directory
4. **Firewall**: The application is fully offline, no internet access required

## Updates

To update the application:
1. Close the running application
2. Replace `desktop.exe` with the new version
3. If dbapi.prg was updated, rebuild and re-register dbapi.exe
4. Start the application normally

## Support Files Included

- `WINDOWS_DEPLOYMENT.md` - This file
- `build_dbapi.txt` - Detailed instructions for building dbapi.exe
- `CLAUDE.md` - Technical documentation

## System Requirements

### Minimum:
- Windows 10 version 1809 (64-bit)
- 4 GB RAM
- 100 MB disk space (plus data)
- 1280x720 display

### Recommended:
- Windows 11 (64-bit)
- 8 GB RAM
- 500 MB disk space
- 1920x1080 display

## Known Issues

1. **32-bit Windows**: Not supported due to library constraints
2. **Windows 7/8**: Not supported (lacks WebView2 support)
3. **Network Drives**: DBF files on network drives may have performance issues

## Version Information

- Application Version: 1.0.0
- COM Server Version: 1.0.1
- Build Date: August 7, 2025
- Architecture: x64
- Framework: Wails 2.10.2
- Runtime: Go 1.21+ with WebView2