@echo off 
setlocal enabledelayedexpansion

if "%1"=="" (
  echo funktionen behöver branch och meddelande!
  exit /b
)

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


echo.
git pull
timeout /t 2 /nobreak >nul

echo.
git add .
timeout /t 1 /nobreak >nul

echo.
git checkout -b !FIRST_PARAM!
timeout /t 2 /nobreak >nul

echo.
git commit -m "!COMBINED_PARAMS!"
timeout /t 2 /nobreak >nul

echo.
git push --set-upstream origin !FIRST_PARAM!
timeout /t 2 /nobreak >nul

echo.
git checkout main
timeout /t 2 /nobreak >nul

echo.
git merge !FIRST_PARAM!
timeout /t 2 /nobreak >nul

exit /b 0
