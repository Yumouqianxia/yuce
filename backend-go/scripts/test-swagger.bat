@echo off
REM 测试 Swagger 集成脚本 (Windows 版本)
REM 用于验证 API 文档和 Swagger UI 是否正常工作

echo 🚀 测试 Swagger 集成...

set SERVER_URL=http://localhost:8080

echo 📡 检查服务器状态...
curl -s --fail "%SERVER_URL%/health" >nul 2>&1
if errorlevel 1 (
    echo ❌ 服务器未运行，请先启动 API 服务器
    echo    运行: go run cmd/api/main.go
    exit /b 1
)

echo ✅ 服务器运行正常

echo 📖 测试 Swagger UI...
set SWAGGER_UI_URL=%SERVER_URL%/swagger/index.html
curl -s --fail "%SWAGGER_UI_URL%" >nul 2>&1
if errorlevel 1 (
    echo ❌ Swagger UI 不可访问
    exit /b 1
) else (
    echo ✅ Swagger UI 可访问: %SWAGGER_UI_URL%
)

echo 📄 测试 OpenAPI JSON 规范...
set OPENAPI_JSON_URL=%SERVER_URL%/swagger/doc.json
curl -s --fail "%OPENAPI_JSON_URL%" >nul 2>&1
if errorlevel 1 (
    echo ❌ OpenAPI JSON 规范无效或不可访问
    exit /b 1
) else (
    echo ✅ OpenAPI JSON 规范有效: %OPENAPI_JSON_URL%
)

echo 📚 测试 API 文档端点...
set API_DOCS_URL=%SERVER_URL%/api/docs
curl -s --fail "%API_DOCS_URL%" >nul 2>&1
if errorlevel 1 (
    echo ❌ API 文档端点异常
    exit /b 1
) else (
    echo ✅ API 文档端点正常: %API_DOCS_URL%
)

echo.
echo 🎉 Swagger 集成测试完成！
echo.
echo 📖 访问链接:
echo    • Swagger UI: %SWAGGER_UI_URL%
echo    • OpenAPI JSON: %OPENAPI_JSON_URL%
echo    • API 文档: %API_DOCS_URL%
echo.
echo 💡 提示:
echo    • 使用 Swagger UI 可以直接测试 API
echo    • 可以导入 OpenAPI JSON 到 Postman 或其他工具
echo    • 查看 docs/API_DOCUMENTATION.md 获取详细使用说明