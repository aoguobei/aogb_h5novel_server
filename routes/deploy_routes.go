package routes

import (
	"brand-config-api/handlers"

	"github.com/gin-gonic/gin"
)

// SetupDeployRoutes 设置部署相关路由
func SetupDeployRoutes(router *gin.Engine) {
	deployHandler := handlers.NewDeployHandler()

	// 部署API路由组
	deploy := router.Group("/api/deploy")
	{
		deploy.POST("/nginx", deployHandler.DeployNginx) // 远程部署nginx配置
		deploy.POST("/local", deployHandler.DeployLocal) // 本地部署nginx配置
	}
}
