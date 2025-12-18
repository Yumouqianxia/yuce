#!/bin/bash

# 数据库迁移验证脚本

set -e

echo "=== 数据库迁移验证 ==="
echo

# 检查SQLite文件是否存在
SQLITE_FILE="../backend-old/yuce_db.sqlite"

if [ ! -f "$SQLITE_FILE" ]; then
    echo "❌ SQLite数据库文件未找到: $SQLITE_FILE"
    echo "请确保旧项目的数据库文件存在"
    exit 1
fi

echo "✅ 找到SQLite数据库文件: $SQLITE_FILE"

# 进入项目目录
cd "$(dirname "$0")/.."

# 下载依赖
echo "📦 下载依赖..."
go mod tidy

# 编译验证工具
echo "🔨 编译数据库验证工具..."
go build -o bin/validate-db ./cmd/validate-db

echo "✅ 编译完成"
echo

# 运行数据库检查
echo "=== 数据库结构检查 ==="
echo "🔍 检查SQLite数据库结构..."
./bin/validate-db inspect "$SQLITE_FILE"

echo
echo "=== 数据库模式检查 ==="
echo "📋 显示数据库模式..."
./bin/validate-db schema "$SQLITE_FILE"

echo
echo "=== 示例数据检查 ==="
echo "📊 显示示例数据..."
./bin/validate-db data "$SQLITE_FILE"

echo
echo "=== 迁移兼容性验证 ==="
echo "✅ 验证迁移兼容性..."
./bin/validate-db validate "$SQLITE_FILE"

echo
echo "=== 编译迁移工具 ==="
echo "🔨 编译数据库迁移工具..."
go build -o bin/migrate ./cmd/migrate

echo "✅ 迁移工具编译完成"

echo
echo "=== 测试配置加载 ==="
echo "📋 测试配置系统..."
go build -o bin/config ./cmd/config
./bin/config validate

echo
echo "=== 清理临时文件 ==="
rm -f bin/validate-db
rm -f bin/migrate  
rm -f bin/config

echo
echo "✅ 数据库迁移验证完成！"
echo
echo "如果所有检查都通过，可以执行以下步骤进行实际迁移："
echo "1. 确保MySQL服务正在运行"
echo "2. 更新 configs/config.yaml 中的数据库配置"
echo "3. 运行: go run ./cmd/migrate up"
echo "4. 运行: go run ./cmd/migrate import $SQLITE_FILE"
echo "5. 运行: go run ./cmd/migrate status"