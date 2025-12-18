@echo off
REM 预测系统 - Windows 停止脚本

echo ========================================
echo   预测系统 - 停止服务
echo ========================================
echo.

echo 正在停止所有服务...
docker-compose -f docker-compose.hub.yml down

echo.
echo [✓] 服务已停止
echo.
pause
