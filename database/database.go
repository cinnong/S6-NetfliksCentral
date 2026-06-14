package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var db *sql.DB

func InitDB() {
	// Ambil konfigurasi dari environment variable
	host := os.Getenv("PGHOST")
	user := os.Getenv("PGUSER")
	password := os.Getenv("PGPASSWORD")
	dbname := os.Getenv("PGDATABASE")
	port := os.Getenv("PGPORT")
	sslmode := os.Getenv("PGSSLMODE")
	if sslmode == "" {
		sslmode = "disable"
	}

	if host == "" {
		host = "localhost"
	}
	if port == "" {
		port = "5432"
	}
	if user == "" {
		user = "postgres"
	}
	if dbname == "" {
		dbname = "netflixdb"
	}
	if password == "" {
		password = "dina2004"
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s", host, user, password, dbname, port, sslmode)
	sqlDB, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}

	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	// Buat tabel jika belum ada (berguna saat teman Anda clone project ini)
	schemas := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			email VARCHAR(255) NOT NULL UNIQUE,
			password_hash VARCHAR(255) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS accounts (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			label VARCHAR(255) NOT NULL,
			netflix_email VARCHAR(255) NOT NULL,
			status VARCHAR(50) NOT NULL DEFAULT 'active',
			chrome_profile VARCHAR(255) NOT NULL UNIQUE,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS tabs (
			id SERIAL PRIMARY KEY,
			account_id INTEGER NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
			title VARCHAR(255) NOT NULL,
			url TEXT NOT NULL,
			position INTEGER NOT NULL
		);`,
		`CREATE INDEX IF NOT EXISTS idx_accounts_user_id ON accounts(user_id);`,
		`CREATE INDEX IF NOT EXISTS idx_tabs_account_id ON tabs(account_id);`,
	}

	for _, stmt := range schemas {
		if _, err := sqlDB.Exec(stmt); err != nil {
			log.Fatalf("failed to apply database schema: %v", err)
		}
	}

	db = sqlDB
}

func GetDB() *sql.DB {
	return db
}
