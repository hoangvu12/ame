@echo off
setlocal

set "AME_DIR=%LOCALAPPDATA%\ame"

echo Uninstalling ame...

if exist "%AME_DIR%" (
    rmdir /s /q "%AME_DIR%"
    echo Done. All ame files removed.
) else (
    echo Nothing to uninstall.
)
pause
