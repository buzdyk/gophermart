package model

import (
	"time"
)

type User struct {
	ID        int64      `json:"id" db:"id"`
	Login     string     `json:"login" db:"login"`
	Password  string     `json:"-" db:"password"` // Password is not exposed in JSON
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"-" db:"deleted_at"`
}

type UserCredentials struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}
