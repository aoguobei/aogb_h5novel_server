package models

import (
	"time"
)

// BaseConfig 基础配置模型
type BaseConfig struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	ClientID  int       `json:"client_id" gorm:"not null;uniqueIndex"` // 关联的客户端ID，唯一约束
	Platform  string    `json:"platform" gorm:"type:varchar(20);not null"`
	AppName   string    `json:"app_name" gorm:"type:varchar(100);not null"`
	AppCode   string    `json:"app_code" gorm:"type:varchar(100);not null"`
	Product   string    `json:"product" gorm:"type:varchar(50);not null"`
	Customer  string    `json:"customer" gorm:"type:varchar(100);not null"`
	AppID     string    `json:"appid" gorm:"column:appid;type:varchar(100);default:''"`
	Version   string    `json:"version" gorm:"type:varchar(20);default:'1.0.0'"`
	CL        string    `json:"cl" gorm:"type:varchar(100);not null"`
	UC        string    `json:"uc" gorm:"type:varchar(100);default:''"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 关联关系
	Client Client `json:"client" gorm:"foreignKey:ClientID"`
}

// CommonConfig 通用配置模型
type CommonConfig struct {
	ID       int `json:"id" gorm:"primaryKey"`
	ClientID int `json:"client_id" gorm:"not null;uniqueIndex"` // 关联的客户端ID，唯一约束

	// Deliver 相关配置
	DeliverBusinessIDEnable bool   `json:"deliver_business_id_enable" gorm:"default:false"`
	DeliverBusinessID       string `json:"deliver_business_id" gorm:"type:varchar(100)"`
	DeliverSwitchIDEnable   bool   `json:"deliver_switch_id_enable" gorm:"default:false"`
	DeliverSwitchID         string `json:"deliver_switch_id" gorm:"type:varchar(100)"`

	// Protocol 相关配置
	ProtocolCompany    string `json:"protocol_company" gorm:"type:varchar(200)"`
	ProtocolAbout      string `json:"protocol_about" gorm:"type:text"`
	ProtocolPrivacy    string `json:"protocol_privacy" gorm:"type:text"`
	ProtocolVod        string `json:"protocol_vod" gorm:"type:text"`
	ProtocolUserCancel string `json:"protocol_user_cancel" gorm:"type:text"`

	// 其他配置
	ContactURL string `json:"contact_url" gorm:"type:text"`
	ScriptBase string `json:"script_base" gorm:"type:varchar(100)"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 关联关系
	Client Client `json:"client" gorm:"foreignKey:ClientID"`
}

// PayConfig 支付配置模型
type PayConfig struct {
	ID                      int       `json:"id" gorm:"primaryKey"`
	ClientID                int       `json:"client_id" gorm:"not null;uniqueIndex"` // 关联的客户端ID，唯一约束
	NormalPayEnable         bool      `json:"normal_pay_enable" gorm:"default:false"`
	NormalPayGatewayAndroid *int      `json:"normal_pay_gateway_android"`
	NormalPayGatewayIOS     *int      `json:"normal_pay_gateway_ios"`
	RenewPayEnable          bool      `json:"renew_pay_enable" gorm:"default:false"`
	RenewPayGatewayAndroid  *int      `json:"renew_pay_gateway_android"`
	RenewPayGatewayIOS      *int      `json:"renew_pay_gateway_ios"`
	CreatedAt               time.Time `json:"created_at"`
	UpdatedAt               time.Time `json:"updated_at"`

	// 关联关系
	Client Client `json:"client" gorm:"foreignKey:ClientID"`
}

// UIConfig UI配置模型
type UIConfig struct {
	ID            int       `json:"id" gorm:"primaryKey"`
	ClientID      int       `json:"client_id" gorm:"not null;uniqueIndex"` // 关联的客户端ID，唯一约束
	ThemeBgMain   string    `json:"theme_bg_main" gorm:"type:varchar(20);not null"`
	ThemeBgSecond string    `json:"theme_bg_second" gorm:"type:varchar(20);not null"`
	ThemeTextMain *string   `json:"theme_text_main" gorm:"type:varchar(20)"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	// 关联关系
	Client Client `json:"client" gorm:"foreignKey:ClientID"`
}

// NovelConfig 小说配置模型
type NovelConfig struct {
	ID                    int       `json:"id" gorm:"primaryKey"`
	ClientID              int       `json:"client_id" gorm:"not null;uniqueIndex"` // 关联的客户端ID，唯一约束
	TTJumpHomeUrl         string    `json:"tt_jump_home_url" gorm:"column:tt_jump_home_url;type:varchar(500)"`
	TTLoginCallbackDomain string    `json:"tt_login_callback_domain" gorm:"column:tt_login_callback_domain;type:varchar(200)"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`

	// 关联关系
	Client Client `json:"client" gorm:"foreignKey:ClientID"`
}
