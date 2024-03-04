package entity

import "time"

// Пользователь
type User struct {
	ID        int
	Login     string
	Password  string
	CreatedAt time.Time
	DeletedAt time.Time
}
