@echo off
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
echo      Mencopot PHPForge Core Service dari Windows
echo =======================================================
echo.

cd /d "%~dp0"
echo Menghentikan service PHForgeCoreSvc...
sc stop PHForgeCoreSvc

echo Mencopot service...
phforge-core.exe --uninstall

echo.
echo [SELESAI] PHPForge Core Service telah dicopot.
echo.
pause
