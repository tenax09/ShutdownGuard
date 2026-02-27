@echo off
chcp 65001 > nul
echo 正在启动 ShutdownGuard 服务...
ShutdownGuard.exe start
if %errorlevel% equ 0 (
    echo.
    echo ✓ 服务已成功启动
) else (
    echo.
    echo ✗ 启动服务失败（可能需要管理员权限，请右键点击脚本以管理员身份运行）
)
pause
