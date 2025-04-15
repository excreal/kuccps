@echo off
SETLOCAL

:: Configuration
set "APP_NAME=kuccps"
set "GO_ARCH=amd64"             :: Can be amd64, 386, arm, arm64
set "GO_OS=windows"             :: Target OS
set "OUTPUT_DIR=bin"            :: Output directory
set "ICON_FILE=app.ico"         :: Optional icon file
set "LD_FLAGS=-s -w"            :: Linker flags to reduce binary size
set "BUILD_FLAGS=-trimpath"     :: Additional build flags

:: Create output directory if it doesn't exist
if not exist "%OUTPUT_DIR%" mkdir "%OUTPUT_DIR%"

:: Check for Go installation
where go >nul 2>&1
if %ERRORLEVEL% neq 0 (
    echo Error: Go compiler not found in PATH
    pause
    exit /b 1
)

:: Build with optional icon
if exist "%ICON_FILE%" (
    echo Building with icon...
    rsrc -ico %ICON_FILE% -o %ICON_FILE%.syso
    go build %BUILD_FLAGS% -ldflags="%LD_FLAGS%" -o %OUTPUT_DIR%\%APP_NAME%.exe
    del %ICON_FILE%.syso
) else (
    echo Building without icon...
    go build %BUILD_FLAGS% -ldflags="%LD_FLAGS%" -o %OUTPUT_DIR%\%APP_NAME%.exe
)

:: Verify build
if exist "%OUTPUT_DIR%\%APP_NAME%.exe" (
    echo Build successful!
    echo Output: %OUTPUT_DIR%\%APP_NAME%.exe
    echo File size: %%~z %OUTPUT_DIR%\%APP_NAME%.exe bytes
) else (
    echo Build failed!
    exit /b 1
)

pause