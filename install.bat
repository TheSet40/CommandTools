@echo off
where py >nul 2>&1 || goto :python
py -3 "%~dp0install.py" %*
exit /b %ERRORLEVEL%
:python
python "%~dp0install.py" %*
exit /b %ERRORLEVEL%
