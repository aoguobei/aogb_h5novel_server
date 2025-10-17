package handlers

import (
	"fmt"
	"log"
	"os"
	"strings"

	"brand-config-api/models"
	"brand-config-api/services"
	"brand-config-api/utils"

	"github.com/gin-gonic/gin"
)

// EmailRequest 邮件请求结构（使用配置的默认发件人）
type EmailRequest struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"content"`
}

// EmailRequestAs 邮件请求结构（前端控制发件人）
type EmailRequestAs struct {
	FromName  string `json:"from_name"`
	FromEmail string `json:"from_email"`
	To        string `json:"to"`
	Subject   string `json:"subject"`
	Body      string `json:"content"`
}

// EmailRequestWithUserAuth 用户输入邮箱和授权码发送邮件请求结构
type EmailRequestWithUserAuth struct {
	UserEmail    string   `json:"user_email" binding:"required,email"` // 用户邮箱账户
	UserPassword string   `json:"user_password" binding:"required"`    // 用户授权码
	ToEmails     []string `json:"to_emails" binding:"required,min=1"`  // 收件人邮箱列表
	CcEmails     []string `json:"cc_emails"`                           // 抄送人邮箱列表（可选）
	Subject      string   `json:"subject" binding:"required"`          // 邮件主题
	Content      string   `json:"content" binding:"required"`          // 邮件内容
}

// EmailResponse 邮件响应结构
type EmailResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// EmailHandler 邮件处理器
type EmailHandler struct {
	emailService *services.EmailService
}

// NewEmailHandler 创建邮件处理器
func NewEmailHandler() *EmailHandler {
	return &EmailHandler{
		emailService: services.NewEmailService(),
	}
}

// SendEmailWithUserAuth 用户输入邮箱和授权码发送邮件
func (h *EmailHandler) SendEmailWithUserAuth(c *gin.Context) {
	var req EmailRequestWithUserAuth

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	// 创建用户邮箱配置
	userConfig := &models.EmailConfig{
		Email:    req.UserEmail,
		Password: req.UserPassword,
	}

	// 构建邮件内容，包含用户信息
	enrichedContent := h.buildEnrichedContent(req)

	// 发送邮件（支持多个收件人和抄送人）
	err := h.emailService.SendEmailWithCC(userConfig, req.ToEmails, req.CcEmails, req.Subject, enrichedContent)
	if err != nil {
		log.Printf("❌ 发送邮件失败: %v", err)
		utils.InternalServerError(c, "发送失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{
		"from":    req.UserEmail,
		"to":      req.ToEmails,
		"cc":      req.CcEmails,
		"subject": req.Subject,
	}, "邮件发送成功！")
}

// SendEmailHandler 发送邮件处理器（使用配置的默认发件人）
func (h *EmailHandler) SendEmailHandler(c *gin.Context) {
	// 记录请求信息
	log.Printf("📧 收到邮件发送请求: %s %s", c.Request.Method, c.Request.URL.Path)

	// 解析请求体
	var req EmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("❌ JSON解析失败: %v", err)
		utils.BadRequest(c, "请求体解析失败: "+err.Error())
		return
	}

	log.Printf("📧 解析后的请求参数: To=%s, Subject=%s, Body长度=%d", req.To, req.Subject, len(req.Body))

	// 验证请求参数
	if req.To == "" || req.Subject == "" || req.Body == "" {
		log.Printf("❌ 参数验证失败: To='%s', Subject='%s', Body长度=%d", req.To, req.Subject, len(req.Body))
		utils.BadRequest(c, "收件人、主题和内容不能为空")
		return
	}

	// 获取邮件配置
	config := &models.EmailConfig{
		Email:    os.Getenv("SMTP_USERNAME"),
		Password: os.Getenv("SMTP_PASSWORD"),
	}

	// 将纯文本转换为HTML格式
	formattedBody := h.convertTextToHTML(req.Body)

	// 构建HTML邮件内容
	htmlBody := h.buildSimpleHTMLContent(req.Subject, formattedBody)

	// 发送邮件
	if err := h.emailService.SendEmail(config, req.To, req.Subject, htmlBody); err != nil {
		log.Printf("❌ 发送邮件失败: %v", err)
		utils.InternalServerError(c, fmt.Sprintf("发送邮件失败: %v", err))
		return
	}

	// 返回成功响应
	utils.Success(c, nil, "邮件发送成功")

	log.Printf("✅ 邮件发送成功: %s -> %s", req.Subject, req.To)
}

// SendEmailAsHandler 以指定发件人身份发送邮件处理器（前端控制发件人）
func (h *EmailHandler) SendEmailAsHandler(c *gin.Context) {
	// 解析请求体
	var req EmailRequestAs
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求体解析失败: "+err.Error())
		return
	}

	// 验证请求参数
	if req.FromName == "" || req.FromEmail == "" || req.To == "" || req.Subject == "" || req.Body == "" {
		utils.BadRequest(c, "发件人姓名、发件人邮箱、收件人、主题和内容不能为空")
		return
	}

	// 将纯文本转换为HTML格式
	formattedBody := h.convertTextToHTML(req.Body)

	// 构建HTML邮件内容
	htmlBody := h.buildSimpleHTMLContent(req.Subject, formattedBody)

	// 发送邮件
	if err := h.emailService.SendEmailAs(req.FromName, req.FromEmail, req.To, req.Subject, htmlBody); err != nil {
		log.Printf("❌ 发送邮件失败: %v", err)
		utils.InternalServerError(c, fmt.Sprintf("发送邮件失败: %v", err))
		return
	}

	// 返回成功响应
	utils.Success(c, nil, fmt.Sprintf("邮件发送成功 (发件人: %s <%s>)", req.FromName, req.FromEmail))

	log.Printf("✅ 邮件发送成功: %s -> %s (发件人: %s <%s>)", req.Subject, req.To, req.FromName, req.FromEmail)
}

// buildEnrichedContent 构建邮件内容（保持换行和缩进格式）
func (h *EmailHandler) buildEnrichedContent(req EmailRequestWithUserAuth) string {
	// 将纯文本转换为HTML格式
	formattedContent := h.convertTextToHTML(req.Content)

	// 构建HTML邮件内容，保持格式
	enrichedContent := `<html><head><meta charset="UTF-8"><title>` + req.Subject + `</title><style>body { font-family: "Microsoft YaHei UI", "Microsoft YaHei", "PingFang SC", "Hiragino Sans GB", sans-serif; font-size: 10.5pt; line-height: 1.8; color: #333; } .container { max-width: 800px; margin: 0 auto; padding: 30px; } .content { background: white; padding: 30px; border: 1px solid #e9ecef; border-radius: 8px; } .content br { margin: 3px 0; }</style></head><body><div class="container"><div class="content">` + formattedContent + `</div></div></body></html>`

	return enrichedContent
}

// buildSimpleHTMLContent 构建简单的HTML邮件内容
func (h *EmailHandler) buildSimpleHTMLContent(subject, content string) string {
	htmlContent := `<html><head><meta charset="UTF-8"><title>` + subject + `</title><style>body { font-family: "Microsoft YaHei UI", "Microsoft YaHei", "PingFang SC", "Hiragino Sans GB", sans-serif; font-size: 10.5pt; line-height: 1.8; color: #333; } .container { max-width: 800px; margin: 0 auto; padding: 30px; } .content { background: white; padding: 30px; border: 1px solid #e9ecef; border-radius: 8px; } .content br { margin: 3px 0; }</style></head><body><div class="container"><div class="content">` + content + `</div></div></body></html>`

	return htmlContent
}

// convertTextToHTML 将纯文本转换为HTML格式，支持换行和基本格式
func (h *EmailHandler) convertTextToHTML(content string) string {
	// 如果内容已经包含HTML标签（如<br>），则直接返回
	if strings.Contains(content, "<") && strings.Contains(content, ">") {
		return content
	}
	// 将换行符转换为HTML换行标签，保持格式
	html := strings.ReplaceAll(content, "\n", "<br>")
	return html
}
