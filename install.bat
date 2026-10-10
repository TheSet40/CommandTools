@echo off
pushd "%~dp0"
go run ./cmd/install %*
set code=%ERRORLEVEL%
popd
exit /b %code%
