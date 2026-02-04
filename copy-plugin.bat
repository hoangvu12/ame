@echo off
xcopy /Y "src\*.js" "%LOCALAPPDATA%\ame\pengu\plugins\ame\"
xcopy /E /Y "src\locales" "%LOCALAPPDATA%\ame\pengu\plugins\ame\locales\"
