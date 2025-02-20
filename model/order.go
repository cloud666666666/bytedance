package model

import (
	"time"
)

type Order struct {
	OrderID    string `gorm:"primaryKey;type:varchar(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"`
	UserID     string `gorm:"type:varchar(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"`
	Address    string `gorm:"type:text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"`
	Email      string `gorm:"type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"`
	OrderItems string `gorm:"type:text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"`
	Currency   string `gorm:"type:varchar(10) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"`
	Status     string `gorm:"type:varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
	ExpireAt   time.Time
}
