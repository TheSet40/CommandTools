@echo off
setlocal EnableDelayedExpansion

if "%1"=="-list" (

    set "SCRIPT_DIR=%~dp0"

    set "FILEPATH=!SCRIPT_DIR!VersionInfo.txt"

    if not exist "!FILEPATH!" (
        echo File not found: "!FILEPATH!"
        exit /b
    )

    for /F "tokens=*" %%A in (!FILEPATH!) do (
        echo %%A
    )

) 

if "%1"=="-commands" (
    set "SCRIPT_DIR=%~dp0"
    echo Available commands:

    for /F "tokens=*" %%f in ('dir "!SCRIPT_DIR!*.bat" /B') do (
        echo %%~nf
    )
)

endlocal
