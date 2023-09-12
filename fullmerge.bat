@echo off 
if "%1"=="-s" (
  git status

  timeout /t 1 /nobreak >nul
  echo.
  echo use fullmerge with: branchname and then commit message
  echo.
  exit /b
)

echo.
git pull
timeout /t 2 /nobreak >nul

echo.
git add .
timeout /t 1 /nobreak >nul

echo.
git checkout -b %1
timeout /t 2 /nobreak >nul

echo.
git commit -m "%2"
timeout /t 2 /nobreak >nul

echo.
git push --set-upstream origin %1
timeout /t 2 /nobreak >nul

echo.
git checkout main
timeout /t 2 /nobreak >nul

echo.
git merge %1
timeout /t 2 /nobreak >nul

git branch -d %1

exit /b 0
