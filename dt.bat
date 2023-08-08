@echo off
setlocal enabledelayedexpansion

set "input=%~1"
set "pagepath=!input:lib\=!"
for %%A in (!input!) do set "input=%%~nxA"
set "first_char=!input:~0,1!"
set "restofString=!input:~1!"
for %%A in (A B C D E F G H I J K L M N O P Q R S T U V W X Y Z) do (
    set "first_char=!first_char:%%A=%%A!"
)
set "capitalName=!first_char!!restofString!"

set "Dir=%CD%\%pagepath%.dart"

echo %Dir% | find /I "\lib" > nul
if errorlevel 1 (
    set "Dir=%CD%\lib\%pagepath%.dart"
)

if not exist "!Dir!\.." (
    echo not a flutter project.
    exit /b
)

echo import "package:flutter/material.dart";> %Dir%
echo.>> %Dir%
echo class !capitalName!Tab extends StatefulWidget {>> %Dir%
echo   const !capitalName!Tab({super.key});>> %Dir%
echo.>> %Dir%
echo   @override>> %Dir%
echo   State^<!capitalName!Tab^> createState^(^) ^=^> _!capitalName!TabState^(^);>> %Dir%
echo }>> %Dir%
echo.>> %Dir%
echo.>> %Dir%
echo class _!capitalName!TabState extends State^<!capitalName!Tab^> {>> %Dir%
echo   @override>> %Dir%
echo   Widget build(BuildContext context) {>> %Dir%
echo     return WillPopScope(>> %Dir%
echo       onWillPop: () async {>> %Dir%
echo         return true;>> %Dir%
echo       },>> %Dir%
echo       child: Scaffold(>> %Dir%
echo         extendBody: true,>> %Dir%
echo         appBar: AppBar(>> %Dir%
echo           centerTitle: true,>> %Dir%
echo           automaticallyImplyLeading: false,>> %Dir%
echo           leading: null,>> %Dir%
echo           actions: null,>> %Dir%
echo           titleSpacing: 0,>> %Dir%
echo           elevation: 0,>> %Dir%
echo           backgroundColor: const Color(0xFFFFFFFF),>> %Dir%
echo           title: const Stack(>> %Dir%
echo             alignment: Alignment.bottomLeft,>> %Dir%
echo             children: [>> %Dir%
echo             ],>> %Dir%
echo           ),>> %Dir%
echo         ),>> %Dir%
echo         body: Expanded(>> %Dir%
echo           child: Padding(>> %Dir%
echo            padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 5),>> %Dir%
echo            child: Container(>> %Dir%
echo              color: const Color(0xFFFFFFFF),>> %Dir%
echo              child: Column(>> %Dir%
echo                children: [>> %Dir%
echo.>> %Dir%
echo                ],>> %Dir%
echo              ),>> %Dir%
echo            ),>> %Dir%
echo          ),>> %Dir%
echo         ),>> %Dir%
echo       ),>> %Dir%
echo     );>> %Dir%
echo   }>> %Dir%
echo }>> %Dir%

for %%A in ("%Dir%\..") do set "parent_dir=%%~dpnA"

:search_gitignore
if "%parent_dir%"=="%Dir%" (
    echo no github found.
) else (
    if exist "%parent_dir%\.gitignore" (
       echo found git ignore, adding file to github.
       for %%A in (!Dir!) do set "Dir=%%~nxA"
       git add "%Dir%
    ) else (
        set "parent_dir=%parent_dir%\.."
        goto search_gitignore
    )
)

