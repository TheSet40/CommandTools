@echo off
if "%1"=="" (
  eas build -p android
) else (
  eas build -p ios
)
