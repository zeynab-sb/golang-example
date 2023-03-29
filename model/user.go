package model

import "time"

type User struct {
	ID        uint      `gorm:"Column:id"`
	UserName  string    `gorm:"Column:user_name"`
	Password  string    `gorm:"Column:password"`
	UpdatedAt time.Time `gorm:"Column:updated_at"`
	CreatedAt time.Time `gorm:"Column:created_at"`
}
