package models

import (
	"time"
)

// EmailConfig 邮箱配置模型
type EmailConfig struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id" gorm:"not null;index"`
	Email     string    `json:"email" gorm:"not null;size:100"`
	Password  string    `json:"password" gorm:"not null;size:100"` // 授权码
	IsActive  bool      `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 指定表名
func (EmailConfig) TableName() string {
	return "email_configs"
}

// GetSMTPConfig 获取SMTP配置（写死的部分）
func (e *EmailConfig) GetSMTPConfig() map[string]interface{} {
	return map[string]interface{}{
		"host":     "smtp.exmail.qq.com", // QQ企业邮箱SMTP服务器
		"port":     465,                  // SSL端口
		"username": e.Email,              // 使用用户配置的邮箱
		"password": e.Password,           // 使用用户配置的授权码
		"ssl":      true,                 // 启用SSL
	}
}
