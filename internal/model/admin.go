package model

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Admin struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Username     string    `gorm:"size:191;uniqueIndex;not null" json:"username"`
	PasswordHash string    `gorm:"column:password_hash;size:255;not null" json:"-"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

func (Admin) TableName() string {
	return "admins"
}

func (a *Admin) SetPassword(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	a.PasswordHash = string(hash)
	return nil
}

func (a *Admin) CheckPassword(password string) bool {
	if a.PasswordHash == "" {
		return false
	}

	return bcrypt.CompareHashAndPassword([]byte(a.PasswordHash), []byte(password)) == nil
}
