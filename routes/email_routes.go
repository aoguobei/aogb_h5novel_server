package routes

import (
	"brand-config-api/handlers"

	"github.com/gin-gonic/gin"
)

// SetupEmailRoutes 设置邮件相关路由
func SetupEmailRoutes(router *gin.Engine) {
	emailHandler := handlers.NewEmailHandler()

	// 邮件相关路由组
	emailGroup := router.Group("/api/email")
	{
		// 用户输入邮箱和授权码发送邮件
		emailGroup.POST("/send-user", emailHandler.SendEmailWithUserAuth)
	}
}
