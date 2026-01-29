@echo off
setlocal EnableDelayedExpansion

:: Keep window open on error
if "%~1"=="" (
    cmd /k "%~f0" run
    exit /b
)


:: Setup ANSI escape codes
for /F %%a in ('"prompt $E$S & echo on & for %%b in (1) do rem"') do set "ESC=%%a"

:: Colors
set "R=%ESC%[0m"
set "GREEN=%ESC%[92m"
set "RED=%ESC%[91m"
set "CYAN=%ESC%[96m"
set "YELLOW=%ESC%[93m"
set "DIM=%ESC%[90m"
set "BOLD=%ESC%[1m"

set "AME_DIR=%LOCALAPPDATA%\ame"
set "PENGU_DIR=%AME_DIR%\pengu"

:: Header
echo.
echo   %CYAN%======================================%R%
echo   %CYAN%  %BOLD%ame uninstaller%R%
echo   %CYAN%======================================%R%
echo.

:: Confirm
echo   %YELLOW%This will remove ame and all its data.%R%
echo.
set /p "CONFIRM=  Are you sure? (y/N): "
if /i not "%CONFIRM%"=="y" (
    echo.
    echo   %DIM%Uninstall cancelled.%R%
    echo.
    pause
    exit /b 0
)

echo.

:: Kill running processes
echo   %DIM%Stopping processes...%R%
taskkill /F /IM "ame.exe" >nul 2>&1
taskkill /F /IM "ame-server.exe" >nul 2>&1
taskkill /F /IM "mod-tools.exe" >nul 2>&1
taskkill /F /IM "Pengu Loader.exe" >nul 2>&1

:: Check if Pengu is installed in ame's directory or externally
:: Query registry for Pengu installation path
set "PENGU_EXTERNAL=0"
set "EXTERNAL_PLUGIN_DIR="

for /f "tokens=2*" %%a in ('reg query "HKLM\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Image File Execution Options\LeagueClientUx.exe" /v Debugger 2^>nul ^| findstr /i "Debugger"') do (
    set "REG_VALUE=%%b"
)

:: Parse the path from rundll32 "path\to\core.dll", #6000
if defined REG_VALUE (
    :: Extract path between quotes
    for /f tokens^=2^ delims^=^" %%p in ("!REG_VALUE!") do (
        set "CORE_PATH=%%p"
    )
    if defined CORE_PATH (
        :: Get parent directory of core.dll
        for %%d in ("!CORE_PATH!") do set "PENGU_INSTALL_DIR=%%~dpd"
        :: Remove trailing backslash
        if "!PENGU_INSTALL_DIR:~-1!"=="\" set "PENGU_INSTALL_DIR=!PENGU_INSTALL_DIR:~0,-1!"
        :: Check if it's NOT in ame's directory
        echo !PENGU_INSTALL_DIR! | findstr /i /c:"%LOCALAPPDATA%\ame\pengu" >nul
        if errorlevel 1 (
            set "PENGU_EXTERNAL=1"
            set "EXTERNAL_PLUGIN_DIR=!PENGU_INSTALL_DIR!\plugins\ame"
        )
    )
)

if "!PENGU_EXTERNAL!"=="1" (
    echo   %CYAN%Pengu Loader is installed externally.%R%
    echo   %DIM%Only removing ame plugin, keeping Pengu Loader intact.%R%
    echo.
    if exist "!EXTERNAL_PLUGIN_DIR!" (
        rmdir /s /q "!EXTERNAL_PLUGIN_DIR!" >nul 2>&1
        if exist "!EXTERNAL_PLUGIN_DIR!" (
            echo   %YELLOW%[--]%R% Could not remove plugin from external Pengu
        ) else (
            echo   %GREEN%[OK]%R% Removed ame plugin from external Pengu
        )
    ) else (
        echo   %DIM%[--]%R% No ame plugin found in external Pengu
    )
) else (
    :: Pengu is in ame's directory or not installed - do full uninstall
    echo   %DIM%Uninstalling Pengu Loader...%R%
    echo   %DIM%A new window will open requesting admin rights.%R%
    echo.

    :: Create temp PowerShell script
    set "PS_SCRIPT=%TEMP%\pengu_uninstall.ps1"
    echo Write-Host "Uninstalling Pengu Loader..." -ForegroundColor Yellow > "!PS_SCRIPT!"
    echo Write-Host "" >> "!PS_SCRIPT!"
    echo irm https://pengu.lol/clean ^| iex >> "!PS_SCRIPT!"
    echo Write-Host "" >> "!PS_SCRIPT!"
    echo Write-Host "Done!" -ForegroundColor Green >> "!PS_SCRIPT!"
    echo Write-Host "" >> "!PS_SCRIPT!"
    echo Write-Host "Press any key to close..." -ForegroundColor Cyan >> "!PS_SCRIPT!"
    echo $null = $Host.UI.RawUI.ReadKey^("NoEcho,IncludeKeyDown"^) >> "!PS_SCRIPT!"

    :: Run with admin rights
    powershell -Command "Start-Process powershell -ArgumentList '-NoProfile -ExecutionPolicy Bypass -File \"!PS_SCRIPT!\"' -Verb RunAs -Wait"
    if !errorlevel!==0 (
        echo   %GREEN%[OK]%R% Pengu Loader uninstall completed
    ) else (
        echo   %YELLOW%[--]%R% Pengu Loader uninstall skipped or cancelled
    )

    :: Cleanup temp script
    del "!PS_SCRIPT!" >nul 2>&1
)

:: Remove ame directory
echo   %DIM%Removing ame files...%R%
if exist "%AME_DIR%" (
    rmdir /s /q "%AME_DIR%" >nul 2>&1
    if exist "%AME_DIR%" (
        echo   %RED%[ERR]%R% Could not fully remove %AME_DIR%
        echo         %DIM%Some files may be in use. Try closing League client first.%R%
    ) else (
        echo   %GREEN%[OK]%R% Removed %AME_DIR%
    )
) else (
    echo   %DIM%[--]%R% ame directory not found
)

:: Footer
echo.
echo   %DIM%--------------------------------------%R%
echo   %GREEN%Uninstall complete!%R%
echo.
echo   %DIM%Note: This uninstaller can be deleted manually.%R%
echo.
pause
exit /b 0
