package main

import (
	"database/sql"
	"time"
)

// Old Postgres models from master branch

type NullString struct {
	sql.NullString
}

type PgModel struct {
	ID        int       `gorm:"primaryKey" json:"-"`
	CreatedAt time.Time `json:"created_at" gorm:"default:now()"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PgBook struct {
	PgModel
	UUID      string   `json:"uuid" gorm:"uniqueIndex;type:uuid;default:uuid_generate_v4()"`
	UserID    int      `json:"user_id" gorm:"index"`
	Label     string   `json:"label" gorm:"index"`
	Notes     []PgNote `json:"notes" gorm:"foreignKey:BookUUID;references:UUID"`
	AddedOn   int64    `json:"added_on"`
	EditedOn  int64    `json:"edited_on"`
	USN       int      `json:"-" gorm:"index"`
	Deleted   bool     `json:"-" gorm:"default:false"`
	Encrypted bool     `json:"-" gorm:"default:false"`
}

func (PgBook) TableName() string {
	return "books"
}

type PgNote struct {
	PgModel
	UUID      string `json:"uuid" gorm:"index;type:uuid;default:uuid_generate_v4()"`
	UserID    int    `json:"user_id" gorm:"index"`
	BookUUID  string `json:"book_uuid" gorm:"index;type:uuid"`
	Body      string `json:"content"`
	AddedOn   int64  `json:"added_on"`
	EditedOn  int64  `json:"edited_on"`
	TSV       string `json:"-" gorm:"type:tsvector"`
	Public    bool   `json:"public" gorm:"default:false"`
	USN       int    `json:"-" gorm:"index"`
	Deleted   bool   `json:"-" gorm:"default:false"`
	Encrypted bool   `json:"-" gorm:"default:false"`
	Client    string `gorm:"index"`
}

func (PgNote) TableName() string {
	return "notes"
}

type PgUser struct {
	PgModel
	UUID        string      `json:"uuid" gorm:"type:uuid;index;default:uuid_generate_v4()"`
	Account     PgAccount   `gorm:"foreignKey:UserID"`
	LastLoginAt *time.Time  `json:"-"`
	MaxUSN      int         `json:"-" gorm:"default:0"`
	Cloud       bool        `json:"-" gorm:"default:false"`
}

func (PgUser) TableName() string {
	return "users"
}

type PgAccount struct {
	PgModel
	UserID        int        `gorm:"index"`
	Email         NullString
	EmailVerified bool       `gorm:"default:false"`
	Password      NullString
}

func (PgAccount) TableName() string {
	return "accounts"
}

type PgToken struct {
	PgModel
	UserID int        `gorm:"index"`
	Value  string     `gorm:"index"`
	Type   string
	UsedAt *time.Time
}

func (PgToken) TableName() string {
	return "tokens"
}

type PgSession struct {
	PgModel
	UserID     int       `gorm:"index"`
	Key        string    `gorm:"index"`
	LastUsedAt time.Time
	ExpiresAt  time.Time
}

func (PgSession) TableName() string {
	return "sessions"
}
