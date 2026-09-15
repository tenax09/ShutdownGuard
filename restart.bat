@echo off
cd /d "%~dp0"
echo 正在重启 ShutdownGuard 服务...
echo.
echo 正在停止服务...
ShutdownGuard.exe stop
timeout /t 2 /nobreak
echo.
echo 正在启动服务...
ShutdownGuard.exe start
echo.
echo [OK] 服务已重启
pause
