@echo off 

echo Running git pull
git pull
timeout /t 2 /nobreak >nul

echo Running git add .
git add .
timeout /t 2 /nobreak >nul

echo Running git checkout -b %1
git checkout -b %1
timeout /t 2 /nobreak >nul

echo Running git commit -m "%2"
git commit -m "%2"
timeout /t 2 /nobreak >nul

echo Running git push --set-upstream origin %1
git push --set-upstream origin %1
timeout /t 2 /nobreak >nul

echo Running git checkout main
git checkout main
timeout /t 2 /nobreak >nul

echo Running git merge %1
git merge %1
timeout /t 2 /nobreak >nul

exit /b 0
