package config

// EmailConfig 邮件配置结构
type EmailConfig struct {
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	FromEmail    string
	FromName     string
	UseSSL       bool
	UseTLS       bool
}

// LoadEmailConfig 加载邮件配置
func LoadEmailConfig() *EmailConfig {
	// 直接写死邮箱配置 - 腾讯企业邮箱配置
	config := &EmailConfig{
		SMTPHost:     "smtp.exmail.qq.com", // 腾讯企业邮箱SMTP服务器
		SMTPPort:     465,                  // 腾讯企业邮箱SMTP端口（SSL）
		SMTPUsername: "aogb@example.com",        // 企业邮箱地址
		SMTPPassword: "your_auth_code",     // 企业邮箱授权码（不是登录密码）
		FromEmail:    "aogb@example.com",        // 发件人邮箱
		FromName:     "H5网站配置系统",             // 发件人名称
		UseSSL:       true,                 // 使用SSL（端口465）
		UseTLS:       false,                // 不使用TLS（使用SSL）
	}

	return config
}
