@echo off
setlocal enabledelayedexpansion

set "SCRIPT_DIR=%~dp0"
for %%I in ("%SCRIPT_DIR%..") do set "PROJECT_DIR=%%~fI"
set "TEMP_DIR=%PROJECT_DIR%\temp"
set "PORT=6379"

where wsl >nul 2>nul
if errorlevel 1 (
  echo WSL is required to run Valkey on Windows.
  echo Install WSL and a Linux distribution, then try again.
  exit /b 1
)

if not exist "%TEMP_DIR%" (
  mkdir "%TEMP_DIR%"
)

wsl bash -lc "cd \"$(wslpath '%PROJECT_DIR%')\" && VALKEY_PORT=%PORT% bash ./scripts/run_valkey_linux.sh"
set "EXIT_CODE=%ERRORLEVEL%"

if "%EXIT_CODE%"=="0" (
  echo Valkey is running on port %PORT%.
)

exit /b %EXIT_CODE%
