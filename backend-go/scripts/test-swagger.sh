#!/bin/bash

# 测试 Swagger 集成脚本
# 用于验证 API 文档和 Swagger UI 是否正常工作

set -e

echo "🚀 测试 Swagger 集成..."

# 检查服务器是否运行
SERVER_URL="http://localhost:8080"
echo "📡 检查服务器状态..."

if ! curl -s --fail "$SERVER_URL/health" > /dev/null; then
    echo "❌ 服务器未运行，请先启动 API 服务器"
    echo "   运行: go run cmd/api/main.go"
    exit 1
fi

echo "✅ 服务器运行正常"

# 测试 Swagger UI 访问
echo "📖 测试 Swagger UI..."
SWAGGER_UI_URL="$SERVER_URL/swagger/index.html"

if curl -s --fail "$SWAGGER_UI_URL" > /dev/null; then
    echo "✅ Swagger UI 可访问: $SWAGGER_UI_URL"
else
    echo "❌ Swagger UI 不可访问"
    exit 1
fi

# 测试 OpenAPI JSON 规范
echo "📄 测试 OpenAPI JSON 规范..."
OPENAPI_JSON_URL="$SERVER_URL/swagger/doc.json"

if curl -s --fail "$OPENAPI_JSON_URL" | jq . > /dev/null 2>&1; then
    echo "✅ OpenAPI JSON 规范有效: $OPENAPI_JSON_URL"
else
    echo "❌ OpenAPI JSON 规范无效或不可访问"
    exit 1
fi

# 测试 API 文档端点
echo "📚 测试 API 文档端点..."
API_DOCS_URL="$SERVER_URL/api/docs"

if curl -s --fail "$API_DOCS_URL" | jq . > /dev/null 2>&1; then
    echo "✅ API 文档端点正常: $API_DOCS_URL"
else
    echo "❌ API 文档端点异常"
    exit 1
fi

# 检查关键 API 端点是否在文档中
echo "🔍 检查关键 API 端点..."
OPENAPI_CONTENT=$(curl -s "$OPENAPI_JSON_URL")

# 检查认证端点
if echo "$OPENAPI_CONTENT" | jq -e '.paths["/auth/login"]' > /dev/null; then
    echo "✅ 认证端点已文档化"
else
    echo "⚠️  认证端点未在文档中找到"
fi

# 检查比赛端点
if echo "$OPENAPI_CONTENT" | jq -e '.paths["/matches"]' > /dev/null; then
    echo "✅ 比赛端点已文档化"
else
    echo "⚠️  比赛端点未在文档中找到"
fi

# 检查预测端点
if echo "$OPENAPI_CONTENT" | jq -e '.paths["/predictions"]' > /dev/null; then
    echo "✅ 预测端点已文档化"
else
    echo "⚠️  预测端点未在文档中找到"
fi

# 检查排行榜端点
if echo "$OPENAPI_CONTENT" | jq -e '.paths["/leaderboard"]' > /dev/null; then
    echo "✅ 排行榜端点已文档化"
else
    echo "⚠️  排行榜端点未在文档中找到"
fi

# 检查安全定义
if echo "$OPENAPI_CONTENT" | jq -e '.securityDefinitions.BearerAuth' > /dev/null; then
    echo "✅ JWT 认证已配置"
else
    echo "⚠️  JWT 认证配置未找到"
fi

# 统计 API 端点数量
ENDPOINT_COUNT=$(echo "$OPENAPI_CONTENT" | jq '.paths | keys | length')
echo "📊 已文档化的 API 端点数量: $ENDPOINT_COUNT"

# 统计数据模型数量
MODEL_COUNT=$(echo "$OPENAPI_CONTENT" | jq '.definitions | keys | length')
echo "📊 已定义的数据模型数量: $MODEL_COUNT"

echo ""
echo "🎉 Swagger 集成测试完成！"
echo ""
echo "📖 访问链接:"
echo "   • Swagger UI: $SWAGGER_UI_URL"
echo "   • OpenAPI JSON: $OPENAPI_JSON_URL"
echo "   • API 文档: $API_DOCS_URL"
echo ""
echo "💡 提示:"
echo "   • 使用 Swagger UI 可以直接测试 API"
echo "   • 可以导入 OpenAPI JSON 到 Postman 或其他工具"
echo "   • 查看 docs/API_DOCUMENTATION.md 获取详细使用说明"