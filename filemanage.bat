@echo off
setlocal

set "currentDir=%CD%"

mkdir "%currentDir%\%1"
(
echo A new project: %1.cpp
echo write a short description of your project
echo for information on all commands go to dev/cmd
) > "%currentDir%\%1\Readme.md"
mkdir "%currentDir%\%1\src"
mkdir "%currentDir%\%1\src\assets"
echo nul > "%currentDir%\%1\src\assets\empty.png"
(
echo #include ^<iostream^>
echo.
echo int a = 0;
echo.
echo int main^(^) {
echo std::cout ^<^< a ^<^< std::endl^;
echo ^}
) > "%currentDir%\%1\src\main.cpp"
