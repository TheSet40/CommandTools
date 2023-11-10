@echo off

if "%1"=="manual" (
  cd android && ./gradlew assembleRelease
  exit /b
)

if "%1"=="ios" (
  eas build -p ios
) else (
  eas build -p android
)
