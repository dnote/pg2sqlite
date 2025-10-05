package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Config struct {
	PgHost     string
	PgPort     string
	PgDatabase string
	PgUser     string
	PgPassword string
	SqlitePath string
}

func main() {
	var config Config

	flag.StringVar(&config.PgHost, "pg-host", "", "PostgreSQL host")
	flag.StringVar(&config.PgPort, "pg-port", "5432", "PostgreSQL port")
	flag.StringVar(&config.PgDatabase, "pg-database", "", "PostgreSQL database name")
	flag.StringVar(&config.PgUser, "pg-user", "", "PostgreSQL user")
	flag.StringVar(&config.PgPassword, "pg-password", "", "PostgreSQL password")
	flag.StringVar(&config.SqlitePath, "sqlite-path", "", "SQLite database path")
	flag.Parse()

	if err := validate(config); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		flag.Usage()
		os.Exit(1)
	}

	if err := run(config); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	fmt.Println("Migration completed successfully!")
}

func validate(c Config) error {
	if c.PgHost == "" {
		return fmt.Errorf("--pg-host is required")
	}
	if c.PgDatabase == "" {
		return fmt.Errorf("--pg-database is required")
	}
	if c.PgUser == "" {
		return fmt.Errorf("--pg-user is required")
	}
	if c.SqlitePath == "" {
		return fmt.Errorf("--sqlite-path is required")
	}
	return nil
}

func run(config Config) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(config.SqlitePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating database directory at %s: %w", dir, err)
	}

	// Check if SQLite file already exists
	if _, err := os.Stat(config.SqlitePath); err == nil {
		return fmt.Errorf("SQLite database already exists at %s - refusing to overwrite. Please remove the file or choose a different path", config.SqlitePath)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("checking if SQLite file exists: %w", err)
	}

	// Connect to PostgreSQL
	pgDSN := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=disable",
		config.PgHost, config.PgPort, config.PgDatabase, config.PgUser, config.PgPassword)

	pgDB, err := sql.Open("postgres", pgDSN)
	if err != nil {
		return fmt.Errorf("connecting to PostgreSQL: %w", err)
	}
	defer pgDB.Close()

	if err := pgDB.Ping(); err != nil {
		return fmt.Errorf("pinging PostgreSQL: %w", err)
	}

	fmt.Println("Connected to PostgreSQL")

	// Connect to SQLite with GORM
	sqliteDB, err := sql.Open("sqlite3", config.SqlitePath)
	if err != nil {
		return fmt.Errorf("opening SQLite: %w", err)
	}
	defer sqliteDB.Close()

	if err := sqliteDB.Ping(); err != nil {
		return fmt.Errorf("pinging SQLite: %w", err)
	}

	fmt.Println("Connected to SQLite")

	// Initialize SQLite schema using GORM
	fmt.Println("Creating SQLite schema...")
	if err := initSQLiteSchema(config.SqlitePath); err != nil {
		return fmt.Errorf("initializing SQLite schema: %w", err)
	}

	// Run migration
	return migrate(pgDB, sqliteDB)
}

func initSQLiteSchema(sqlitePath string) error {
	db, err := gorm.Open(sqlite.Open(sqlitePath), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("opening SQLite with GORM: %w", err)
	}

	// AutoMigrate SQLite models
	if err := db.AutoMigrate(
		&SqliteUser{},
		&SqliteAccount{},
		&SqliteBook{},
		&SqliteNote{},
		&SqliteToken{},
		&SqliteSession{},
	); err != nil {
		return fmt.Errorf("running AutoMigrate: %w", err)
	}

	return nil
}
