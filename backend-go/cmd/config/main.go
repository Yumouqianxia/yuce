package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"backend-go/internal/config"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "validate":
		validateConfig()
	case "generate":
		generateConfig()
	case "diff":
		diffConfigs()
	case "export":
		exportConfig()
	case "import":
		importConfig()
	case "template":
		applyTemplate()
	case "health":
		checkHealth()
	case "profile":
		profileConfig()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Configuration Management Tool")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  config validate [file]           - Validate configuration file")
	fmt.Println("  config generate <env>            - Generate configuration template")
	fmt.Println("  config diff <file1> <file2>      - Compare two configuration files")
	fmt.Println("  config export <format> [file]    - Export configuration to format (json/yaml)")
	fmt.Println("  config import <file>             - Import configuration from file")
	fmt.Println("  config template <name> [file]    - Apply configuration template")
	fmt.Println("  config health [file]             - Check configuration health")
	fmt.Println("  config profile [file]            - Profile configuration loading")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  config validate configs/config.yaml")
	fmt.Println("  config generate development")
	fmt.Println("  config diff config.yaml config.prod.yaml")
	fmt.Println("  config export json config.json")
	fmt.Println("  config template production config.yaml")
}

func validateConfig() {
	var configFile string
	if len(os.Args) > 2 {
		configFile = os.Args[2]
	}

	var cfg *config.Config
	var err error

	if configFile != "" {
		cfg, err = config.LoadFromFile(configFile)
	} else {
		cfg, err = config.Load()
	}

	if err != nil {
		fmt.Printf("❌ Configuration validation failed: %v\n", err)

		// 尝试格式化验证错误
		if errors := config.FormatValidationErrors(err); len(errors) > 0 {
			fmt.Println("\nValidation errors:")
			for _, e := range errors {
				fmt.Printf("  - %s: %s\n", e.Field, e.Message)
			}
		}
		os.Exit(1)
	}

	fmt.Println("✅ Configuration is valid")

	// 显示配置摘要
	summary := config.GetConfigSummary(cfg)
	fmt.Println("\nConfiguration Summary:")
	for key, value := range summary {
		fmt.Printf("  %s: %v\n", key, value)
	}

	// 显示元数据
	metadata := config.GetConfigMetadata(cfg)
	fmt.Println("\nMetadata:")
	for key, value := range metadata {
		fmt.Printf("  %s: %v\n", key, value)
	}
}

func generateConfig() {
	if len(os.Args) < 3 {
		fmt.Println("Error: Environment is required")
		fmt.Println("Usage: config generate <environment>")
		fmt.Println("Environments: development, testing, production")
		os.Exit(1)
	}

	envStr := os.Args[2]
	var env config.Environment

	switch strings.ToLower(envStr) {
	case "dev", "development":
		env = config.EnvDevelopment
	case "test", "testing":
		env = config.EnvTesting
	case "prod", "production":
		env = config.EnvProduction
	default:
		fmt.Printf("Error: Unknown environment '%s'\n", envStr)
		os.Exit(1)
	}

	// 加载默认配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load default config: %v", err)
	}

	// 应用环境模板
	templateName := strings.ToLower(string(env))
	newCfg, err := config.ApplyTemplate(templateName, cfg)
	if err != nil {
		log.Fatalf("Failed to apply template: %v", err)
	}

	// 导出配置
	exporter := config.NewConfigExporter(newCfg)
	data, err := exporter.ExportToYAML()
	if err != nil {
		log.Fatalf("Failed to export config: %v", err)
	}

	// 生成文件名
	filename := fmt.Sprintf("config.%s.yaml", envStr)

	if err := os.WriteFile(filename, data, 0644); err != nil {
		log.Fatalf("Failed to write config file: %v", err)
	}

	fmt.Printf("✅ Generated configuration file: %s\n", filename)
}

func diffConfigs() {
	if len(os.Args) < 4 {
		fmt.Println("Error: Two configuration files are required")
		fmt.Println("Usage: config diff <file1> <file2>")
		os.Exit(1)
	}

	file1 := os.Args[2]
	file2 := os.Args[3]

	cfg1, err := config.LoadFromFile(file1)
	if err != nil {
		log.Fatalf("Failed to load config from %s: %v", file1, err)
	}

	cfg2, err := config.LoadFromFile(file2)
	if err != nil {
		log.Fatalf("Failed to load config from %s: %v", file2, err)
	}

	diffs := config.CompareConfigs(cfg1, cfg2)

	if len(diffs) == 0 {
		fmt.Println("✅ No differences found between configurations")
		return
	}

	fmt.Printf("Found %d differences:\n\n", len(diffs))

	for _, diff := range diffs {
		fmt.Printf("Field: %s\n", diff.Field)
		fmt.Printf("  %s: %v\n", filepath.Base(file1), diff.OldValue)
		fmt.Printf("  %s: %v\n", filepath.Base(file2), diff.NewValue)
		fmt.Printf("  Type: %s\n\n", diff.Type)
	}
}

