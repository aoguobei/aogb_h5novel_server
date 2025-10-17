package routes

import (
	"brand-config-api/handlers"

	"github.com/gin-gonic/gin"
)

// SetupTestRoutes 设置测试管理相关路由
func SetupTestRoutes(api *gin.RouterGroup) {
	// 创建处理器实例
	testWebsiteHandler := handlers.NewTestWebsiteHandler()
	testLinkHandler := handlers.NewTestLinkHandler()

	// 测试链接路由
	testLinks := api.Group("/test-links")
	{
		testLinks.POST("", testLinkHandler.CreateTestLink)
		testLinks.POST("/batch", testLinkHandler.BatchCreateTestLinks)
		testLinks.PUT("/:id", testLinkHandler.UpdateTestLink)
		testLinks.DELETE("/:id", testLinkHandler.DeleteTestLink)
	}

	// 测试网站路由
	testWebsites := api.Group("/test-websites")
	{
		testWebsites.GET("", testWebsiteHandler.GetTestWebsites)
		testWebsites.GET("/:id", testWebsiteHandler.GetTestWebsite)
		testWebsites.POST("", testWebsiteHandler.CreateTestWebsite)
		testWebsites.PUT("/:id", testWebsiteHandler.UpdateTestWebsite)
		testWebsites.DELETE("/:id", testWebsiteHandler.DeleteTestWebsite)
		testWebsites.GET("/:id/test-links", testWebsiteHandler.GetTestLinksByWebsiteID)
	}
}
