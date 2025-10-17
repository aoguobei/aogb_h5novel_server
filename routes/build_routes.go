package routes

import (
	"brand-config-api/handlers"
	"brand-config-api/utils"

	"github.com/gin-gonic/gin"
)

// SetupBuildRoutes 设置构建相关路由
func SetupBuildRoutes(router *gin.Engine, wsManager *utils.WebSocketManager, taskManager *utils.TaskManager) {
	buildHandler := handlers.NewBuildHandler(wsManager, taskManager)

	// 构建API路由组
	build := router.Group("/api/build")
	{
		build.POST("/h5", buildHandler.BuildH5)                // H5项目构建
		build.GET("/task/:taskId", buildHandler.GetTaskStatus) // 获取任务状态
	}
}
