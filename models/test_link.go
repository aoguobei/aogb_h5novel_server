package models

import (
	"time"
)

// TestWebsite 测试网站模型
type TestWebsite struct {
	ID             uint       `json:"id" gorm:"primaryKey"`
	Name           string     `json:"name" gorm:"not null;size:255"`
	Type           string     `json:"type" gorm:"not null;size:50"`
	ScriptBase     string     `json:"script_base" gorm:"size:255;default:''"`
	ProdDomain     string     `json:"prod_domain" gorm:"size:255;default:''"`
	TestDomain     string     `json:"test_domain" gorm:"size:255;default:''"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	TestLinks      []TestLink `json:"test_links" gorm:"foreignKey:WebsiteID"` // 关联测试链接
	TestLinksCount int        `json:"test_links_count" gorm:"-"`              // 前端展示用的链接数量
}

// TableName 指定表名
func (TestWebsite) TableName() string {
	return "test_websites"
}

// TestLink 测试链接模型
type TestLink struct {
	ID        uint        `json:"id" gorm:"primaryKey"`
	WebsiteID uint        `json:"website_id" gorm:"not null"` // 关联测试网站ID，必填
	TestURL   string      `json:"test_url" gorm:"not null"`
	TestTitle string      `json:"test_title" gorm:"not null"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
	Website   TestWebsite `json:"website" gorm:"foreignKey:WebsiteID"` // 关联测试网站
}

// TableName 指定表名
func (TestLink) TableName() string {
	return "test_links"
}
