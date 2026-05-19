@echo off
setlocal enabledelayedexpansion

:: ── Read current version ──────────────────────────────────────────────────────
set "VERSION_FILE=version.txt"
if not exist "%VERSION_FILE%" (
    echo 0.0.0 > "%VERSION_FILE%"
)
set /p CURRENT_VERSION=<"%VERSION_FILE%"
:: Trim whitespace/newline
for /f "tokens=* delims= " %%a in ("%CURRENT_VERSION%") do set CURRENT_VERSION=%%a

:: ── Parse MAJOR.MINOR.PATCH ───────────────────────────────────────────────────
for /f "tokens=1,2,3 delims=." %%a in ("%CURRENT_VERSION%") do (
    set MAJOR=%%a
    set MINOR=%%b
    set PATCH=%%c
)

:: ── Determine bump type (default: patch) ─────────────────────────────────────
set "BUMP=%~1"
if "%BUMP%"=="" set "BUMP=patch"

if /i "%BUMP%"=="major" (
    set /a MAJOR=MAJOR+1
    set MINOR=0
    set PATCH=0
) else if /i "%BUMP%"=="minor" (
    set /a MINOR=MINOR+1
    set PATCH=0
) else if /i "%BUMP%"=="patch" (
    set /a PATCH=PATCH+1
) else (
    echo [ERROR] Unknown bump type: "%BUMP%". Use: patch, minor, major
    exit /b 1
)

set "NEW_VERSION=%MAJOR%.%MINOR%.%PATCH%"

:: ── Save new version ──────────────────────────────────────────────────────────
echo %NEW_VERSION%> "%VERSION_FILE%"

echo [build] Version: %CURRENT_VERSION% ^-^> %NEW_VERSION%

:: ── Build frontend ────────────────────────────────────────────────────────────
echo [1/3] Building frontend...
cd src\frontend
call npm run build
if errorlevel 1 (
    echo [ERROR] Frontend build failed.
    exit /b 1
)
cd ..\..

:: ── Prepare output directories ────────────────────────────────────────────────
set "OUT_DIR=bin\%NEW_VERSION%"
set "LATEST_DIR=bin\latest"
if not exist "%OUT_DIR%" mkdir "%OUT_DIR%"
if not exist "%LATEST_DIR%" mkdir "%LATEST_DIR%"

:: ── Build Go binary ───────────────────────────────────────────────────────────
echo [2/3] Building Go binary...
cd src\backend
go build -ldflags "-X main.Version=%NEW_VERSION%" -o "..\..\%OUT_DIR%\ai-engine.exe" ./cmd/ai-engine
if errorlevel 1 (
    echo [ERROR] Go build failed.
    exit /b 1
)
cd ..\..

:: ── Copy to latest ────────────────────────────────────────────────────────────
echo [3/3] Updating bin\latest...
copy /y "%OUT_DIR%\ai-engine.exe" "%LATEST_DIR%\ai-engine.exe" >nul

echo.
echo Done.
echo   Version : %NEW_VERSION%
echo   Binary  : %OUT_DIR%\ai-engine.exe
echo   Latest  : %LATEST_DIR%\ai-engine.exe

endlocal
