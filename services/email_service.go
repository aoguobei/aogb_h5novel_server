package services

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"os"
	"strings"

	"brand-config-api/models"
)

// EmailService 邮箱服务
type EmailService struct{}

// NewEmailService 创建邮箱服务实例
func NewEmailService() *EmailService {
	return &EmailService{}
}

// SendEmail 发送邮件
func (s *EmailService) SendEmail(config *models.EmailConfig, to, subject, body string) error {
	return s.SendEmailWithCC(config, to, nil, subject, body)
}

// SendEmailWithCC 发送邮件（支持抄送人）
func (s *EmailService) SendEmailWithCC(config *models.EmailConfig, to interface{}, cc []string, subject, body string) error {
	// 获取SMTP配置
	smtpConfig := config.GetSMTPConfig()

	// 构建认证信息
	auth := smtp.PlainAuth("", smtpConfig["username"].(string), smtpConfig["password"].(string), smtpConfig["host"].(string))

	// 处理收件人参数（支持字符串或字符串数组）
	var toEmails []string
	var toHeader string

	switch v := to.(type) {
	case string:
		toEmails = []string{v}
		toHeader = v
	case []string:
		toEmails = v
		toHeader = strings.Join(v, ", ")
	default:
		return fmt.Errorf("不支持的收件人类型")
	}

	// 构建邮件头
	headers := make(map[string]string)
	headers["From"] = smtpConfig["username"].(string)
	headers["To"] = toHeader
	if len(cc) > 0 {
		headers["Cc"] = strings.Join(cc, ", ")
	}
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"

	// 构建邮件内容
	var emailContent strings.Builder
	for key, value := range headers {
		emailContent.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
	}
	emailContent.WriteString("\r\n")
	emailContent.WriteString(body)

	// 准备所有收件人（主收件人 + 抄送人）
	allRecipients := make([]string, 0, len(toEmails)+len(cc))
	allRecipients = append(allRecipients, toEmails...)
	if len(cc) > 0 {
		allRecipients = append(allRecipients, cc...)
	}

	// 发送邮件
	addr := fmt.Sprintf("%s:%d", smtpConfig["host"].(string), smtpConfig["port"].(int))

	if smtpConfig["ssl"].(bool) {
		// 使用SSL连接
		return s.sendEmailSSL(addr, auth, smtpConfig["username"].(string), allRecipients, []byte(emailContent.String()))
	} else {
		// 使用普通连接
		return smtp.SendMail(addr, auth, smtpConfig["username"].(string), allRecipients, []byte(emailContent.String()))
	}
}

// sendEmailSSL 使用SSL发送邮件
func (s *EmailService) sendEmailSSL(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
	// 建立TLS连接
	conn, err := tls.Dial("tcp", addr, &tls.Config{InsecureSkipVerify: true})
	if err != nil {
		return fmt.Errorf("TLS连接失败: %v", err)
	}
	defer conn.Close()

	// 创建SMTP客户端
	client, err := smtp.NewClient(conn, strings.Split(addr, ":")[0])
	if err != nil {
		return fmt.Errorf("创建SMTP客户端失败: %v", err)
	}
	defer client.Close()

	// 认证
	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP认证失败: %v", err)
	}

	// 设置发件人
	if err = client.Mail(from); err != nil {
		return fmt.Errorf("设置发件人失败: %v", err)
	}

	// 设置收件人
	for _, recipient := range to {
		if err = client.Rcpt(recipient); err != nil {
			return fmt.Errorf("设置收件人失败: %v", err)
		}
	}

	// 发送邮件内容
	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("准备发送邮件内容失败: %v", err)
	}

	_, err = writer.Write(msg)
	if err != nil {
		return fmt.Errorf("写入邮件内容失败: %v", err)
	}

	err = writer.Close()
	if err != nil {
		return fmt.Errorf("关闭邮件内容写入器失败: %v", err)
	}

	return client.Quit()
}

// SendEmailAs 以指定发件人身份发送邮件
func (s *EmailService) SendEmailAs(fromName, fromEmail, to, subject, body string) error {
	// 从环境变量获取邮件配置
	config := &models.EmailConfig{
		Email:    os.Getenv("SMTP_USERNAME"),
		Password: os.Getenv("SMTP_PASSWORD"),
	}

	// 获取SMTP配置
	smtpConfig := config.GetSMTPConfig()

	// 构建认证信息
	auth := smtp.PlainAuth("", smtpConfig["username"].(string), smtpConfig["password"].(string), smtpConfig["host"].(string))

	// 构建邮件头
	headers := make(map[string]string)
	headers["From"] = fmt.Sprintf("%s <%s>", fromName, fromEmail)
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"

	// 构建邮件内容
	var emailContent strings.Builder
	for key, value := range headers {
		emailContent.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
	}
	emailContent.WriteString("\r\n")
	emailContent.WriteString(body)

	// 发送邮件
	addr := fmt.Sprintf("%s:%d", smtpConfig["host"].(string), smtpConfig["port"].(int))

	if smtpConfig["ssl"].(bool) {
		// 使用SSL连接
		return s.sendEmailSSL(addr, auth, fromEmail, []string{to}, []byte(emailContent.String()))
	} else {
		// 使用普通连接
		return smtp.SendMail(addr, auth, fromEmail, []string{to}, []byte(emailContent.String()))
	}
}
