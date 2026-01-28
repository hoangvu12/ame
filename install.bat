@echo off
setlocal EnableDelayedExpansion

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

:: Symbols (ASCII only)
set "CHECK=%GREEN%OK%R%"
set "CROSS=%RED%FAIL%R%"
set "ARROW=%CYAN%>%R%"

:: Paths
set "AME_DIR=%LOCALAPPDATA%\ame"
set "TOOLS_DIR=%AME_DIR%\tools"
set "PENGU_DIR=%AME_DIR%\pengu"
set "TOOLS_URL=https://raw.githubusercontent.com/Alban1911/Rose/main/injection/tools"
set "PENGU_URL=https://github.com/PenguLoader/PenguLoader/releases/download/v1.1.6/pengu-loader-v1.1.6.zip"
set "PENGU_ZIP=%AME_DIR%\pengu.zip"

:: Header
echo.
echo   %CYAN%======================================%R%
echo   %CYAN%  %BOLD%ame installer%R%
echo   %CYAN%======================================%R%
echo.

set "FAILED=0"

:: Step 1: mod-tools
echo   %BOLD%[1/4]%R% Checking mod-tools...
if exist "%TOOLS_DIR%\mod-tools.exe" (
    echo         [%CHECK%] Already installed
) else (
    echo         %ARROW% Installing to: %DIM%%TOOLS_DIR%%R%
    if not exist "%TOOLS_DIR%" mkdir "%TOOLS_DIR%"
    for %%f in (
        mod-tools.exe
        cslol-diag.exe
        cslol-dll.dll
        wad-extract.exe
        wad-make.exe
    ) do (
        <nul set /p "=         %DIM%Downloading %%f...%R%"
        curl -sL "%TOOLS_URL%/%%f" -o "%TOOLS_DIR%\%%f"
        if errorlevel 1 (
            echo  [%CROSS%]
            set "FAILED=1"
        ) else (
            echo  [%CHECK%]
        )
    )
)
echo.

:: Step 2: Pengu Loader
echo   %BOLD%[2/4]%R% Checking Pengu Loader...
if exist "%PENGU_DIR%\Pengu Loader.exe" (
    echo         [%CHECK%] Already installed
) else (
    echo         %ARROW% Installing to: %DIM%%PENGU_DIR%%R%
    if not exist "%PENGU_DIR%" mkdir "%PENGU_DIR%"
    <nul set /p "=         %DIM%Downloading Pengu Loader v1.1.6...%R%"
    curl -sL "%PENGU_URL%" -o "%PENGU_ZIP%"
    if errorlevel 1 (
        echo  [%CROSS%]
        set "FAILED=1"
    ) else (
        echo  [%CHECK%]
        <nul set /p "=         %DIM%Extracting...%R%"
        powershell -NoProfile -Command "Expand-Archive -Force '%PENGU_ZIP%' '%PENGU_DIR%'"
        if errorlevel 1 (
            echo  [%CROSS%]
            set "FAILED=1"
        ) else (
            echo  [%CHECK%]
        )
        del "%PENGU_ZIP%" 2>nul
    )
)
echo.

:: Step 3: ame plugin
echo   %BOLD%[3/4]%R% Installing ame plugin...
set "PLUGIN_DIR=%PENGU_DIR%\plugins\ame"
if exist "%PLUGIN_DIR%" rmdir /s /q "%PLUGIN_DIR%"
mkdir "%PLUGIN_DIR%"
powershell -NoProfile -Command "Copy-Item -Recurse -Force '%~dp0src\*' '%PLUGIN_DIR%\'"
if errorlevel 1 (
    echo         [%CROSS%] Failed to copy plugin
    set "FAILED=1"
) else (
    echo         [%CHECK%] Plugin installed
)
:: Copy ame.bat and server.ps1 to install dir
copy /y "%~dp0ame.bat" "%AME_DIR%\" >nul 2>&1
copy /y "%~dp0server.ps1" "%AME_DIR%\" >nul 2>&1
echo         [%CHECK%] ame.bat copied to %AME_DIR%
echo.

:: Step 4: Activate Pengu
echo   %BOLD%[4/4]%R% Activating Pengu Loader...
reg query "HKLM\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Image File Execution Options\LeagueClientUx.exe" /v Debugger >nul 2>&1
if %errorlevel%==0 (
    echo         [%CHECK%] Already activated
) else (
    echo.
    echo         %YELLOW%Please click "Activate" in Pengu Loader,%R%
    echo         %YELLOW%then close the window to continue.%R%
    echo.
    start "" /wait "%PENGU_DIR%\Pengu Loader.exe"
    echo         [%CHECK%] Pengu Loader closed
)

:: Footer
echo.
echo   %DIM%--------------------------------------%R%
if "%FAILED%"=="1" (
    echo   %RED%Some components failed to install.%R%
    echo   %DIM%Check your internet connection and try again.%R%
) else (
    echo   %GREEN%Installation complete!%R%
    echo.
    echo   %DIM%To run ame, open:%R%
    echo   %CYAN%%AME_DIR%\ame.bat%R%
)
echo.
pause
