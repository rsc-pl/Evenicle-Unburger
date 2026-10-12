@echo off
setlocal
cd /d "%~dp0"
go run .\cmd\evenicle-builder %*
exit /b %ERRORLEVEL%
