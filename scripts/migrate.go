package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	log.Println("Starting SQLite to PostgreSQL database migration...")

	// 1. Open SQLite database
	dbDir := "database"
	sqlitePath := filepath.Join(dbDir, "app.db")
	if _, err := os.Stat(sqlitePath); os.IsNotExist(err) {
		log.Fatalf("Error: SQLite file %s not found. Please make sure the database/app.db file exists.", sqlitePath)
	}

	sqliteDB, err := sql.Open("sqlite3", sqlitePath)
	if err != nil {
		log.Fatalf("Failed to open SQLite database: %v", err)
	}
	defer sqliteDB.Close()

	// 2. Fetch connection parameters from environment variables (with fallback defaults)
	host := os.Getenv("PGHOST")
	user := os.Getenv("PGUSER")
	password := os.Getenv("PGPASSWORD")
	dbname := os.Getenv("PGDATABASE")
	port := os.Getenv("PGPORT")
	sslmode := os.Getenv("PGSSLMODE")

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
	if sslmode == "" {
		sslmode = "disable"
	}

	// Print connection info (excluding password for safety)
	log.Printf("Connecting to PostgreSQL at %s:%s (database: %s, user: %s)...", host, port, dbname, user)

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s", host, user, password, dbname, port, sslmode)
	pgDB, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to open PostgreSQL connection: %v", err)
	}
	defer pgDB.Close()

	if err := pgDB.Ping(); err != nil {
		log.Fatalf("Failed to connect to PostgreSQL database: %v. Please verify your connection credentials and check if PostgreSQL is running.", err)
	}
	log.Println("Successfully connected to PostgreSQL database.")

	// 3. Create tables in PostgreSQL if they do not exist
	log.Println("Creating schemas in PostgreSQL (if not exists)...")
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
		if _, err := pgDB.Exec(stmt); err != nil {
			log.Fatalf("Failed to apply database schema: %v", err)
		}
	}
	log.Println("PostgreSQL schemas successfully applied.")

	// 4. Migrate 'users'
	log.Println("Migrating 'users' table...")
	userRows, err := sqliteDB.Query("SELECT id, email, password_hash, created_at FROM users")
	if err != nil {
		log.Fatalf("Failed to query SQLite 'users': %v", err)
	}
	defer userRows.Close()

	userCount := 0
	for userRows.Next() {
		var id int64
		var email, passwordHash string
		var createdAtRaw string
		if err := userRows.Scan(&id, &email, &passwordHash, &createdAtRaw); err != nil {
			log.Fatalf("Failed to scan SQLite user: %v", err)
		}

		createdAt := parseTime(createdAtRaw)

		_, err = pgDB.Exec(
			`INSERT INTO users (id, email, password_hash, created_at) VALUES ($1, $2, $3, $4)
			 ON CONFLICT (id) DO UPDATE SET email = EXCLUDED.email, password_hash = EXCLUDED.password_hash, created_at = EXCLUDED.created_at`,
			id, email, passwordHash, createdAt,
		)
		if err != nil {
			log.Fatalf("Failed to insert user %s into PostgreSQL: %v", email, err)
		}
		userCount++
	}
	log.Printf("Migrated %d users.", userCount)

	// 5. Migrate 'accounts'
	log.Println("Migrating 'accounts' table...")
	accountRows, err := sqliteDB.Query("SELECT id, user_id, label, netflix_email, status, chrome_profile, created_at FROM accounts")
	if err != nil {
		log.Fatalf("Failed to query SQLite 'accounts': %v", err)
	}
	defer accountRows.Close()

	accountCount := 0
	for accountRows.Next() {
		var id, userID int64
		var label, netflixEmail, status, chromeProfile, createdAtRaw string
		if err := accountRows.Scan(&id, &userID, &label, &netflixEmail, &status, &chromeProfile, &createdAtRaw); err != nil {
			log.Fatalf("Failed to scan SQLite account: %v", err)
		}

		createdAt := parseTime(createdAtRaw)

		_, err = pgDB.Exec(
			`INSERT INTO accounts (id, user_id, label, netflix_email, status, chrome_profile, created_at) 
			 VALUES ($1, $2, $3, $4, $5, $6, $7)
			 ON CONFLICT (id) DO UPDATE SET user_id = EXCLUDED.user_id, label = EXCLUDED.label, 
			 netflix_email = EXCLUDED.netflix_email, status = EXCLUDED.status, 
			 chrome_profile = EXCLUDED.chrome_profile, created_at = EXCLUDED.created_at`,
			id, userID, label, netflixEmail, status, chromeProfile, createdAt,
		)
		if err != nil {
			log.Fatalf("Failed to insert account %s into PostgreSQL: %v", label, err)
		}
		accountCount++
	}
	log.Printf("Migrated %d accounts.", accountCount)

	// 6. Migrate 'tabs'
	log.Println("Migrating 'tabs' table...")
	tabRows, err := sqliteDB.Query("SELECT id, account_id, title, url, position FROM tabs")
	if err != nil {
		log.Fatalf("Failed to query SQLite 'tabs': %v", err)
	}
	defer tabRows.Close()

	tabCount := 0
	for tabRows.Next() {
		var id, accountID int64
		var title, url string
		var position int
		if err := tabRows.Scan(&id, &accountID, &title, &url, &position); err != nil {
			log.Fatalf("Failed to scan SQLite tab: %v", err)
		}

		_, err = pgDB.Exec(
			`INSERT INTO tabs (id, account_id, title, url, position) VALUES ($1, $2, $3, $4, $5)
			 ON CONFLICT (id) DO UPDATE SET account_id = EXCLUDED.account_id, title = EXCLUDED.title, 
			 url = EXCLUDED.url, position = EXCLUDED.position`,
			id, accountID, title, url, position,
		)
		if err != nil {
			log.Fatalf("Failed to insert tab %s into PostgreSQL: %v", title, err)
		}
		tabCount++
	}
	log.Printf("Migrated %d tabs.", tabCount)

	// 7. Reset auto-increment sequences in PostgreSQL
	log.Println("Resetting auto-increment sequences...")
	resetSequence(pgDB, "users")
	resetSequence(pgDB, "accounts")
	resetSequence(pgDB, "tabs")

	log.Println("Migration completed successfully!")
}

func parseTime(val string) time.Time {
	if val == "" {
		return time.Now()
	}
	// Try standard RFC3339 formats
	if t, err := time.Parse(time.RFC3339, val); err == nil {
		return t
	}
	if t, err := time.Parse(time.RFC3339Nano, val); err == nil {
		return t
	}
	// Try SQLite default datetime formats
	if t, err := time.Parse("2006-01-02 15:04:05", val); err == nil {
		return t
	}
	if t, err := time.Parse("2006-01-02T15:04:05Z", val); err == nil {
		return t
	}
	return time.Now()
}

func resetSequence(db *sql.DB, tableName string) {
	var maxID sql.NullInt64
	query := fmt.Sprintf("SELECT MAX(id) FROM %s", tableName)
	err := db.QueryRow(query).Scan(&maxID)
	if err != nil {
		log.Printf("Warning: Failed to get max id for resetting sequence on %s: %v", tableName, err)
		return
	}

	if maxID.Valid {
		seqName := fmt.Sprintf("%s_id_seq", tableName)
		_, err = db.Exec(fmt.Sprintf("SELECT setval('%s', %d)", seqName, maxID.Int64))
		if err != nil {
			log.Printf("Warning: Failed to reset sequence %s to %d: %v", seqName, maxID.Int64, err)
		} else {
			log.Printf("Reset sequence %s to %d", seqName, maxID.Int64)
		}
	}
}
