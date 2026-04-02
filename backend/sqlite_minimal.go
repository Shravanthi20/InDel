package main

import (
	"fmt"
	"os"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	os.Remove("test_minimal.db") // Clean up any previous test file
	db, err := gorm.Open(sqlite.Open("test_minimal.db"), &gorm.Config{})
	if err != nil {
		fmt.Println("FAILED to open SQLite DB:", err)
		os.Exit(1)
	}
	fmt.Println("SUCCESS: Opened SQLite DB with GORM")
	db.Exec(`CREATE TABLE IF NOT EXISTS test_table (id INTEGER PRIMARY KEY, name TEXT)`)
	fmt.Println("SUCCESS: Created table")
	os.Remove("test_minimal.db") // Clean up after test
}
