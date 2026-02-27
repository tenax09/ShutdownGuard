@echo off
chcp 65001 > nul
echo 正在安装 ShutdownGuard 服务...
echo.
ShutdownGuard.exe install
if %errorlevel% equ 0 (
    echo.
    echo ✓ 服务已成功安装（开机自启）
    echo.
    echo 将在 3 秒后自动启动服务...
    timeout /t 3 /nobreak
    echo.
    echo 正在启动服务...
    ShutdownGuard.exe start
    echo.
    echo ✓ 服务已启动
) else (
    echo.
    echo ✗ 安装失败，请以管理员身份运行此脚本
    echo.
    echo 操作方法：
    echo 1. 右键点击此脚本
    echo 2. 选择 以管理员身份运行
)
pause
