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

if "%2"=="-s" (
  echo import "package:flutter/material.dart";> %Dir%
  echo.>> %Dir%
  echo import "../main.dart";>> %Dir%
  echo.>> %Dir%
  echo class SettingsTab extends StatefulWidget {>> %Dir%
  echo   const SettingsTab({super.key});>> %Dir%
  echo.>> %Dir%
  echo   @override>> %Dir%
  echo   State<SettingsTab> createState^(^) => _SettingsState^(^);>> %Dir%
  echo }>> %Dir%
  echo.>> %Dir%
  echo class Settings>> %Dir%
  echo {>> %Dir%
  echo   static String appVersion = "0.1.0", commit = "0.1.0";>> %Dir%
  echo.  >> %Dir%
  echo   static bool developer = false;>> %Dir%
  echo.>> %Dir%
  echo   static String listToIndexString(List<dynamic> list, String sign) {>> %Dir%
  echo     return list.asMap().entries.map((entry) {>> %Dir%
  echo       dynamic index = entry.key + 1;>> %Dir%
  echo       dynamic value = entry.value;>> %Dir%
  echo       return "$index: $value $sign ${index % 2 != 0 ? (value.toString().length == 4 ? "    ": "  ") :"\n"}";>> %Dir%
  echo     }).join();>> %Dir%
  echo   }>> %Dir%
  echo }>> %Dir%
  echo.>> %Dir%
  echo class _SettingsState extends State<SettingsTab> {>> %Dir%
  echo.>> %Dir%
  echo   @override>> %Dir%
  echo   Widget build(BuildContext context) {>> %Dir%
  echo     return WillPopScope(>> %Dir%
  echo       onWillPop: () async {>> %Dir%
  echo         return true;>> %Dir%
  echo       },>> %Dir%
  echo       child: Scaffold(>> %Dir%
  echo         appBar: AppBar(>> %Dir%
  echo           title: Text(""),>> %Dir%
  echo           centerTitle: true,>> %Dir%
  echo           leading: IconButton(>> %Dir%
  echo             icon: const Icon(Icons.arrow_back),>> %Dir%
  echo             onPressed: () async>> %Dir%
  echo             {>> %Dir%
  echo             },>> %Dir%
  echo           ),>> %Dir%
  echo         ),>> %Dir%
  echo         body: Padding(>> %Dir%
  echo           padding: const EdgeInsets.symmetric(horizontal: 25),>> %Dir%
  echo           child: Column(>> %Dir%
  echo             mainAxisAlignment: MainAxisAlignment.center,>> %Dir%
  echo             children: [>> %Dir%
  echo               const SizedBox(height: 35),>> %Dir%
  echo               Column(>> %Dir%
  echo                 children: [>> %Dir%
  echo                   ListTile(>> %Dir%
  echo                     title: Text("", style: TextStyle(fontSize: 18, color: const Color(0xFFf5f5f5))),>> %Dir%
  echo                     leading: const Icon(Icons.settings),>> %Dir%
  echo                     trailing: const Icon(Icons.arrow_forward_ios_rounded),>> %Dir%
  echo                     iconColor: const Color(0xFFFFFFFF),>> %Dir%
  echo                     onTap: () {>> %Dir%
  echo                     },>> %Dir%
  echo                   ),>> %Dir%
  echo                   const Divider(color:  Color(0xFF3c3d3f) , thickness: 2),>> %Dir%
  echo                 ],>> %Dir%
  echo               ),>> %Dir%
  echo               const Expanded(>> %Dir%
  echo                 child: Align(>> %Dir%
  echo                   alignment: Alignment.bottomCenter,>> %Dir%
  echo                   child: Column(>> %Dir%
  echo                     mainAxisSize: MainAxisSize.min,>> %Dir%
  echo                     children: [>> %Dir%
  echo                       SizedBox(height: 8.0),>> %Dir%
  echo                       Text("från", style: TextStyle(fontSize: 13, color: Colors.black)),>> %Dir%
  echo                       Text("...",textAlign: TextAlign.center, style: TextStyle(fontSize: 20, color:Colors.black)),>> %Dir%
  echo                     ],>> %Dir%
  echo                   ),>> %Dir%
  echo                 ),>> %Dir%
  echo               ),>> %Dir%
  echo               const SizedBox(height: 20.0)>> %Dir%
  echo             ],>> %Dir%
  echo           ),>> %Dir%
  echo         ),>> %Dir%
  echo       ),>> %Dir%
  echo     );>> %Dir%
  echo   }>> %Dir%
  echo }>> %Dir%
) else (
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
)

for %%A in ("%Dir%\..") do set "parent_dir=%%~dpnA"

:search_gitignore
if "%parent_dir%"=="%Dir%" (
    echo no github repo found.
) else (
    if exist "%parent_dir%\.gitignore" (
       echo found git ignore,trying to add file to github.
       for %%A in (!Dir!) do set "Dir=%%~nxA"
       git add "%Dir%"
    ) else (
        set "parent_dir=%parent_dir%\.."
        goto search_gitignore
    )
)

