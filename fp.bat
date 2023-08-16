@echo off
setlocal enabledelayedexpansion

set "Dir=%CD%\lib\main.dart"

echo import "package:flutter/material.dart";> %Dir%
echo. >> %Dir%
echo import "extra.dart";>> %Dir%
echo import "settings/settingsPage.dart";>> %Dir%
echo.>> %Dir%
echo void main() {>> %Dir%
echo   runApp(const App());>> %Dir%
echo }>> %Dir%
echo. >> %Dir%
echo class App extends StatelessWidget {>> %Dir%
echo   const App({super.key});>> %Dir%
echo.>> %Dir%
echo   @override>> %Dir%
echo   Widget build(BuildContext context) {>> %Dir%
echo     GlobalSize(context);>> %Dir%
echo     return MaterialApp(>> %Dir%
echo       debugShowCheckedModeBanner: false,>> %Dir%
echo       theme: ThemeData(>> %Dir%
echo         colorScheme: ColorScheme.fromSeed(seedColor: Colors.deepPurple,secondary: Colors.white70),>> %Dir%
echo         useMaterial3: true,>> %Dir%
echo       ),>> %Dir%
echo       home: const HomePage(),>> %Dir%
echo       routes: {>> %Dir%
echo.>> %Dir%
echo       },>> %Dir%
echo     );>> %Dir%
echo   }>> %Dir%
echo }>> %Dir%
echo.>> %Dir%
echo class HomePage extends StatefulWidget {>> %Dir%
echo   const HomePage({super.key});>> %Dir%
echo.>> %Dir%
echo   @override>> %Dir%
echo   State^<HomePage^> createState^(^) ^=^> _HomePageState^(^);>> %Dir%
echo }>> %Dir%
echo.>> %Dir%
echo class _HomePageState extends State^<HomePage^> {>> %Dir%
echo.>> %Dir%
echo   @override>> %Dir%
echo   Widget build(BuildContext context) {>> %Dir%
echo     return Scaffold(>> %Dir%
echo       appBar: AppBar(>> %Dir%
echo         backgroundColor: Theme.of(context).colorScheme.inversePrimary,>> %Dir%
echo         title: Text("App"),>> %Dir%
echo       ),>> %Dir%
echo       body: Center(>> %Dir%
echo         child: Column(>> %Dir%
echo           mainAxisAlignment: MainAxisAlignment.center,>> %Dir%
echo           children: ^<Widget^>[>> %Dir%
echo.>> %Dir%
echo           ],>> %Dir%
echo         ),>> %Dir%
echo       ),>> %Dir%
echo     );>> %Dir%
echo   }>> %Dir%
echo }>> %Dir%

set "exDir=%CD%\lib\extra.dart"

echo import "package:flutter/widgets.dart";> %exDir%
echo.>> %exDir%
echo class GlobalSize {>> %exDir%
echo   static final GlobalSize _instance = GlobalSize._internal();>> %exDir%
echo   static late final double height, width;>> %exDir%
echo.>> %exDir%
echo   factory GlobalSize(BuildContext context) {>> %exDir%
echo     height = MediaQuery.of(context).size.height / 100;>> %exDir%
echo     width = MediaQuery.of(context).size.width / 100;>> %exDir%
echo     return _instance;>> %exDir%
echo   }>> %exDir%
echo   GlobalSize._internal();>> %exDir%
echo }>> %exDir%
echo.>> %exDir%
echo extension DoublePercentExtension on double {>> %exDir%
echo   double get hp =^> this * GlobalSize.height;>> %exDir%
echo   double get wp =^> this * GlobalSize.width;>> %exDir%
echo }>> %exDir%
echo.>> %exDir%
echo extension IntPercentExtension on int {>> %exDir%
echo   int get hp =^> ^(this * GlobalSize.height^).toInt^(^);>> %exDir%
echo   int get wp =^> ^(this * GlobalSize.width^).toInt^(^);>> %exDir%
echo }>> %exDir%

cd lib

mkdir settings

cd settings

st settingsPage

cd \..\..

flutter pub get




