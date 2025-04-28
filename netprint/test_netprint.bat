@echo off
echo Testing Network Printer Image Sender

REM Check if an image file was provided
if "%~1"=="" (
    echo Usage: test_netprint.bat ^<image_file^>
    echo Example: test_netprint.bat sample.jpg
    exit /b 1
)

REM Check if the image file exists
if not exist "%~1" (
    echo Error: Image file "%~1" not found
    exit /b 1
)

echo.
echo 1. Testing with default Windows printer...
go run netprint.go "%~1"
if %ERRORLEVEL% neq 0 (
    echo Test failed with default Windows printer
) else (
    echo Test succeeded with default Windows printer
)

echo.
echo 2. To test with a specific Windows printer, run:
echo    go run netprint.go -printer "YOUR_PRINTER_NAME" "%~1"

echo.
echo 3. To test with a network printer using TCP/IP, run:
echo    go run netprint.go -type tcp -address YOUR_PRINTER_IP -port 9100 "%~1"

echo.
echo 4. To test with a network printer using IPP, run:
echo    go run netprint.go -type ipp -address YOUR_PRINTER_IP -port 631 "%~1"

echo.
echo Tests completed.