package model

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
)

func SeedAdmin(db *gorm.DB, username, password string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var admin Admin
		err := tx.Where("username = ?", username).First(&admin).Error
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("find admin: %w", err)
			}

			admin = Admin{
				Username: username,
			}
			if err := admin.SetPassword(password); err != nil {
				return fmt.Errorf("hash admin password: %w", err)
			}

			if err := tx.Create(&admin).Error; err != nil {
				return fmt.Errorf("create admin: %w", err)
			}

			return nil
		}

		if admin.CheckPassword(password) {
			return nil
		}

		if err := admin.SetPassword(password); err != nil {
			return fmt.Errorf("rehash admin password: %w", err)
		}

		if err := tx.Model(&admin).Update("password_hash", admin.PasswordHash).Error; err != nil {
			return fmt.Errorf("update admin password hash: %w", err)
		}

		return nil
	})
}
