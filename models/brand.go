package models

import (
	"time"
)

// Brand 品牌模型（带唯一code）
type Brand struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	Code      string    `json:"code" gorm:"column:code;type:varchar(100);uniqueIndex;not null"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at"`

	// 关联关系
	Clients []Client `json:"clients" gorm:"foreignKey:BrandID"`
}
