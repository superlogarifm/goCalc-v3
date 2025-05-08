package models

import "time"

type User struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Login        string    `json:"login" gorm:"unique;not null"`
	PasswordHash string    `json:"-" gorm:"not null"` // Не отправляем хеш пароля клиенту
	CreatedAt    time.Time `json:"created_at"`
}
