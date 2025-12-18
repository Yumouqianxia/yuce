package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	// 这是一个简单的启动器，实际的应用程序在 cmd/api 目录下
	fmt.Println("Starting prediction system API server...")
	
	// 执行 cmd/api 下的主程序
	cmd := exec.Command("go", "run", "./cmd/api")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		fmt.Printf("Failed to start API server: %v\n", err)
		os.Exit(1)
	}
}