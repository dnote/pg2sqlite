package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
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

	// Connect to SQLite
	sqliteDB, err := sql.Open("sqlite3", config.SqlitePath)
	if err != nil {
		return fmt.Errorf("opening SQLite: %w", err)
	}
	defer sqliteDB.Close()

	if err := sqliteDB.Ping(); err != nil {
		return fmt.Errorf("pinging SQLite: %w", err)
	}

	fmt.Println("Connected to SQLite")

	// Run migration
	return migrate(pgDB, sqliteDB)
}
