package models

import "time"

type Note struct {
	ID        int       `json:"id" db:"id"`
	Title     string    `json:"title" db:"title"`
	Content   string    `json:"content" db:"content"`
	CreatedAt time.Time `json:"created_at" db:"created_at"` // Можно использовать time.Time
	UserID    int       `json:"user_id" db:"user_id"`
}
