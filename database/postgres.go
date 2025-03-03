package database

import "time"

type WikipediaTitle struct {
	ID        string `gorm:"primaryKey"`
	Title     string
	NumericID int
	CreatedAt time.Time
	UpdatedAt time.Time
}
