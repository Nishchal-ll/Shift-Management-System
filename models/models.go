package models

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDB() {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"))

	var err error
	// Retry logic for Docker
	for i := 0; i < 10; i++ {
		DB, err = sql.Open("postgres", connStr)
		if err == nil {
			if err = DB.Ping(); err == nil {
				fmt.Println("✅ Connected to PostgreSQL")
				return
			}
		}
		fmt.Println("⏳ Waiting for DB...", err)
		time.Sleep(2 * time.Second)
	}
	log.Fatal("❌ Could not connect to DB")
}
