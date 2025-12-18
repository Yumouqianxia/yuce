package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "inspect":
		inspectSQLite()
	case "schema":
		showSchema()
	case "data":
		showData()
	case "validate":
		validateForMigration()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("SQLite Database Validation Tool")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  validate-db inspect <sqlite_file>    - Inspect database structure")
	fmt.Println("  validate-db schema <sqlite_file>     - Show table schemas")
	fmt.Println("  validate-db data <sqlite_file>       - Show data samples")
	fmt.Println("  validate-db validate <sqlite_file>   - Validate for migration")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  validate-db inspect ../backend-old/yuce_db.sqlite")
	fmt.Println("  validate-db validate ../backend-old/yuce_db.sqlite")
}

func inspectSQLite() {
	if len(os.Args) < 3 {
		fmt.Println("Error: SQLite file path is required")
		fmt.Println("Usage: validate-db inspect <sqlite_file>")
		os.Exit(1)
	}

	sqliteFile := os.Args[2]

	if _, err := os.Stat(sqliteFile); os.IsNotExist(err) {
		log.Fatalf("SQLite file not found: %s", sqliteFile)
	}

	fmt.Printf("🔍 Inspecting SQLite database: %s\n", sqliteFile)
	fmt.Println(strings.Repeat("=", 60))

	db, err := sql.Open("sqlite3", sqliteFile)
	if err != nil {
		log.Fatalf("Failed to open SQLite database: %v", err)
	}
	defer db.Close()

	// 检查数据库连接
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	fmt.Println("✅ Database connection successful")
	fmt.Println()

	// 获取所有表
	tables, err := getTables(db)
	if err != nil {
		log.Fatalf("Failed to get tables: %v", err)
	}

	fmt.Printf("📊 Found %d tables:\n", len(tables))
	for i, table := range tables {
		fmt.Printf("  %d. %s\n", i+1, table)
	}
	fmt.Println()

	// 显示每个表的信息
	for _, table := range tables {
		showTableInfo(db, table)
	}
}

func showSchema() {
	if len(os.Args) < 3 {
		fmt.Println("Error: SQLite file path is required")
		os.Exit(1)
	}

	sqliteFile := os.Args[2]

	db, err := sql.Open("sqlite3", sqliteFile)
	if err != nil {
		log.Fatalf("Failed to open SQLite database: %v", err)
	}
	defer db.Close()

	fmt.Printf("📋 Database Schema: %s\n", sqliteFile)
	fmt.Println(strings.Repeat("=", 60))

	tables, err := getTables(db)
	if err != nil {
		log.Fatalf("Failed to get tables: %v", err)
	}

	for _, table := range tables {
		showTableSchema(db, table)
	}
}

func showData() {
	if len(os.Args) < 3 {
		fmt.Println("Error: SQLite file path is required")
		os.Exit(1)
	}

	sqliteFile := os.Args[2]

	db, err := sql.Open("sqlite3", sqliteFile)
	if err != nil {
		log.Fatalf("Failed to open SQLite database: %v", err)
	}
	defer db.Close()

	fmt.Printf("📊 Sample Data: %s\n", sqliteFile)
	fmt.Println(strings.Repeat("=", 60))

	tables, err := getTables(db)
	if err != nil {
		log.Fatalf("Failed to get tables: %v", err)
	}

	for _, table := range tables {
		showSampleData(db, table)
	}
}

