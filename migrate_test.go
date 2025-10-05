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

	// Run migration (will create schema automatically)
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

	// Open SQLite with GORM for verification
	sqliteDB, err := gorm.Open(sqlite.Open(sqlitePath), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open SQLite for verification: %v", err)
	}

	// Verify SQLite data using GORM
	// Verify user1
	var sqliteUser1 SqliteUser
	if err := sqliteDB.First(&sqliteUser1, user1.ID).Error; err != nil {
		t.Fatalf("Failed to query user1: %v", err)
	}
	if sqliteUser1.ID != user1.ID {
		t.Errorf("User1 ID: expected %d, got %d", user1.ID, sqliteUser1.ID)
	}
	if sqliteUser1.UUID != user1.UUID {
		t.Errorf("User1 UUID: expected %s, got %s", user1.UUID, sqliteUser1.UUID)
	}
	if sqliteUser1.CreatedAt.Unix() != user1.CreatedAt.Unix() {
		t.Errorf("User1 CreatedAt: expected %v, got %v", user1.CreatedAt, sqliteUser1.CreatedAt)
	}
	if sqliteUser1.UpdatedAt.Unix() != user1.UpdatedAt.Unix() {
		t.Errorf("User1 UpdatedAt: expected %v, got %v", user1.UpdatedAt, sqliteUser1.UpdatedAt)
	}
	if sqliteUser1.LastLoginAt == nil || user1.LastLoginAt == nil {
		if sqliteUser1.LastLoginAt != user1.LastLoginAt {
			t.Errorf("User1 LastLoginAt: one is nil, other is not")
		}
	} else if sqliteUser1.LastLoginAt.Unix() != user1.LastLoginAt.Unix() {
		t.Errorf("User1 LastLoginAt: expected %v, got %v", user1.LastLoginAt, sqliteUser1.LastLoginAt)
	}
	if sqliteUser1.MaxUSN != user1.MaxUSN {
		t.Errorf("User1 MaxUSN: expected %d, got %d", user1.MaxUSN, sqliteUser1.MaxUSN)
	}

	// Verify user2
	var sqliteUser2 SqliteUser
	if err := sqliteDB.First(&sqliteUser2, user2.ID).Error; err != nil {
		t.Fatalf("Failed to query user2: %v", err)
	}
	if sqliteUser2.ID != user2.ID {
		t.Errorf("User2 ID: expected %d, got %d", user2.ID, sqliteUser2.ID)
	}
	if sqliteUser2.UUID != user2.UUID {
		t.Errorf("User2 UUID: expected %s, got %s", user2.UUID, sqliteUser2.UUID)
	}
	if sqliteUser2.LastLoginAt != nil {
		t.Errorf("User2 LastLoginAt should be nil, got %v", sqliteUser2.LastLoginAt)
	}
	if sqliteUser2.MaxUSN != user2.MaxUSN {
		t.Errorf("User2 MaxUSN: expected %d, got %d", user2.MaxUSN, sqliteUser2.MaxUSN)
	}

	// Verify account1
	var sqliteAccount1 SqliteAccount
	if err := sqliteDB.Where("user_id = ?", user1.ID).First(&sqliteAccount1).Error; err != nil {
		t.Fatalf("Failed to query account1: %v", err)
	}
	if sqliteAccount1.ID != account1.ID {
		t.Errorf("Account1 ID: expected %d, got %d", account1.ID, sqliteAccount1.ID)
	}
	if sqliteAccount1.UserID != account1.UserID {
		t.Errorf("Account1 UserID: expected %d, got %d", account1.UserID, sqliteAccount1.UserID)
	}
	if sqliteAccount1.Email.String != account1.Email.String {
		t.Errorf("Account1 Email: expected %s, got %s", account1.Email.String, sqliteAccount1.Email.String)
	}
	if sqliteAccount1.EmailVerified != account1.EmailVerified {
		t.Errorf("Account1 EmailVerified: expected %v, got %v", account1.EmailVerified, sqliteAccount1.EmailVerified)
	}
	if sqliteAccount1.Password.String != account1.Password.String {
		t.Errorf("Account1 Password: expected %s, got %s", account1.Password.String, sqliteAccount1.Password.String)
	}
	if sqliteAccount1.CreatedAt.Unix() != account1.CreatedAt.Unix() {
		t.Errorf("Account1 CreatedAt: expected %v, got %v", account1.CreatedAt, sqliteAccount1.CreatedAt)
	}
	if sqliteAccount1.UpdatedAt.Unix() != account1.UpdatedAt.Unix() {
		t.Errorf("Account1 UpdatedAt: expected %v, got %v", account1.UpdatedAt, sqliteAccount1.UpdatedAt)
	}

	// Verify account2
	var sqliteAccount2 SqliteAccount
	if err := sqliteDB.Where("user_id = ?", user2.ID).First(&sqliteAccount2).Error; err != nil {
		t.Fatalf("Failed to query account2: %v", err)
	}
	if sqliteAccount2.Email.String != account2.Email.String {
		t.Errorf("Account2 Email: expected %s, got %s", account2.Email.String, sqliteAccount2.Email.String)
	}
	if sqliteAccount2.EmailVerified != account2.EmailVerified {
		t.Errorf("Account2 EmailVerified: expected %v, got %v", account2.EmailVerified, sqliteAccount2.EmailVerified)
	}

	// Verify book1
	var sqliteBook1 SqliteBook
	if err := sqliteDB.Where("user_id = ?", user1.ID).First(&sqliteBook1).Error; err != nil {
		t.Fatalf("Failed to query book1: %v", err)
	}
	if sqliteBook1.ID != book1.ID {
		t.Errorf("Book1 ID: expected %d, got %d", book1.ID, sqliteBook1.ID)
	}
	if sqliteBook1.UUID != book1.UUID {
		t.Errorf("Book1 UUID: expected %s, got %s", book1.UUID, sqliteBook1.UUID)
	}
	if sqliteBook1.UserID != book1.UserID {
		t.Errorf("Book1 UserID: expected %d, got %d", book1.UserID, sqliteBook1.UserID)
	}
	if sqliteBook1.Label != book1.Label {
		t.Errorf("Book1 Label: expected %s, got %s", book1.Label, sqliteBook1.Label)
	}
	if sqliteBook1.AddedOn != book1.AddedOn {
		t.Errorf("Book1 AddedOn: expected %d, got %d", book1.AddedOn, sqliteBook1.AddedOn)
	}
	if sqliteBook1.EditedOn != book1.EditedOn {
		t.Errorf("Book1 EditedOn: expected %d, got %d", book1.EditedOn, sqliteBook1.EditedOn)
	}
	if sqliteBook1.USN != book1.USN {
		t.Errorf("Book1 USN: expected %d, got %d", book1.USN, sqliteBook1.USN)
	}
	if sqliteBook1.Deleted != book1.Deleted {
		t.Errorf("Book1 Deleted: expected %v, got %v", book1.Deleted, sqliteBook1.Deleted)
	}
	if sqliteBook1.Encrypted != book1.Encrypted {
		t.Errorf("Book1 Encrypted: expected %v, got %v", book1.Encrypted, sqliteBook1.Encrypted)
	}
	if sqliteBook1.CreatedAt.Unix() != book1.CreatedAt.Unix() {
		t.Errorf("Book1 CreatedAt: expected %v, got %v", book1.CreatedAt, sqliteBook1.CreatedAt)
	}
	if sqliteBook1.UpdatedAt.Unix() != book1.UpdatedAt.Unix() {
		t.Errorf("Book1 UpdatedAt: expected %v, got %v", book1.UpdatedAt, sqliteBook1.UpdatedAt)
	}

	// Verify note1
	var sqliteNote1 SqliteNote
	if err := sqliteDB.Where("user_id = ?", user1.ID).First(&sqliteNote1).Error; err != nil {
		t.Fatalf("Failed to query note1: %v", err)
	}
	if sqliteNote1.ID != note1.ID {
		t.Errorf("Note1 ID: expected %d, got %d", note1.ID, sqliteNote1.ID)
	}
	if sqliteNote1.UUID != note1.UUID {
		t.Errorf("Note1 UUID: expected %s, got %s", note1.UUID, sqliteNote1.UUID)
	}
	if sqliteNote1.UserID != note1.UserID {
		t.Errorf("Note1 UserID: expected %d, got %d", note1.UserID, sqliteNote1.UserID)
	}
	if sqliteNote1.BookUUID != note1.BookUUID {
		t.Errorf("Note1 BookUUID: expected %s, got %s", note1.BookUUID, sqliteNote1.BookUUID)
	}
	if sqliteNote1.Body != note1.Body {
		t.Errorf("Note1 Body: expected %s, got %s", note1.Body, sqliteNote1.Body)
	}
	if sqliteNote1.AddedOn != note1.AddedOn {
		t.Errorf("Note1 AddedOn: expected %d, got %d", note1.AddedOn, sqliteNote1.AddedOn)
	}
	if sqliteNote1.EditedOn != note1.EditedOn {
		t.Errorf("Note1 EditedOn: expected %d, got %d", note1.EditedOn, sqliteNote1.EditedOn)
	}
	if sqliteNote1.Public != note1.Public {
		t.Errorf("Note1 Public: expected %v, got %v", note1.Public, sqliteNote1.Public)
	}
	if sqliteNote1.USN != note1.USN {
		t.Errorf("Note1 USN: expected %d, got %d", note1.USN, sqliteNote1.USN)
	}
	if sqliteNote1.Deleted != note1.Deleted {
		t.Errorf("Note1 Deleted: expected %v, got %v", note1.Deleted, sqliteNote1.Deleted)
	}
	if sqliteNote1.Encrypted != note1.Encrypted {
		t.Errorf("Note1 Encrypted: expected %v, got %v", note1.Encrypted, sqliteNote1.Encrypted)
	}
	if sqliteNote1.Client != note1.Client {
		t.Errorf("Note1 Client: expected %s, got %s", note1.Client, sqliteNote1.Client)
	}
	if sqliteNote1.CreatedAt.Unix() != note1.CreatedAt.Unix() {
		t.Errorf("Note1 CreatedAt: expected %v, got %v", note1.CreatedAt, sqliteNote1.CreatedAt)
	}
	if sqliteNote1.UpdatedAt.Unix() != note1.UpdatedAt.Unix() {
		t.Errorf("Note1 UpdatedAt: expected %v, got %v", note1.UpdatedAt, sqliteNote1.UpdatedAt)
	}

	// Verify token1
	var sqliteToken1 SqliteToken
	if err := sqliteDB.Where("user_id = ?", user1.ID).First(&sqliteToken1).Error; err != nil {
		t.Fatalf("Failed to query token1: %v", err)
	}
	if sqliteToken1.ID != token1.ID {
		t.Errorf("Token1 ID: expected %d, got %d", token1.ID, sqliteToken1.ID)
	}
	if sqliteToken1.UserID != token1.UserID {
		t.Errorf("Token1 UserID: expected %d, got %d", token1.UserID, sqliteToken1.UserID)
	}
	if sqliteToken1.Value != token1.Value {
		t.Errorf("Token1 Value: expected %s, got %s", token1.Value, sqliteToken1.Value)
	}
	if sqliteToken1.Type != token1.Type {
		t.Errorf("Token1 Type: expected %s, got %s", token1.Type, sqliteToken1.Type)
	}
	if sqliteToken1.UsedAt == nil || token1.UsedAt == nil {
		if sqliteToken1.UsedAt != token1.UsedAt {
			t.Errorf("Token1 UsedAt: one is nil, other is not")
		}
	} else if sqliteToken1.UsedAt.Unix() != token1.UsedAt.Unix() {
		t.Errorf("Token1 UsedAt: expected %v, got %v", token1.UsedAt, sqliteToken1.UsedAt)
	}
	if sqliteToken1.CreatedAt.Unix() != token1.CreatedAt.Unix() {
		t.Errorf("Token1 CreatedAt: expected %v, got %v", token1.CreatedAt, sqliteToken1.CreatedAt)
	}
	if sqliteToken1.UpdatedAt.Unix() != token1.UpdatedAt.Unix() {
		t.Errorf("Token1 UpdatedAt: expected %v, got %v", token1.UpdatedAt, sqliteToken1.UpdatedAt)
	}

	// Verify session1
	var sqliteSession1 SqliteSession
	if err := sqliteDB.Where("user_id = ?", user1.ID).First(&sqliteSession1).Error; err != nil {
		t.Fatalf("Failed to query session1: %v", err)
	}
	if sqliteSession1.ID != session1.ID {
		t.Errorf("Session1 ID: expected %d, got %d", session1.ID, sqliteSession1.ID)
	}
	if sqliteSession1.UserID != session1.UserID {
		t.Errorf("Session1 UserID: expected %d, got %d", session1.UserID, sqliteSession1.UserID)
	}
	if sqliteSession1.Key != session1.Key {
		t.Errorf("Session1 Key: expected %s, got %s", session1.Key, sqliteSession1.Key)
	}
	if sqliteSession1.LastUsedAt.Unix() != session1.LastUsedAt.Unix() {
		t.Errorf("Session1 LastUsedAt: expected %v, got %v", session1.LastUsedAt, sqliteSession1.LastUsedAt)
	}
	if sqliteSession1.ExpiresAt.Unix() != session1.ExpiresAt.Unix() {
		t.Errorf("Session1 ExpiresAt: expected %v, got %v", session1.ExpiresAt, sqliteSession1.ExpiresAt)
	}
	if sqliteSession1.CreatedAt.Unix() != session1.CreatedAt.Unix() {
		t.Errorf("Session1 CreatedAt: expected %v, got %v", session1.CreatedAt, sqliteSession1.CreatedAt)
	}
	if sqliteSession1.UpdatedAt.Unix() != session1.UpdatedAt.Unix() {
		t.Errorf("Session1 UpdatedAt: expected %v, got %v", session1.UpdatedAt, sqliteSession1.UpdatedAt)
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
