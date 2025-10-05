package main

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestMigration(t *testing.T) {
	// Setup Postgres test database
	pgHost := getEnvOrDefault("TEST_PG_HOST", "localhost")
	pgPort := getEnvOrDefault("TEST_PG_PORT", "5432")
	pgUser := getEnvOrDefault("TEST_PG_USER", "postgres")
	pgPassword := getEnvOrDefault("TEST_PG_PASSWORD", "")
	pgDB := getEnvOrDefault("TEST_PG_DB", "dnote_test")

	// Connect and setup schema
	pgDSN := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=disable",
		pgHost, pgPort, pgDB, pgUser, pgPassword)

	db, err := gorm.Open(postgres.Open(pgDSN), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to Postgres: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("Failed to get sql.DB: %v", err)
	}
	defer sqlDB.Close()

	// Drop and recreate schema
	if err := db.Exec("DROP SCHEMA public CASCADE; CREATE SCHEMA public;").Error; err != nil {
		t.Fatalf("Failed to reset schema: %v", err)
	}

	// Enable UUID extension
	if err := db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`).Error; err != nil {
		t.Fatalf("Failed to create uuid extension: %v", err)
	}

	// Create tables
	if err := db.AutoMigrate(
		&PgUser{},
		&PgAccount{},
		&PgBook{},
		&PgNote{},
		&PgToken{},
		&PgSession{},
	); err != nil {
		t.Fatalf("Failed to migrate schema: %v", err)
	}

	// Create test data
	now := time.Now()
	lastLogin := now.Add(-24 * time.Hour)

	user1 := PgUser{
		PgModel:     PgModel{CreatedAt: now, UpdatedAt: now},
		LastLoginAt: &lastLogin,
		MaxUSN:      10,
		Cloud:       true,
	}
	if err := db.Create(&user1).Error; err != nil {
		t.Fatalf("Failed to create user1: %v", err)
	}

	user2 := PgUser{
		PgModel:     PgModel{CreatedAt: now, UpdatedAt: now},
		LastLoginAt: nil,
		MaxUSN:      0,
		Cloud:       false,
	}
	if err := db.Create(&user2).Error; err != nil {
		t.Fatalf("Failed to create user2: %v", err)
	}

	account1 := PgAccount{
		PgModel:       PgModel{CreatedAt: now, UpdatedAt: now},
		UserID:        user1.ID,
		Email:         NullString{sql.NullString{String: "user1@example.com", Valid: true}},
		EmailVerified: true,
		Password:      NullString{sql.NullString{String: "hashedpassword1", Valid: true}},
	}
	if err := db.Create(&account1).Error; err != nil {
		t.Fatalf("Failed to create account1: %v", err)
	}

	account2 := PgAccount{
		PgModel:       PgModel{CreatedAt: now, UpdatedAt: now},
		UserID:        user2.ID,
		Email:         NullString{sql.NullString{String: "user2@example.com", Valid: true}},
		EmailVerified: false,
		Password:      NullString{sql.NullString{String: "hashedpassword2", Valid: true}},
	}
	if err := db.Create(&account2).Error; err != nil {
		t.Fatalf("Failed to create account2: %v", err)
	}

	book1 := PgBook{
		PgModel:  PgModel{CreatedAt: now, UpdatedAt: now},
		UserID:   user1.ID,
		Label:    "golang",
		AddedOn:  now.Unix(),
		EditedOn: now.Unix(),
		USN:      1,
		Deleted:  false,
	}
	if err := db.Create(&book1).Error; err != nil {
		t.Fatalf("Failed to create book1: %v", err)
	}

	book2 := PgBook{
		PgModel:  PgModel{CreatedAt: now, UpdatedAt: now},
		UserID:   user2.ID,
		Label:    "javascript",
		AddedOn:  now.Unix(),
		EditedOn: now.Unix(),
		USN:      1,
		Deleted:  false,
	}
	if err := db.Create(&book2).Error; err != nil {
		t.Fatalf("Failed to create book2: %v", err)
	}

	note1 := PgNote{
		PgModel:  PgModel{CreatedAt: now, UpdatedAt: now},
		UserID:   user1.ID,
		BookUUID: book1.UUID,
		Body:     "This is a test note about golang",
		AddedOn:  now.Unix(),
		EditedOn: now.Unix(),
		Public:   true,
		USN:      2,
		Client:   "cli",
	}
	if err := db.Create(&note1).Error; err != nil {
		t.Fatalf("Failed to create note1: %v", err)
	}

	note2 := PgNote{
		PgModel:  PgModel{CreatedAt: now, UpdatedAt: now},
		UserID:   user2.ID,
		BookUUID: book2.UUID,
		Body:     "JavaScript note",
		AddedOn:  now.Unix(),
		EditedOn: now.Unix(),
		Public:   false,
		USN:      2,
		Client:   "web",
	}
	if err := db.Create(&note2).Error; err != nil {
		t.Fatalf("Failed to create note2: %v", err)
	}

	token1 := PgToken{
		PgModel: PgModel{CreatedAt: now, UpdatedAt: now},
		UserID:  user1.ID,
		Value:   "token123",
		Type:    "access",
		UsedAt:  &now,
	}
	if err := db.Create(&token1).Error; err != nil {
		t.Fatalf("Failed to create token1: %v", err)
	}

	session1 := PgSession{
		PgModel:    PgModel{CreatedAt: now, UpdatedAt: now},
		UserID:     user1.ID,
		Key:        "session123",
		LastUsedAt: now,
		ExpiresAt:  now.Add(24 * time.Hour),
	}
	if err := db.Create(&session1).Error; err != nil {
		t.Fatalf("Failed to create session1: %v", err)
	}

	// Setup SQLite
	sqlitePath := "/tmp/dnote_test.db"
	os.Remove(sqlitePath) // Clean up any existing test db

	// Create SQLite schema using GORM AutoMigrate
	sqliteDB, err := gorm.Open(sqlite.Open(sqlitePath), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open SQLite: %v", err)
	}

	// AutoMigrate SQLite models
	if err := sqliteDB.AutoMigrate(
		&SqliteUser{},
		&SqliteAccount{},
		&SqliteBook{},
		&SqliteNote{},
		&SqliteToken{},
		&SqliteSession{},
	); err != nil {
		t.Fatalf("Failed to migrate SQLite schema: %v", err)
	}

	// Run migration
	config := Config{
		PgHost:     pgHost,
		PgPort:     pgPort,
		PgDatabase: pgDB,
		PgUser:     pgUser,
		PgPassword: pgPassword,
		SqlitePath: sqlitePath,
	}

	if err := run(config); err != nil {
		t.Fatalf("Migration failed: %v", err)
	}

	// Verify SQLite data
	sqliteRawDB, err := sql.Open("sqlite3", sqlitePath)
	if err != nil {
		t.Fatalf("Failed to open SQLite: %v", err)
	}
	defer sqliteRawDB.Close()

	// Verify user1
	var u1UUID string
	var u1LastLogin sql.NullTime
	var u1MaxUSN int
	if err := sqliteRawDB.QueryRow("SELECT uuid, last_login_at, max_usn FROM users WHERE id = ?", user1.ID).Scan(&u1UUID, &u1LastLogin, &u1MaxUSN); err != nil {
		t.Fatalf("Failed to query user1: %v", err)
	}
	if u1UUID != user1.UUID {
		t.Errorf("User1 UUID: expected %s, got %s", user1.UUID, u1UUID)
	}
	if !u1LastLogin.Valid {
		t.Errorf("User1 LastLoginAt should be valid")
	}
	if u1MaxUSN != user1.MaxUSN {
		t.Errorf("User1 MaxUSN: expected %d, got %d", user1.MaxUSN, u1MaxUSN)
	}

	// Verify user2
	var u2UUID string
	var u2LastLogin sql.NullTime
	var u2MaxUSN int
	if err := sqliteRawDB.QueryRow("SELECT uuid, last_login_at, max_usn FROM users WHERE id = ?", user2.ID).Scan(&u2UUID, &u2LastLogin, &u2MaxUSN); err != nil {
		t.Fatalf("Failed to query user2: %v", err)
	}
	if u2UUID != user2.UUID {
		t.Errorf("User2 UUID: expected %s, got %s", user2.UUID, u2UUID)
	}
	if u2LastLogin.Valid {
		t.Errorf("User2 LastLoginAt should be null")
	}
	if u2MaxUSN != user2.MaxUSN {
		t.Errorf("User2 MaxUSN: expected %d, got %d", user2.MaxUSN, u2MaxUSN)
	}

	// Verify account1
	var a1Email sql.NullString
	var a1EmailVerified bool
	var a1Password sql.NullString
	if err := sqliteRawDB.QueryRow("SELECT email, email_verified, password FROM accounts WHERE user_id = ?", user1.ID).Scan(&a1Email, &a1EmailVerified, &a1Password); err != nil {
		t.Fatalf("Failed to query account1: %v", err)
	}
	if a1Email.String != "user1@example.com" {
		t.Errorf("Account1 Email: expected 'user1@example.com', got '%s'", a1Email.String)
	}
	if !a1EmailVerified {
		t.Errorf("Account1 EmailVerified: expected true, got false")
	}
	if a1Password.String != "hashedpassword1" {
		t.Errorf("Account1 Password: expected 'hashedpassword1', got '%s'", a1Password.String)
	}

	// Verify account2
	var a2Email sql.NullString
	var a2EmailVerified bool
	var a2Password sql.NullString
	if err := sqliteRawDB.QueryRow("SELECT email, email_verified, password FROM accounts WHERE user_id = ?", user2.ID).Scan(&a2Email, &a2EmailVerified, &a2Password); err != nil {
		t.Fatalf("Failed to query account2: %v", err)
	}
	if a2Email.String != "user2@example.com" {
		t.Errorf("Account2 Email: expected 'user2@example.com', got '%s'", a2Email.String)
	}
	if a2EmailVerified {
		t.Errorf("Account2 EmailVerified: expected false, got true")
	}

	// Verify book1
	var b1UUID, b1Label string
	var b1AddedOn, b1EditedOn int64
	var b1USN int
	var b1Deleted, b1Encrypted bool
	if err := sqliteRawDB.QueryRow("SELECT uuid, label, added_on, edited_on, usn, deleted, encrypted FROM books WHERE user_id = ?", user1.ID).Scan(&b1UUID, &b1Label, &b1AddedOn, &b1EditedOn, &b1USN, &b1Deleted, &b1Encrypted); err != nil {
		t.Fatalf("Failed to query book1: %v", err)
	}
	if b1UUID != book1.UUID {
		t.Errorf("Book1 UUID: expected %s, got %s", book1.UUID, b1UUID)
	}
	if b1Label != "golang" {
		t.Errorf("Book1 Label: expected 'golang', got '%s'", b1Label)
	}
	if b1AddedOn != book1.AddedOn {
		t.Errorf("Book1 AddedOn: expected %d, got %d", book1.AddedOn, b1AddedOn)
	}
	if b1USN != book1.USN {
		t.Errorf("Book1 USN: expected %d, got %d", book1.USN, b1USN)
	}
	if b1Deleted != book1.Deleted {
		t.Errorf("Book1 Deleted: expected %v, got %v", book1.Deleted, b1Deleted)
	}

	// Verify note1
	var n1UUID, n1BookUUID, n1Body, n1Client string
	var n1AddedOn, n1EditedOn int64
	var n1Public, n1Deleted, n1Encrypted bool
	var n1USN int
	if err := sqliteRawDB.QueryRow("SELECT uuid, book_uuid, body, added_on, edited_on, public, usn, deleted, encrypted, client FROM notes WHERE user_id = ?", user1.ID).Scan(&n1UUID, &n1BookUUID, &n1Body, &n1AddedOn, &n1EditedOn, &n1Public, &n1USN, &n1Deleted, &n1Encrypted, &n1Client); err != nil {
		t.Fatalf("Failed to query note1: %v", err)
	}
	if n1UUID != note1.UUID {
		t.Errorf("Note1 UUID: expected %s, got %s", note1.UUID, n1UUID)
	}
	if n1BookUUID != book1.UUID {
		t.Errorf("Note1 BookUUID: expected %s, got %s", book1.UUID, n1BookUUID)
	}
	if n1Body != "This is a test note about golang" {
		t.Errorf("Note1 Body: expected 'This is a test note about golang', got '%s'", n1Body)
	}
	if !n1Public {
		t.Errorf("Note1 Public: expected true, got false")
	}
	if n1Client != "cli" {
		t.Errorf("Note1 Client: expected 'cli', got '%s'", n1Client)
	}

	// Verify token1
	var t1Value, t1Type string
	var t1UsedAt sql.NullTime
	if err := sqliteRawDB.QueryRow("SELECT value, type, used_at FROM tokens WHERE user_id = ?", user1.ID).Scan(&t1Value, &t1Type, &t1UsedAt); err != nil {
		t.Fatalf("Failed to query token1: %v", err)
	}
	if t1Value != "token123" {
		t.Errorf("Token1 Value: expected 'token123', got '%s'", t1Value)
	}
	if t1Type != "access" {
		t.Errorf("Token1 Type: expected 'access', got '%s'", t1Type)
	}
	if !t1UsedAt.Valid {
		t.Errorf("Token1 UsedAt should be valid")
	}

	// Verify session1
	var s1Key string
	var s1LastUsedAt, s1ExpiresAt time.Time
	if err := sqliteRawDB.QueryRow("SELECT key, last_used_at, expires_at FROM sessions WHERE user_id = ?", user1.ID).Scan(&s1Key, &s1LastUsedAt, &s1ExpiresAt); err != nil {
		t.Fatalf("Failed to query session1: %v", err)
	}
	if s1Key != "session123" {
		t.Errorf("Session1 Key: expected 'session123', got '%s'", s1Key)
	}

	// Clean up
	os.Remove(sqlitePath)
}

func getEnvOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