func validateForMigration() {
	if len(os.Args) < 3 {
		fmt.Println("Error: SQLite file path is required")
		os.Exit(1)
	}

	sqliteFile := os.Args[2]

	db, err := sql.Open("sqlite3", sqliteFile)
	if err != nil {
		log.Fatalf("Failed to open SQLite database: %v", err)
	}
	defer db.Close()

	fmt.Printf("✅ Migration Validation: %s\n", sqliteFile)
	fmt.Println(strings.Repeat("=", 60))

	// 检查必需的表
	requiredTables := []string{"users", "matches", "predictions", "votes", "prediction_modifications"}

	tables, err := getTables(db)
	if err != nil {
		log.Fatalf("Failed to get tables: %v", err)
	}

	tableMap := make(map[string]bool)
	for _, table := range tables {
		tableMap[table] = true
	}

	allTablesExist := true
	fmt.Println("📋 Required Tables Check:")
	for _, table := range requiredTables {
		if tableMap[table] {
			fmt.Printf("  ✅ %s - Found\n", table)
		} else {
			fmt.Printf("  ❌ %s - Missing\n", table)
			allTablesExist = false
		}
	}
	fmt.Println()

	if !allTablesExist {
		fmt.Println("❌ Some required tables are missing. Migration may not work correctly.")
		return
	}

	// 检查数据完整性
	fmt.Println("🔍 Data Integrity Check:")

	// 检查用户表
	userCount := getRowCount(db, "users")
	fmt.Printf("  Users: %d records\n", userCount)

	// 检查比赛表
	matchCount := getRowCount(db, "matches")
	fmt.Printf("  Matches: %d records\n", matchCount)

	// 检查预测表
	predictionCount := getRowCount(db, "predictions")
	fmt.Printf("  Predictions: %d records\n", predictionCount)

	// 检查投票表
	voteCount := getRowCount(db, "votes")
	fmt.Printf("  Votes: %d records\n", voteCount)

	// 检查预测修改记录表
	modificationCount := getRowCount(db, "prediction_modifications")
	fmt.Printf("  Prediction Modifications: %d records\n", modificationCount)

	fmt.Println()

	// 检查关联关系
	fmt.Println("🔗 Relationship Validation:")

	// 检查预测表中的用户ID是否都存在
	orphanPredictions := checkOrphanRecords(db, "predictions", "userId", "users", "id")
	if orphanPredictions > 0 {
		fmt.Printf("  ❌ Found %d predictions with invalid user IDs\n", orphanPredictions)
	} else {
		fmt.Printf("  ✅ All predictions have valid user IDs\n")
	}

	// 检查预测表中的比赛ID是否都存在
	orphanPredictionsMatch := checkOrphanRecords(db, "predictions", "matchId", "matches", "id")
	if orphanPredictionsMatch > 0 {
		fmt.Printf("  ❌ Found %d predictions with invalid match IDs\n", orphanPredictionsMatch)
	} else {
		fmt.Printf("  ✅ All predictions have valid match IDs\n")
	}

	// 检查投票表中的用户ID是否都存在
	orphanVotesUser := checkOrphanRecords(db, "votes", "user_id", "users", "id")
	if orphanVotesUser > 0 {
		fmt.Printf("  ❌ Found %d votes with invalid user IDs\n", orphanVotesUser)
	} else {
		fmt.Printf("  ✅ All votes have valid user IDs\n")
	}

	// 检查投票表中的预测ID是否都存在
	orphanVotesPrediction := checkOrphanRecords(db, "votes", "prediction_id", "predictions", "id")
	if orphanVotesPrediction > 0 {
		fmt.Printf("  ❌ Found %d votes with invalid prediction IDs\n", orphanVotesPrediction)
	} else {
		fmt.Printf("  ✅ All votes have valid prediction IDs\n")
	}

	fmt.Println()

	// 检查数据类型兼容性
	fmt.Println("🔄 Data Type Compatibility:")
	checkDataTypes(db)

	fmt.Println()
	fmt.Println("✅ Migration validation completed!")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("1. Run: go run ./cmd/migrate up")
	fmt.Println("2. Run: go run ./cmd/migrate import " + sqliteFile)
	fmt.Println("3. Run: go run ./cmd/migrate validate")
}

// 辅助函数

func getTables(db *sql.DB) ([]string, error) {
	query := "SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'"
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return nil, err
		}
		tables = append(tables, table)
	}

	return tables, nil
}

func showTableInfo(db *sql.DB, tableName string) {
	fmt.Printf("📋 Table: %s\n", tableName)
	fmt.Println(strings.Repeat("-", 40))

	// 获取行数
	count := getRowCount(db, tableName)
	fmt.Printf("  Records: %d\n", count)

	// 获取列信息
	columns, err := getColumns(db, tableName)
	if err != nil {
		fmt.Printf("  Error getting columns: %v\n", err)
		return
	}

	fmt.Printf("  Columns: %d\n", len(columns))
	for _, col := range columns {
		fmt.Printf("    - %s (%s)\n", col.Name, col.Type)
	}

	fmt.Println()
}

func showTableSchema(db *sql.DB, tableName string) {
	fmt.Printf("CREATE TABLE %s:\n", tableName)

	query := fmt.Sprintf("SELECT sql FROM sqlite_master WHERE type='table' AND name='%s'", tableName)
	var schema string
	err := db.QueryRow(query).Scan(&schema)
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
		return
	}

	fmt.Println(schema)
	fmt.Println()
}

func showSampleData(db *sql.DB, tableName string) {
	fmt.Printf("📊 Sample data from %s (first 3 rows):\n", tableName)

	query := fmt.Sprintf("SELECT * FROM %s LIMIT 3", tableName)
	rows, err := db.Query(query)
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
		return
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		fmt.Printf("  Error getting columns: %v\n", err)
		return
	}

	fmt.Printf("  Columns: %s\n", strings.Join(columns, " | "))
	fmt.Printf("  %s\n", strings.Repeat("-", len(strings.Join(columns, " | "))+10))

	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range columns {
		valuePtrs[i] = &values[i]
	}

	rowCount := 0
	for rows.Next() && rowCount < 3 {
		err := rows.Scan(valuePtrs...)
		if err != nil {
			fmt.Printf("  Error scanning row: %v\n", err)
			continue
		}

		var rowData []string
		for _, val := range values {
			if val == nil {
				rowData = append(rowData, "NULL")
			} else {
				rowData = append(rowData, fmt.Sprintf("%v", val))
			}
		}
		fmt.Printf("  %s\n", strings.Join(rowData, " | "))
		rowCount++
	}

	fmt.Println()
}

