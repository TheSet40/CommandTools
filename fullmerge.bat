@echo off 
setlocal enabledelayedexpansion

if "%1"=="" (
  echo funktionen behöver ny branch, merge branch och meddelande!
  exit /b
)

if "%1"=="-s" (
  git status

  timeout /t 1 /nobreak >nul
  echo.
  echo use fullmerge with: new branchname, target branch and then commit message
  echo.
  exit /b
)

set "COMBINED_PARAMS="
set "First_PARAM=%1"
set "Second_PARAM=%2"

REM
shift
shift

:ConcatLoop
if "%1"=="" goto AfterConcat
set "COMBINED_PARAMS=!COMBINED_PARAMS! %1"
shift
goto ConcatLoop
:AfterConcat

echo "!First_PARAM! -> !Second_PARAM!"
echo "!COMBINED_PARAMS!"

echo.
git pull
timeout /t 2 /nobreak > nul

echo.
git add .
timeout /t 1 /nobreak > nul

echo.
git checkout -b !First_PARAM!
timeout /t 1 /nobreak > nul

echo.
git commit -m "!COMBINED_PARAMS!"
timeout /t 2 /nobreak > nul

echo.
git push --set-upstream origin !First_PARAM!
timeout /t 2 /nobreak > nul

echo.
git checkout !Second_PARAM!
timeout /t 1 /nobreak > nul

echo.
git merge !First_PARAM!

exit /b 0
