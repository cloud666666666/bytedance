package model

import (
	"time"
)

// Cart 购物车数据模型
type Cart struct {
	CartID    string `gorm:"primaryKey;type:varchar(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"`
	UserID    string `gorm:"type:varchar(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;index"`
	Items     string `gorm:"type:text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"` // JSON格式存储的购物车项目
	Currency  string `gorm:"type:varchar(10) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"`
	CreatedAt time.Time
	UpdatedAt time.Time
}