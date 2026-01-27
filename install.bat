@echo off
setlocal

set "AME_DIR=%LOCALAPPDATA%\ame"
set "TOOLS_DIR=%AME_DIR%\tools"
set "PENGU_DIR=%AME_DIR%\pengu"
set "TOOLS_URL=https://raw.githubusercontent.com/Alban1911/Rose/main/injection/tools"
set "PENGU_URL=https://github.com/PenguLoader/PenguLoader/releases/download/v1.1.6/pengu-loader-v1.1.6.zip"
set "PENGU_ZIP=%AME_DIR%\pengu.zip"

echo === ame installer ===
echo.

:: Download mod-tools
echo [1/2] Installing mod-tools to: %TOOLS_DIR%
if not exist "%TOOLS_DIR%" mkdir "%TOOLS_DIR%"

set "FAILED=0"
for %%f in (
    mod-tools.exe
    cslol-diag.exe
    cslol-dll.dll
    wad-extract.exe
    wad-make.exe
) do (
    echo   Downloading %%f...
    curl -sL "%TOOLS_URL%/%%f" -o "%TOOLS_DIR%\%%f"
    if errorlevel 1 (
        echo   FAILED: %%f
        set "FAILED=1"
    )
)

:: Download Pengu Loader
echo.
echo [2/2] Installing Pengu Loader to: %PENGU_DIR%
if not exist "%PENGU_DIR%" mkdir "%PENGU_DIR%"

echo   Downloading Pengu Loader v1.1.6...
curl -sL "%PENGU_URL%" -o "%PENGU_ZIP%"
if errorlevel 1 (
    echo   FAILED: Pengu Loader download
    set "FAILED=1"
) else (
    echo   Extracting...
    powershell -NoProfile -Command "Expand-Archive -Force '%PENGU_ZIP%' '%PENGU_DIR%'"
    if errorlevel 1 (
        echo   FAILED: Pengu Loader extraction
        set "FAILED=1"
    )
    del "%PENGU_ZIP%" 2>nul
)

:: Copy ame plugin to Pengu plugins folder
echo.
echo Copying ame plugin to Pengu plugins...
set "PLUGIN_DIR=%PENGU_DIR%\plugins\ame"
if not exist "%PLUGIN_DIR%" mkdir "%PLUGIN_DIR%"
powershell -NoProfile -Command "Copy-Item -Force '%~dp0index.js' '%PLUGIN_DIR%\index.js'"
if errorlevel 1 (
    echo   FAILED: could not copy plugin
    set "FAILED=1"
) else (
    echo   Plugin copied to: %PLUGIN_DIR%
)

echo.
if "%FAILED%"=="1" (
    echo Some files failed to download. Check your internet connection.
) else (
    echo Done. All components installed to: %AME_DIR%
)
pause
