@echo off
setlocal enabledelayedexpansion

REM Build script for Tucha
REM Produces both Linux and Windows builds
REM
REM Usage:
REM   build.cmd

echo ======================================
echo Building Tucha (Linux + Windows)
echo ======================================
echo.

REM Step 1: Cleanup old builds
echo [1/5] Cleaning old builds...
if exist "cmd\tucha\*.syso" del /q "cmd\tucha\*.syso" 2>nul
if exist "tucha" del /q "tucha" 2>nul
if exist "tucha.exe" del /q "tucha.exe" 2>nul
if exist "winres\*.syso" del /q "winres\*.syso" 2>nul
echo [OK] Cleanup complete
echo.

REM Step 2: Check for go-winres
echo [2/5] Checking for go-winres...
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

REM Step 3: Generate Windows resources (.syso files)
echo [3/5] Generating Windows resources...
if not exist "winres\winres.json" (
    echo [!] winres\winres.json not found
    echo     Skipping resource generation
    echo.
) else (
    !WINRES_CMD! make --in winres\winres.json --out cmd\tucha\rsrc >nul 2>&1
    if %errorlevel% equ 0 (
        echo [OK] Resource files generated in cmd\tucha\
    ) else (
        echo [!] Warning: go-winres failed
        echo     Continuing without embedded resources
    )
)
echo.

REM Step 4: Build Linux executable
echo [4/5] Building Linux executable...
set GOOS=linux
set GOARCH=amd64
go build -ldflags="-s -w" -o tucha ./cmd/tucha
if %errorlevel% neq 0 (
    echo.
    echo [X] Linux build failed!
    exit /b %errorlevel%
)
echo [OK] Linux build complete: tucha
echo.

REM Step 5: Build Windows executable
echo [5/5] Building Windows executable...
set GOOS=windows
set GOARCH=amd64
go build -ldflags="-s -w" -o tucha.exe ./cmd/tucha
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
echo Build Summary
echo ======================================
dir tucha tucha.exe 2>nul | findstr /i "tucha"
echo.
echo [OK] Build complete!