func getRowCount(db *sql.DB, tableName string) int {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)
	var count int
	err := db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0
	}
	return count
}

type Column struct {
	Name string
	Type string
}

func getColumns(db *sql.DB, tableName string) ([]Column, error) {
	query := fmt.Sprintf("PRAGMA table_info(%s)", tableName)
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []Column
	for rows.Next() {
		var cid int
		var name, dataType string
		var notNull int
		var defaultValue interface{}
		var pk int

		err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &pk)
		if err != nil {
			return nil, err
		}

		columns = append(columns, Column{
			Name: name,
			Type: dataType,
		})
	}

	return columns, nil
}

func checkOrphanRecords(db *sql.DB, childTable, childColumn, parentTable, parentColumn string) int {
	query := fmt.Sprintf(`
		SELECT COUNT(*) 
		FROM %s c 
		LEFT JOIN %s p ON c.%s = p.%s 
		WHERE p.%s IS NULL AND c.%s IS NOT NULL
	`, childTable, parentTable, childColumn, parentColumn, parentColumn, childColumn)

	var count int
	err := db.QueryRow(query).Scan(&count)
	if err != nil {
		return -1
	}
	return count
}

func checkDataTypes(db *sql.DB) {
	// 检查用户表的数据类型
	fmt.Println("  Users table:")
	checkUserDataTypes(db)

	// 检查比赛表的数据类型
	fmt.Println("  Matches table:")
	checkMatchDataTypes(db)

	// 检查预测表的数据类型
	fmt.Println("  Predictions table:")
	checkPredictionDataTypes(db)
}

func checkUserDataTypes(db *sql.DB) {
	// 检查用户名长度
	query := "SELECT MAX(LENGTH(username)) FROM users"
	var maxLen sql.NullInt64
	db.QueryRow(query).Scan(&maxLen)
	if maxLen.Valid {
		if maxLen.Int64 > 50 {
			fmt.Printf("    ⚠️  Username max length: %d (exceeds 50)\n", maxLen.Int64)
		} else {
			fmt.Printf("    ✅ Username max length: %d\n", maxLen.Int64)
		}
	}

	// 检查邮箱长度
	query = "SELECT MAX(LENGTH(email)) FROM users"
	db.QueryRow(query).Scan(&maxLen)
	if maxLen.Valid {
		if maxLen.Int64 > 255 {
			fmt.Printf("    ⚠️  Email max length: %d (exceeds 255)\n", maxLen.Int64)
		} else {
			fmt.Printf("    ✅ Email max length: %d\n", maxLen.Int64)
		}
	}
}

func checkMatchDataTypes(db *sql.DB) {
	// 检查标题长度
	query := "SELECT MAX(LENGTH(title)) FROM matches"
	var maxLen sql.NullInt64
	db.QueryRow(query).Scan(&maxLen)
	if maxLen.Valid {
		if maxLen.Int64 > 255 {
			fmt.Printf("    ⚠️  Title max length: %d (exceeds 255)\n", maxLen.Int64)
		} else {
			fmt.Printf("    ✅ Title max length: %d\n", maxLen.Int64)
		}
	}

	// 检查选项长度
	query = "SELECT MAX(LENGTH(optionA)) FROM matches"
	db.QueryRow(query).Scan(&maxLen)
	if maxLen.Valid {
		if maxLen.Int64 > 255 {
			fmt.Printf("    ⚠️  OptionA max length: %d (exceeds 255)\n", maxLen.Int64)
		} else {
			fmt.Printf("    ✅ OptionA max length: %d\n", maxLen.Int64)
		}
	}
}

func checkPredictionDataTypes(db *sql.DB) {
	// 检查预测获胜者的值
	query := "SELECT DISTINCT predictedWinner FROM predictions"
	rows, err := db.Query(query)
	if err != nil {
		return
	}
	defer rows.Close()

	validWinners := true
	var winners []string
	for rows.Next() {
		var winner string
		rows.Scan(&winner)
		winners = append(winners, winner)
		if winner != "A" && winner != "B" {
			validWinners = false
		}
	}

	if validWinners {
		fmt.Printf("    ✅ Predicted winners: %v\n", winners)
	} else {
		fmt.Printf("    ⚠️  Invalid predicted winners found: %v\n", winners)
	}
}
