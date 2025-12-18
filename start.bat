@echo off
REM 预测系统 - Windows 启动脚本
REM 使用方法: start.bat

echo ========================================
echo   预测系统 - Docker 启动脚本
echo ========================================
echo.

REM 检查 Docker 是否运行
docker info >nul 2>&1
if errorlevel 1 (
    echo [错误] Docker 未运行，请先启动 Docker Desktop
    pause
    exit /b 1
)

echo [1/4] 检查 Docker 环境...
docker-compose --version >nul 2>&1
if errorlevel 1 (
    echo [错误] docker-compose 未安装
    pause
    exit /b 1
)
echo [✓] Docker 环境正常

echo.
echo [2/4] 停止旧容器...
docker-compose -f docker-compose.hub.yml down >nul 2>&1
docker-compose -f docker-compose.dev.yml down >nul 2>&1
echo [✓] 旧容器已停止

echo.
echo [3/4] 启动服务...
docker-compose -f docker-compose.hub.yml --profile local up -d
if errorlevel 1 (
    echo [错误] 服务启动失败
    pause
    exit /b 1
)

echo.
echo [4/4] 等待服务就绪...
timeout /t 5 /nobreak >nul

echo.
echo ========================================
echo   服务启动成功！
echo ========================================
echo.
echo 访问地址:
echo   前端: http://localhost:5408
echo   后端: http://localhost:1874
echo   数据库管理: http://localhost:8082
echo.
echo 容器状态:
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
echo.
echo 查看日志: docker-compose -f docker-compose.hub.yml logs -f
echo 停止服务: docker-compose -f docker-compose.hub.yml down
echo.
pause