func exportConfig() {
	if len(os.Args) < 3 {
		fmt.Println("Error: Export format is required")
		fmt.Println("Usage: config export <format> [output_file]")
		fmt.Println("Formats: json, yaml")
		os.Exit(1)
	}

	format := strings.ToLower(os.Args[2])
	var outputFile string
	if len(os.Args) > 3 {
		outputFile = os.Args[3]
	}

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	exporter := config.NewConfigExporter(cfg)

	if outputFile != "" {
		// 导出到文件
		if err := exporter.ExportToFile(outputFile); err != nil {
			log.Fatalf("Failed to export config to file: %v", err)
		}
		fmt.Printf("✅ Configuration exported to: %s\n", outputFile)
	} else {
		// 输出到标准输出
		var data []byte
		switch format {
		case "json":
			data, err = exporter.ExportToJSON()
		case "yaml", "yml":
			data, err = exporter.ExportToYAML()
		default:
			fmt.Printf("Error: Unsupported format '%s'\n", format)
			os.Exit(1)
		}

		if err != nil {
			log.Fatalf("Failed to export config: %v", err)
		}

		fmt.Println(string(data))
	}
}

func importConfig() {
	if len(os.Args) < 3 {
		fmt.Println("Error: Configuration file is required")
		fmt.Println("Usage: config import <file>")
		os.Exit(1)
	}

	filename := os.Args[2]

	importer := config.NewConfigImporter()
	cfg, err := importer.ImportFromFile(filename)
	if err != nil {
		log.Fatalf("Failed to import config: %v", err)
	}

	// 验证导入的配置
	validator := config.NewConfigValidator()
	if err := validator.Validate(cfg); err != nil {
		fmt.Printf("❌ Imported configuration is invalid: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Configuration imported successfully from: %s\n", filename)

	// 显示配置摘要
	summary := config.GetConfigSummary(cfg)
	fmt.Println("\nImported Configuration Summary:")
	for key, value := range summary {
		fmt.Printf("  %s: %v\n", key, value)
	}
}

func applyTemplate() {
	if len(os.Args) < 3 {
		fmt.Println("Error: Template name is required")
		fmt.Println("Usage: config template <name> [config_file]")

		templates := config.GetConfigTemplates()
		fmt.Println("\nAvailable templates:")
		for _, t := range templates {
			fmt.Printf("  %s - %s\n", t.Name, t.Description)
		}
		os.Exit(1)
	}

	templateName := os.Args[2]
	var configFile string
	if len(os.Args) > 3 {
		configFile = os.Args[3]
	}

	// 加载基础配置
	var baseCfg *config.Config
	var err error

	if configFile != "" {
		baseCfg, err = config.LoadFromFile(configFile)
	} else {
		baseCfg, err = config.Load()
	}

	if err != nil {
		log.Fatalf("Failed to load base config: %v", err)
	}

	// 应用模板
	newCfg, err := config.ApplyTemplate(templateName, baseCfg)
	if err != nil {
		log.Fatalf("Failed to apply template: %v", err)
	}

	// 导出结果
	exporter := config.NewConfigExporter(newCfg)
	data, err := exporter.ExportToYAML()
	if err != nil {
		log.Fatalf("Failed to export config: %v", err)
	}

	// 生成输出文件名
	outputFile := fmt.Sprintf("config.%s.yaml", templateName)
	if err := os.WriteFile(outputFile, data, 0644); err != nil {
		log.Fatalf("Failed to write config file: %v", err)
	}

	fmt.Printf("✅ Template '%s' applied successfully\n", templateName)
	fmt.Printf("Output saved to: %s\n", outputFile)
}

func checkHealth() {
	var configFile string
	if len(os.Args) > 2 {
		configFile = os.Args[2]
	}

	var cfg *config.Config
	var err error

	if configFile != "" {
		cfg, err = config.LoadFromFile(configFile)
	} else {
		cfg, err = config.Load()
	}

	if err != nil {
		fmt.Printf("❌ Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// 检查配置健康状态
	health := config.GetConfigHealth(cfg)

	status := health["status"].(string)
	if status == "healthy" {
		fmt.Println("✅ Configuration is healthy")
	} else {
		fmt.Println("❌ Configuration has issues")

		if issues, exists := health["issues"]; exists {
			fmt.Println("\nIssues found:")
			for _, issue := range issues.([]string) {
				fmt.Printf("  - %s\n", issue)
			}
		}
	}

	// 显示检查详情
	if checks, exists := health["checks"]; exists {
		fmt.Println("\nHealth Checks:")
		checksMap := checks.(map[string]interface{})
		for service, details := range checksMap {
			fmt.Printf("  %s: %v\n", service, details)
		}
	}
}

func profileConfig() {
	var configFile string
	if len(os.Args) > 2 {
		configFile = os.Args[2]
	}

	profiler := config.NewConfigProfiler()

	// 执行多次加载来收集性能数据
	fmt.Println("Profiling configuration loading...")

	for i := 0; i < 10; i++ {
		_, err := profiler.ProfileLoad(func() (*config.Config, error) {
			if configFile != "" {
				return config.LoadFromFile(configFile)
			}
			return config.Load()
		})

		if err != nil {
			log.Fatalf("Failed to load config during profiling: %v", err)
		}
	}

	// 显示性能统计
	stats := profiler.GetStats()
	fmt.Println("\nPerformance Statistics:")

	statsJSON, _ := json.MarshalIndent(stats, "", "  ")
	fmt.Println(string(statsJSON))
}
