@echo off 

echo .
git pull
timeout /t 2 /nobreak >nul

echo .
git add .
timeout /t 2 /nobreak >nul

echo .
git checkout -b %1
timeout /t 2 /nobreak >nul

echo .
git commit -m "%2"
timeout /t 2 /nobreak >nul

echo .
git push --set-upstream origin %1
timeout /t 2 /nobreak >nul

echo .
git checkout main
timeout /t 2 /nobreak >nul

echo .
git merge %1
timeout /t 2 /nobreak >nul

exit /b 0
