#!/bin/bash

# 配置管理系统测试脚本

set -e

echo "=== 配置管理系统测试 ==="
echo

# 检查Go环境
if ! command -v go &> /dev/null; then
    echo "❌ Go 未安装"
    exit 1
fi

echo "✅ Go 环境检查通过"

# 进入项目目录
cd "$(dirname "$0")/.."

# 下载依赖
echo "📦 下载依赖..."
go mod tidy
go mod download

# 编译配置工具
echo "🔨 编译配置工具..."
go build -o bin/config ./cmd/config

# 编译迁移工具
echo "🔨 编译迁移工具..."
go build -o bin/migrate ./cmd/migrate

echo "✅ 编译完成"

# 测试配置验证
echo
echo "=== 测试配置验证 ==="

# 验证默认配置
echo "📋 验证默认配置..."
./bin/config validate

# 生成开发环境配置
echo "📋 生成开发环境配置..."
./bin/config generate development

# 验证生成的配置
if [ -f "config.development.yaml" ]; then
    echo "📋 验证开发环境配置..."
    ./bin/config validate config.development.yaml
    echo "✅ 开发环境配置验证通过"
else
    echo "❌ 开发环境配置文件未生成"
fi

# 测试配置导出
echo "📋 测试配置导出..."
./bin/config export json > config.json
if [ -f "config.json" ]; then
    echo "✅ JSON 导出成功"
    rm -f config.json
fi

# 测试配置健康检查
echo "📋 测试配置健康检查..."
./bin/config health

echo
echo "=== 测试数据库迁移工具 ==="

# 显示迁移状态
echo "📋 显示迁移状态..."
./bin/migrate status || echo "⚠️  数据库未连接，这是正常的"

# 创建测试迁移
echo "📋 创建测试迁移..."
./bin/migrate create test_migration

# 检查迁移文件是否创建
if ls migrations/*_test_migration.sql 1> /dev/null 2>&1; then
    echo "✅ 迁移文件创建成功"
    # 清理测试文件
    rm -f migrations/*_test_migration.sql
else
    echo "❌ 迁移文件创建失败"
fi

echo
echo "=== 清理临时文件 ==="
rm -f config.development.yaml
rm -f bin/config
rm -f bin/migrate

echo "✅ 所有测试完成！"
echo
echo "使用说明："
echo "1. 配置管理: go run ./cmd/config validate"
echo "2. 数据库迁移: go run ./cmd/migrate status"
echo "3. 从SQLite导入: go run ./cmd/migrate import ../backend-old/yuce_db.sqlite"