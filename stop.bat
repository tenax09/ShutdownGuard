@echo off
chcp 65001 > nul
echo 正在停止 ShutdownGuard 服务...
ShutdownGuard.exe stop
if %errorlevel% equ 0 (
    echo.
    echo ✓ 服务已成功停止
) else (
    echo.
    echo ✗ 停止服务失败（可能需要管理员权限，请右键点击脚本以管理员身份运行）
)
pause
