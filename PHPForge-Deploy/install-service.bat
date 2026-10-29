@echo off
:: Cek hak akses Administrator
net session >nul 2>&1
if %errorLevel% neq 0 (
    echo =======================================================
    echo [ERROR] Script ini membutuhkan hak akses Administrator!
    echo Silakan klik kanan file ini dan pilih "Run as administrator".
    echo =======================================================
    echo.
    pause
    exit /b 1
)

echo =======================================================
echo      Memasang PHPForge Core Service ke Windows
echo =======================================================
echo.

cd /d "%~dp0"
phforge-core.exe --install

if %errorLevel% equ 0 (
    echo.
    echo Menyalakan service PHForgeCoreSvc...
    sc start PHForgeCoreSvc
    sc config PHForgeCoreSvc start= auto
    echo.
    echo [SUKSES] PHPForge Core Service telah aktif dan diatur Automatic Startup!
) else (
    echo.
    echo [PERINGATAN] Service mungkin sudah terpasang. Mencoba menyalakan...
    sc start PHForgeCoreSvc
)

echo.
pause

