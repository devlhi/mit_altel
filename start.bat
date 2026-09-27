@echo off
title Ali Telematika ISP - Backend API
rem ============================================================
rem  Menjalankan aplikasi Ali Telematika ISP (Golang API + web)
rem  Aplikasi dapat diakses di:
rem    - http://localhost:8080        (langsung)
rem    - http://alitelematika.test    (lewat nginx Laragon, jika aktif)
rem  Login pertama: admin / admin123
rem ============================================================
cd /d "%~dp0backend"

rem Cek jika sudah berjalan
tasklist /FI "IMAGENAME eq isp-api.exe" 2>nul | find /I "isp-api.exe" >nul
if %ERRORLEVEL%==0 (
    echo [!] Aplikasi sudah berjalan di http://localhost:8080
    echo     Jalankan stop.bat dulu bila ingin restart.
    pause
    exit /b
)

rem Terapkan binary pembaruan hanya setelah aplikasi lama dihentikan pengguna.
if exist "isp-api.next.exe" (
    echo [*] Menerapkan pembaruan aplikasi yang sudah disiapkan ...
    move /Y "isp-api.next.exe" "isp-api.exe" >nul
    if errorlevel 1 (
        echo [!] Pembaruan belum dapat diterapkan. Pastikan aplikasi sudah berhenti.
        pause
        exit /b 1
    )
)

echo [*] Menjalankan Ali Telematika ISP di http://localhost:8080 ...
echo [*] Tekan Ctrl+C untuk menghentikan.
rem 127.0.0.1 = hanya bisa diakses dari komputer ini; nginx Laragon tetap bisa proxy ke sini.
isp-api.exe -addr 127.0.0.1:8080 -db data\isp.db -static "..\frontend\.output\public"
pause
