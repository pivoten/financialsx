# FinancialsX Desktop Icons

This directory contains the application icons for the FinancialsX desktop application.

## Files:
- `appicon.png` - macOS application icon (1024x1024 PNG)
- `icon.ico` - Windows application icon (multi-resolution ICO)

## Usage:
These icons are automatically copied to the build directory during the build process:
- `appicon.png` → `build/appicon.png`
- `icon.ico` → `build/windows/icon.ico`

## To update icons:
1. Replace the files in this directory
2. Copy them to the build directory:
   ```bash
   cp assets/icons/appicon.png build/
   cp assets/icons/icon.ico build/windows/
   ```
3. Rebuild the application:
   ```bash
   wails build
   ```

## Icon Requirements:
- **macOS**: 1024x1024 PNG with transparency
- **Windows**: Multi-resolution ICO (16x16, 32x32, 48x48, 64x64, 128x128, 256x256)