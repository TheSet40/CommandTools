@echo off 
setlocal enabledelayedexpansion

if "%1"=="-s" (
  git status

  timeout /t 1 /nobreak >nul
  echo.
  echo use fullmerge with: branchname and then commit message
  echo.
  exit /b
)

set "COMBINED_PARAMS="
set "FIRST_PARAM=%1"
shift
:ConcatLoop
if "%1"=="" goto AfterConcat
set "COMBINED_PARAMS=!COMBINED_PARAMS! %1"
shift
goto ConcatLoop
:AfterConcat

echo  "!FIRST_PARAM!"
echo "!COMBINED_PARAMS!"