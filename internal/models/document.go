package models

import "time"

type Document struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Filename  string    `json:"filename"`
	CreatedAt time.Time `json:"created_at"`
}
