package models

import "time"

// Note is the example model. Delete it (and its query interface, page and
// AutoMigrate entry) when you start a real project.
type Note struct {
	ID        uint   `gorm:"primaryKey"`
	Body      string `gorm:"not null"`
	CreatedAt time.Time
}
