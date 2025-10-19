package main

import (
	"time"
)

// SQLite models from sqlite branch

type SqliteModel struct {
	ID        int       `gorm:"primaryKey" json:"-"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type SqliteBook struct {
	SqliteModel
	UUID      string       `json:"uuid" gorm:"uniqueIndex;type:text"`
	UserID    int          `json:"user_id" gorm:"index"`
	Label     string       `json:"label" gorm:"index"`
	Notes     []SqliteNote `json:"notes" gorm:"foreignKey:BookUUID;references:UUID"`
	AddedOn  int64 `json:"added_on"`
	EditedOn int64 `json:"edited_on"`
	USN      int   `json:"-" gorm:"index"`
	Deleted  bool  `json:"-" gorm:"default:false"`
}

func (SqliteBook) TableName() string {
	return "books"
}

type SqliteNote struct {
	SqliteModel
	UUID      string `json:"uuid" gorm:"index;type:text"`
	UserID    int    `json:"user_id" gorm:"index"`
	BookUUID  string `json:"book_uuid" gorm:"index;type:text"`
	Body      string `json:"content"`
	AddedOn   int64  `json:"added_on"`
	EditedOn  int64  `json:"edited_on"`
	Public  bool   `json:"public" gorm:"default:false"`
	USN     int    `json:"-" gorm:"index"`
	Deleted bool   `json:"-" gorm:"default:false"`
	Client  string `gorm:"index"`
}

func (SqliteNote) TableName() string {
	return "notes"
}

type SqliteUser struct {
	SqliteModel
	UUID        string         `json:"uuid" gorm:"type:text;index"`
	Account     SqliteAccount  `gorm:"foreignKey:UserID"`
	LastLoginAt *time.Time     `json:"-"`
	MaxUSN      int            `json:"-" gorm:"default:0"`
}

func (SqliteUser) TableName() string {
	return "users"
}

type SqliteAccount struct {
	SqliteModel
	UserID   int        `gorm:"index"`
	Email    NullString
	Password NullString
}

func (SqliteAccount) TableName() string {
	return "accounts"
}

type SqliteToken struct {
	SqliteModel
	UserID int        `gorm:"index"`
	Value  string     `gorm:"index"`
	Type   string
	UsedAt *time.Time
}

func (SqliteToken) TableName() string {
	return "tokens"
}

type SqliteSession struct {
	SqliteModel
	UserID     int       `gorm:"index"`
	Key        string    `gorm:"index"`
	LastUsedAt time.Time
	ExpiresAt  time.Time
}

func (SqliteSession) TableName() string {
	return "sessions"
}
