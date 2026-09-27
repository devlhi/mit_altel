@echo off
title Ali Telematika ISP - Stop
taskkill /IM isp-api.exe /F >nul 2>&1
if %ERRORLEVEL%==0 (
    echo [OK] Ali Telematika ISP dihentikan.
) else (
    echo [!] Aplikasi sedang tidak berjalan.
)
pause
