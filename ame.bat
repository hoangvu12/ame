@echo off

:: Self-elevate to admin (needed to suspend game process)
net session >nul 2>&1
if errorlevel 1 (
    powershell -NoProfile -Command "Start-Process -Verb RunAs -FilePath '%~f0'"
    exit /b
)

setlocal
set "AME_DIR=%LOCALAPPDATA%\ame"
set "PENGU_DIR=%AME_DIR%\pengu"

:: Start Pengu Loader
if exist "%PENGU_DIR%\Pengu Loader.exe" (
    echo Starting Pengu Loader...
    start "" "%PENGU_DIR%\Pengu Loader.exe"
) else (
    echo Pengu Loader not found. Run install.bat first.
    pause
    exit /b 1
)

:: Start WebSocket server (stays in foreground)
echo Starting ame server...
echo.
powershell -NoProfile -ExecutionPolicy Bypass -File "%~dp0server.ps1"
echo.
echo Cleaning up...
taskkill /F /IM "Pengu Loader.exe" >nul 2>&1
taskkill /F /IM "mod-tools.exe" >nul 2>&1
echo Server exited.
pause
