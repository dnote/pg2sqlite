package main

import (
	"database/sql"
	"fmt"
	"time"
)

type MigrationStats struct {
	Users    int
	Accounts int
	Books    int
	Notes    int
	Tokens   int
	Sessions int
}

func migrate(pgDB, sqliteDB *sql.DB) error {
	// Start transaction
	tx, err := sqliteDB.Begin()
	if err != nil {
		return fmt.Errorf("starting transaction: %w", err)
	}
	defer tx.Rollback()

	// Create schema (simplified, assuming GORM already created tables)
	// In production, this would run the same migrations as the server

	var stats MigrationStats

	// Migrate users
	fmt.Println("Migrating users...")
	if err := migrateUsers(pgDB, tx, &stats); err != nil {
		return fmt.Errorf("migrating users: %w", err)
	}
	fmt.Printf("  Migrated %d users\n", stats.Users)

	// Migrate accounts
	fmt.Println("Migrating accounts...")
	if err := migrateAccounts(pgDB, tx, &stats); err != nil {
		return fmt.Errorf("migrating accounts: %w", err)
	}
	fmt.Printf("  Migrated %d accounts\n", stats.Accounts)

	// Migrate books
	fmt.Println("Migrating books...")
	if err := migrateBooks(pgDB, tx, &stats); err != nil {
		return fmt.Errorf("migrating books: %w", err)
	}
	fmt.Printf("  Migrated %d books\n", stats.Books)

	// Migrate tokens
	fmt.Println("Migrating tokens...")
	if err := migrateTokens(pgDB, tx, &stats); err != nil {
		return fmt.Errorf("migrating tokens: %w", err)
	}
	fmt.Printf("  Migrated %d tokens\n", stats.Tokens)

	// Migrate sessions
	fmt.Println("Migrating sessions...")
	if err := migrateSessions(pgDB, tx, &stats); err != nil {
		return fmt.Errorf("migrating sessions: %w", err)
	}
	fmt.Printf("  Migrated %d sessions\n", stats.Sessions)

	// Migrate notes (last so FTS triggers work)
	fmt.Println("Migrating notes...")
	if err := migrateNotes(pgDB, tx, &stats); err != nil {
		return fmt.Errorf("migrating notes: %w", err)
	}
	fmt.Printf("  Migrated %d notes\n", stats.Notes)

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}

	// Print summary
	fmt.Println("\nMigration Summary:")
	fmt.Printf("  Users:    %d\n", stats.Users)
	fmt.Printf("  Accounts: %d\n", stats.Accounts)
	fmt.Printf("  Books:    %d\n", stats.Books)
	fmt.Printf("  Notes:    %d\n", stats.Notes)
	fmt.Printf("  Tokens:   %d\n", stats.Tokens)
	fmt.Printf("  Sessions: %d\n", stats.Sessions)

	return nil
}

