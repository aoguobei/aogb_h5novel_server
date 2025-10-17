package routes

import (
	"brand-config-api/handlers"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupDatabaseRoutes 设置数据库相关路由
func SetupDatabaseRoutes(router *gin.Engine, db *gorm.DB) {
	databaseHandler := handlers.NewDatabaseHandler(db)

	// 数据库API组
	databaseGroup := router.Group("/api/database")
	{
		// 导出数据库
		databaseGroup.POST("/export", databaseHandler.ExportDatabase)
	}
}
