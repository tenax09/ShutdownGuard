@echo off
chcp 65001 > nul
echo 正在卸载 ShutdownGuard 服务...
echo.
ShutdownGuard.exe uninstall
if %errorlevel% equ 0 (
    echo.
    echo ✓ 服务已成功卸载
) else (
    echo.
    echo ✗ 卸载失败，请以管理员身份运行此脚本
    echo.
    echo 操作方法：
    echo 1. 右键点击此脚本
    echo 2. 选择 以管理员身份运行
)
pause
