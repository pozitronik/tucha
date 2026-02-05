@echo off
setlocal enabledelayedexpansion

REM Build script for Tucha
REM Produces both Linux and Windows builds
REM
REM Usage:
REM   build.cmd [version]
REM
REM If version is not provided, it will be extracted from git tags or default to 0.0.0

echo ======================================
echo Building Tucha (Linux + Windows)
echo ======================================
echo.

REM Step 1: Determine version
echo [1/6] Determining version...
set VERSION=
set VERSION_FULL=

if not "%~1"=="" (
    set VERSION=%~1
    echo [OK] Using provided version: !VERSION!
) else (
    for /f "tokens=* USEBACKQ" %%g in (`git describe --tags --abbrev^=0 2^>nul`) do (
        set VERSION=%%g
    )
    if defined VERSION (
        REM Remove 'v' prefix if present
        if "!VERSION:~0,1!"=="v" set VERSION=!VERSION:~1!
        echo [OK] Version from git tag: !VERSION!
    ) else (
        set VERSION=0.0.0
        echo [!] No git tag found, using default: !VERSION!
    )
)

REM Convert to 4-component format for Windows resources
REM Count dots by removing them and comparing length
set "TEMP_VER=!VERSION!"
set "TEMP_VER_NO_DOTS=!TEMP_VER:.=!"
call :strlen ORIG_LEN TEMP_VER
call :strlen NO_DOT_LEN TEMP_VER_NO_DOTS
set /a DOT_COUNT=ORIG_LEN-NO_DOT_LEN

if !DOT_COUNT! equ 2 (
    set VERSION_FULL=!VERSION!.0
) else if !DOT_COUNT! equ 3 (
    set VERSION_FULL=!VERSION!
) else (
    set VERSION_FULL=!VERSION!.0.0.0
)
echo     Version (4-component): !VERSION_FULL!
echo.
goto after_strlen

:strlen <resultVar> <stringVar>
setlocal enabledelayedexpansion
set "s=!%~2!#"
set len=0
for %%P in (4096 2048 1024 512 256 128 64 32 16 8 4 2 1) do (
    if "!s:~%%P,1!" neq "" (
        set /a "len+=%%P"
        set "s=!s:~%%P!"
    )
)
endlocal & set "%~1=%len%"
exit /b

:after_strlen

REM Step 2: Cleanup old builds
echo [2/6] Cleaning old builds...
if exist "cmd\tucha\*.syso" del /q "cmd\tucha\*.syso" 2>nul
if exist "tucha" del /q "tucha" 2>nul
if exist "tucha.exe" del /q "tucha.exe" 2>nul
if exist "winres\*.syso" del /q "winres\*.syso" 2>nul
echo [OK] Cleanup complete
echo.

REM Step 3: Check for go-winres
echo [3/6] Checking for go-winres...
set WINRES_CMD=
set WINRES_FOUND=0

where go-winres.exe >nul 2>&1
if %errorlevel% equ 0 (
    set WINRES_CMD=go-winres.exe
    set WINRES_FOUND=1
    echo [OK] go-winres found in PATH
) else (
    if exist "%USERPROFILE%\go\bin\go-winres.exe" (
        set WINRES_CMD=%USERPROFILE%\go\bin\go-winres.exe
        set WINRES_FOUND=1
        echo [OK] go-winres found in %%USERPROFILE%%\go\bin
    ) else (
        echo [!] go-winres not found
        echo.
        echo Installing go-winres...
        go install github.com/tc-hib/go-winres@latest
        if %errorlevel% equ 0 (
            set WINRES_CMD=%USERPROFILE%\go\bin\go-winres.exe
            set WINRES_FOUND=1
            echo [OK] go-winres installed successfully
        ) else (
            echo [X] Failed to install go-winres
            echo.
            echo Please install manually:
            echo   go install github.com/tc-hib/go-winres@latest
            echo.
            exit /b 1
        )
    )
)
echo.

REM Step 4: Update winres.json with version and generate resources
echo [4/6] Generating Windows resources...
if not exist "winres\winres.json" (
    echo [!] winres\winres.json not found
    echo     Skipping resource generation
    echo.
) else (
    REM Backup original winres.json
    copy /y "winres\winres.json" "winres\winres.json.bak" >nul

    REM Update version using PowerShell
    powershell -Command "$json = Get-Content 'winres\winres.json' -Raw | ConvertFrom-Json; $json.RT_MANIFEST.'#1'.'0409'.identity.version = '!VERSION_FULL!'; $json.RT_VERSION.'#1'.'0000'.fixed.file_version = '!VERSION_FULL!'; $json.RT_VERSION.'#1'.'0000'.fixed.product_version = '!VERSION_FULL!'; $json.RT_VERSION.'#1'.'0000'.info.'0409'.FileVersion = '!VERSION!'; $json.RT_VERSION.'#1'.'0000'.info.'0409'.ProductVersion = '!VERSION!'; $json | ConvertTo-Json -Depth 10 | Set-Content 'winres\winres.json'" 2>nul
    if %errorlevel% equ 0 (
        echo [OK] Updated winres.json with version !VERSION!
    ) else (
        echo [!] Failed to update version in winres.json
    )

    REM Generate resources
    !WINRES_CMD! make --in winres\winres.json --out cmd\tucha\rsrc >nul 2>&1
    if %errorlevel% equ 0 (
        echo [OK] Resource files generated in cmd\tucha\
    ) else (
        echo [!] Warning: go-winres failed
        echo     Continuing without embedded resources
    )

    REM Restore original winres.json
    if exist "winres\winres.json.bak" (
        move /y "winres\winres.json.bak" "winres\winres.json" >nul
    )
)
echo.

REM Step 5: Build Linux executable
echo [5/6] Building Linux executable...
set GOOS=linux
set GOARCH=amd64
go build -ldflags="-s -w -X 'main.version=!VERSION!'" -o tucha ./cmd/tucha
if %errorlevel% neq 0 (
    echo.
    echo [X] Linux build failed!
    exit /b %errorlevel%
)
echo [OK] Linux build complete: tucha
echo.

REM Step 6: Build Windows executable
echo [6/6] Building Windows executable...
set GOOS=windows
set GOARCH=amd64
go build -ldflags="-s -w -X 'main.version=!VERSION!'" -o tucha.exe ./cmd/tucha
if %errorlevel% neq 0 (
    echo.
    echo [X] Windows build failed!
    exit /b %errorlevel%
)
echo [OK] Windows build complete: tucha.exe
echo.

REM Cleanup intermediate files
if exist "cmd\tucha\*.syso" del /q "cmd\tucha\*.syso" 2>nul
if exist "winres\*.syso" del /q "winres\*.syso" 2>nul

REM Summary
echo ======================================
echo Build Summary (v!VERSION!)
echo ======================================
dir tucha tucha.exe 2>nul | findstr /i "tucha"
echo.
echo [OK] Build complete!
