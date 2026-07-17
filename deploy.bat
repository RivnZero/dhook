@echo off
setlocal enabledelayedexpansion

echo ============================================
echo   dhook - Local Pre-Release Gatekeeper
echo ============================================
echo.

echo [1/4] Running go vet...
go vet ./...
if %errorlevel% neq 0 (
    echo FAILED: go vet encountered errors.
    exit /b 1
)
echo PASSED
echo.

echo [2/4] Running tests...
go test -v ./...
if %errorlevel% neq 0 (
    echo FAILED: Tests did not pass.
    exit /b 1
)
echo PASSED
echo.

if exist dist rmdir /s /q dist

echo [3/4] Cross-compiling for 8 targets...
echo.

set GOOS=windows&set GOARCH=386&call :build
if errorlevel 1 exit /b 1
set GOOS=windows&set GOARCH=amd64&call :build
if errorlevel 1 exit /b 1
set GOOS=windows&set GOARCH=arm64&call :build
if errorlevel 1 exit /b 1
set GOOS=linux&set GOARCH=386&call :build
if errorlevel 1 exit /b 1
set GOOS=linux&set GOARCH=amd64&call :build
if errorlevel 1 exit /b 1
set GOOS=linux&set GOARCH=arm64&call :build
if errorlevel 1 exit /b 1
set GOOS=darwin&set GOARCH=amd64&call :build
if errorlevel 1 exit /b 1
set GOOS=darwin&set GOARCH=arm64&call :build
if errorlevel 1 exit /b 1

echo.
echo All 8 targets compiled successfully.
echo.

echo [4/4] Build artifacts:
echo.
dir /s /b dist
echo.

echo ============================================
echo   BUILD GATE PASSED
echo.
echo   The project is safe to tag and push:
echo.
echo     git tag vX.Y.Z
echo     git push origin vX.Y.Z
echo.
echo   GoReleaser will handle the release.
echo ============================================
exit /b 0

:build
set "EXT="
if "%GOOS%"=="windows" set "EXT=.exe"
echo   [%GOOS%/%GOARCH%]
if not exist "dist\%GOOS%\%GOARCH%" mkdir "dist\%GOOS%\%GOARCH%"
if !errorlevel! neq 0 (
    echo FAILED: Could not create dist\%GOOS%\%GOARCH%.
    exit /b 1
)
go build -o dist\%GOOS%\%GOARCH%\dhook%EXT% .\examples\advanced
if !errorlevel! neq 0 (
    echo FAILED: %GOOS%/%GOARCH% build failed.
    exit /b 1
)
echo   OK
exit /b 0
