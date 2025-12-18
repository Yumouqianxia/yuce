package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	fmt.Println("🔍 快速数据库迁移验证")
	fmt.Println("========================")

	// 尝试多个可能的路径
	possiblePaths := []string{
		"../backend-old/yuce_db.sqlite",
		"../../backend-old/yuce_db.sqlite",
		"./backend-old/yuce_db.sqlite",
		"backend-old/yuce_db.sqlite",
	}

	var foundPath string
	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			foundPath = path
			break
		}
	}

	if foundPath == "" {
		fmt.Println("❌ 未找到SQLite数据库文件")
		fmt.Println("请检查以下路径是否存在 yuce_db.sqlite 文件:")
		for _, path := range possiblePaths {
			absPath, _ := filepath.Abs(path)
			fmt.Printf("  - %s\n", absPath)
		}
		fmt.Println()
		fmt.Println("如果文件在其他位置，请将其复制到 backend-old/ 目录下")
		os.Exit(1)
	}

	fmt.Printf("✅ 找到SQLite数据库: %s\n", foundPath)

	// 获取文件信息
	fileInfo, err := os.Stat(foundPath)
	if err != nil {
		fmt.Printf("❌ 无法获取文件信息: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("📊 文件大小: %.2f KB\n", float64(fileInfo.Size())/1024)
	fmt.Printf("📅 修改时间: %s\n", fileInfo.ModTime().Format("2006-01-02 15:04:05"))

	if fileInfo.Size() == 0 {
		fmt.Println("❌ 数据库文件为空")
		os.Exit(1)
	}

	if fileInfo.Size() < 1024 {
		fmt.Println("⚠️  数据库文件很小，可能没有数据")
	}

	fmt.Println()
	fmt.Println("✅ 基本验证通过！")
	fmt.Println()
	fmt.Println("下一步验证:")
	fmt.Println("1. 运行详细验证: go run ./cmd/validate-db validate " + foundPath)
	fmt.Println("2. 或运行验证脚本: chmod +x scripts/validate-migration.sh && ./scripts/validate-migration.sh")
	fmt.Println()
	fmt.Println("如果需要进行数据迁移:")
	fmt.Println("1. 确保MySQL服务运行")
	fmt.Println("2. 配置 configs/config.yaml 中的数据库连接")
	fmt.Println("3. 运行: go run ./cmd/migrate up")
	fmt.Println("4. 运行: go run ./cmd/migrate import " + foundPath)
}
