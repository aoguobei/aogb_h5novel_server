package models

import (
	"time"
)

// Client 客户端模型
type Client struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	BrandID   int       `json:"brand_id" gorm:"column:brand_id;not null"`          // 关联的品牌ID
	Host      string    `json:"host" gorm:"column:host;type:varchar(20);not null"` // 端类型：h5/tth5/ksh5
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at"`

	// 关联关系
	Brand         Brand          `json:"brand" gorm:"foreignKey:BrandID"`
	BaseConfigs   []BaseConfig   `json:"base_configs" gorm:"foreignKey:ClientID"`
	CommonConfigs []CommonConfig `json:"common_configs" gorm:"foreignKey:ClientID"`
	PayConfigs    []PayConfig    `json:"pay_configs" gorm:"foreignKey:ClientID"`
	UIConfigs     []UIConfig     `json:"ui_configs" gorm:"foreignKey:ClientID"`
	NovelConfigs  []NovelConfig  `json:"novel_configs" gorm:"foreignKey:ClientID"`
}
