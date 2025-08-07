package routes

import (
	"brand-config-api/handlers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// SetupRoutes 设置路由
func SetupRoutes() *gin.Engine {
	r := gin.Default()

	// CORS配置
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	r.Use(cors.New(config))

	// API路由组
	api := r.Group("/api")
	{
		// 品牌相关路由
		brands := api.Group("/brands")
		{
			brands.GET("", handlers.GetBrands)
			brands.GET("/:id", handlers.GetBrand)
			brands.POST("", handlers.CreateBrand)
			brands.PUT("/:id", handlers.UpdateBrand)
			brands.DELETE("/:id", handlers.DeleteBrand)
		}

		// 客户端相关路由
		clients := api.Group("/clients")
		{
			clients.GET("", handlers.GetClients)
			clients.GET("/:id", handlers.GetClient)
			clients.POST("", handlers.CreateClient)
			clients.PUT("/:id", handlers.UpdateClient)
			clients.DELETE("/:id", handlers.DeleteClient)
		}

		// 基础配置路由
		baseConfigs := api.Group("/base-configs")
		{
			baseConfigs.GET("", handlers.GetBaseConfigs)
			baseConfigs.GET("/:id", handlers.GetBaseConfig)
			baseConfigs.POST("", handlers.CreateBaseConfig)
			baseConfigs.PUT("/:id", handlers.UpdateBaseConfig)
			baseConfigs.DELETE("/:id", handlers.DeleteBaseConfig)
		}

		// 通用配置路由
		commonConfigs := api.Group("/common-configs")
		{
			commonConfigs.GET("", handlers.GetCommonConfigs)
			commonConfigs.POST("", handlers.CreateCommonConfig)
			commonConfigs.PUT("/:id", handlers.UpdateCommonConfig)
			commonConfigs.DELETE("/:id", handlers.DeleteCommonConfig)
		}

		// 支付配置路由
		payConfigs := api.Group("/pay-configs")
		{
			payConfigs.GET("", handlers.GetPayConfigs)
			payConfigs.POST("", handlers.CreatePayConfig)
			payConfigs.PUT("/:id", handlers.UpdatePayConfig)
			payConfigs.DELETE("/:id", handlers.DeletePayConfig)
		}

		// UI配置路由
		uiConfigs := api.Group("/ui-configs")
		{
			uiConfigs.GET("", handlers.GetUIConfigs)
			uiConfigs.POST("", handlers.CreateUIConfig)
			uiConfigs.PUT("/:id", handlers.UpdateUIConfig)
			uiConfigs.DELETE("/:id", handlers.DeleteUIConfig)
		}

		// 网站创建路由
		api.POST("/create-website", handlers.CreateWebsite)
	}

	return r
}
