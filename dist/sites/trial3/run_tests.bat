@echo off
setlocal enabledelayedexpansion
title PHPForge k6 Benchmark Runner - Trial 3
color 0b

:MENU
cls
echo ======================================================================
echo           PHPForge k6 Load Testing Suite - Target: Trial 3 (:8084)
echo ======================================================================
echo.
echo  Target URL: http://localhost:8084
echo.
echo  Pilih skenario pengujian beban:
echo  [1] Smoke Test      (5 VU   - 15 detik  - Verifikasi semua endpoint)
echo  [2] Load Test       (50 VU  - 50 detik  - Uji beban kerja normal)
echo  [3] Stress Test     (300 VU - 75 detik  - Uji batas antrean semaphore)
echo  [4] Spike Test      (250 VU - 45 detik  - Uji lonjakan instan/rush hour)
echo  [5] Custom URL      (Ubah target URL, default: http://localhost:8084)
echo  [0] Keluar
echo.
echo ======================================================================
set /p CHOICE="Pilihan Anda [1-5, 0]: "

if "%CHOICE%"=="1" goto RUN_SMOKE
if "%CHOICE%"=="2" goto RUN_LOAD
if "%CHOICE%"=="3" goto RUN_STRESS
if "%CHOICE%"=="4" goto RUN_SPIKE
if "%CHOICE%"=="5" goto SET_URL
if "%CHOICE%"=="0" goto EXIT
goto MENU

:RUN_SMOKE
cls
echo [MENJALANKAN SMOKE TEST...]
k6 run k6_smoke.js
echo.
pause
goto MENU

:RUN_LOAD
cls
echo [MENJALANKAN STANDARD LOAD TEST (50 VU)...]
k6 run k6_load.js
echo.
pause
goto MENU

:RUN_STRESS
cls
echo [MENJALANKAN HIGH-CONCURRENCY STRESS TEST (UP TO 300 VU)...]
k6 run k6_stress.js
echo.
pause
goto MENU

:RUN_SPIKE
cls
echo [MENJALANKAN SUDDEN SPIKE TEST (250 VU INSTANT)...]
k6 run k6_spike.js
echo.
pause
goto MENU

:SET_URL
cls
echo Masukkan target URL baru (misal: http://192.168.1.100:8084):
set /p TARGET_URL="URL: "
echo TARGET_URL telah diatur ke: %TARGET_URL%
echo.
pause
goto MENU

:EXIT
echo Selesai.
exit /b 0