func migrateUsers(pgDB *sql.DB, tx *sql.Tx, stats *MigrationStats) error {
	rows, err := pgDB.Query(`
		SELECT id, created_at, updated_at, uuid, last_login_at, max_usn, cloud
		FROM users
		ORDER BY id
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	stmt, err := tx.Prepare(`
		INSERT INTO users (id, created_at, updated_at, uuid, last_login_at, max_usn)
		VALUES (?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for rows.Next() {
		var id, maxUSN int
		var createdAt, updatedAt time.Time
		var uuid string
		var lastLoginAt sql.NullTime
		var cloud bool // Read but ignore

		if err := rows.Scan(&id, &createdAt, &updatedAt, &uuid, &lastLoginAt, &maxUSN, &cloud); err != nil {
			return err
		}

		if _, err := stmt.Exec(id, createdAt, updatedAt, uuid, lastLoginAt, maxUSN); err != nil {
			return err
		}
		stats.Users++
	}

	return rows.Err()
}

func migrateAccounts(pgDB *sql.DB, tx *sql.Tx, stats *MigrationStats) error {
	rows, err := pgDB.Query(`
		SELECT id, created_at, updated_at, user_id, email, password
		FROM accounts
		ORDER BY id
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	stmt, err := tx.Prepare(`
		INSERT INTO accounts (id, created_at, updated_at, user_id, email, password)
		VALUES (?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for rows.Next() {
		var id, userID int
		var createdAt, updatedAt time.Time
		var email, password sql.NullString

		if err := rows.Scan(&id, &createdAt, &updatedAt, &userID, &email, &password); err != nil {
			return err
		}

		if _, err := stmt.Exec(id, createdAt, updatedAt, userID, email, password); err != nil {
			return err
		}
		stats.Accounts++
	}

	return rows.Err()
}

func migrateBooks(pgDB *sql.DB, tx *sql.Tx, stats *MigrationStats) error {
	rows, err := pgDB.Query(`
		SELECT id, created_at, updated_at, uuid, user_id, label, added_on, edited_on, usn, deleted, encrypted
		FROM books
		ORDER BY id
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	stmt, err := tx.Prepare(`
		INSERT INTO books (id, created_at, updated_at, uuid, user_id, label, added_on, edited_on, usn, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for rows.Next() {
		var id, userID, usn int
		var addedOn, editedOn int64
		var createdAt, updatedAt time.Time
		var uuid, label string
		var deleted, encrypted bool

		// Read encrypted from Postgres but don't write it to SQLite
		if err := rows.Scan(&id, &createdAt, &updatedAt, &uuid, &userID, &label, &addedOn, &editedOn, &usn, &deleted, &encrypted); err != nil {
			return err
		}

		if _, err := stmt.Exec(id, createdAt, updatedAt, uuid, userID, label, addedOn, editedOn, usn, deleted); err != nil {
			return err
		}
		stats.Books++
	}

	return rows.Err()
}

func migrateNotes(pgDB *sql.DB, tx *sql.Tx, stats *MigrationStats) error {
	rows, err := pgDB.Query(`
		SELECT id, created_at, updated_at, uuid, user_id, book_uuid, body, added_on, edited_on, public, usn, deleted, encrypted, client
		FROM notes
		ORDER BY id
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	stmt, err := tx.Prepare(`
		INSERT INTO notes (id, created_at, updated_at, uuid, user_id, book_uuid, body, added_on, edited_on, public, usn, deleted, client)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for rows.Next() {
		var id, userID, usn int
		var addedOn, editedOn int64
		var createdAt, updatedAt time.Time
		var uuid, bookUUID, body, client string
		var public, deleted, encrypted bool

		// Read encrypted from Postgres but don't write it to SQLite
		if err := rows.Scan(&id, &createdAt, &updatedAt, &uuid, &userID, &bookUUID, &body, &addedOn, &editedOn, &public, &usn, &deleted, &encrypted, &client); err != nil {
			return err
		}

		if _, err := stmt.Exec(id, createdAt, updatedAt, uuid, userID, bookUUID, body, addedOn, editedOn, public, usn, deleted, client); err != nil {
			return err
		}
		stats.Notes++
	}

	return rows.Err()
}

func migrateTokens(pgDB *sql.DB, tx *sql.Tx, stats *MigrationStats) error {
	rows, err := pgDB.Query(`
		SELECT id, created_at, updated_at, user_id, value, type, used_at
		FROM tokens
		ORDER BY id
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	stmt, err := tx.Prepare(`
		INSERT INTO tokens (id, created_at, updated_at, user_id, value, type, used_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for rows.Next() {
		var id, userID int
		var createdAt, updatedAt time.Time
		var value, tokenType string
		var usedAt sql.NullTime

		if err := rows.Scan(&id, &createdAt, &updatedAt, &userID, &value, &tokenType, &usedAt); err != nil {
			return err
		}

		if _, err := stmt.Exec(id, createdAt, updatedAt, userID, value, tokenType, usedAt); err != nil {
			return err
		}
		stats.Tokens++
	}

	return rows.Err()
}

func migrateSessions(pgDB *sql.DB, tx *sql.Tx, stats *MigrationStats) error {
	rows, err := pgDB.Query(`
		SELECT id, created_at, updated_at, user_id, key, last_used_at, expires_at
		FROM sessions
		ORDER BY id
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	stmt, err := tx.Prepare(`
		INSERT INTO sessions (id, created_at, updated_at, user_id, key, last_used_at, expires_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for rows.Next() {
		var id, userID int
		var createdAt, updatedAt, lastUsedAt, expiresAt time.Time
		var key string

		if err := rows.Scan(&id, &createdAt, &updatedAt, &userID, &key, &lastUsedAt, &expiresAt); err != nil {
			return err
		}

		if _, err := stmt.Exec(id, createdAt, updatedAt, userID, key, lastUsedAt, expiresAt); err != nil {
			return err
		}
		stats.Sessions++
	}

	return rows.Err()
}
