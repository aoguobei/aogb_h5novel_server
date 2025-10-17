package routes

import (
	"brand-config-api/handlers"
	"brand-config-api/utils"

	"github.com/gin-gonic/gin"
)

// SetupDeployRoutes 设置部署相关路由
func SetupDeployRoutes(router *gin.Engine, wsManager *utils.WebSocketManager, taskManager *utils.TaskManager) {
	deployHandler := handlers.NewDeployHandler(wsManager, taskManager)

	// 部署API路由组
	deploy := router.Group("/api/deploy")
	{
		deploy.POST("/nginx", deployHandler.DeployNginx) // 远程部署nginx配置
	}
}
