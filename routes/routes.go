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

	// 创建控制器实例
	brandHandler := handlers.NewBrandHandler()
	clientHandler := handlers.NewClientHandler()
	configHandler := handlers.NewConfigHandler()
	websiteHandler := handlers.NewWebsiteHandler()

	// API路由组
	api := r.Group("/api")
	{
		// 品牌相关路由
		brands := api.Group("/brands")
		{
			brands.GET("", brandHandler.GetBrands)
			brands.GET("/:id", brandHandler.GetBrand)
			brands.POST("", brandHandler.CreateBrand)
			brands.PUT("/:id", brandHandler.UpdateBrand)
			brands.DELETE("/:id", brandHandler.DeleteBrand)
		}

		// 客户端相关路由
		clients := api.Group("/clients")
		{
			clients.GET("", clientHandler.GetClients)
			clients.GET("/:id", clientHandler.GetClient)
			clients.POST("", clientHandler.CreateClient)
			clients.PUT("/:id", clientHandler.UpdateClient)
			clients.DELETE("/:id", clientHandler.DeleteClient)
		}

		// 基础配置路由
		baseConfigs := api.Group("/base-configs")
		{
			baseConfigs.GET("", configHandler.GetBaseConfigs)
			baseConfigs.GET("/:id", configHandler.GetBaseConfig)
			baseConfigs.POST("", configHandler.CreateBaseConfig)
			baseConfigs.PUT("/:id", configHandler.UpdateBaseConfig)
			baseConfigs.DELETE("/:id", configHandler.DeleteBaseConfig)
		}

		// 通用配置路由
		commonConfigs := api.Group("/common-configs")
		{
			commonConfigs.GET("", configHandler.GetCommonConfigs)
			commonConfigs.POST("", configHandler.CreateCommonConfig)
			commonConfigs.PUT("/:id", configHandler.UpdateCommonConfig)
			commonConfigs.DELETE("/:id", configHandler.DeleteCommonConfig)
		}

		// 支付配置路由
		payConfigs := api.Group("/pay-configs")
		{
			payConfigs.GET("", configHandler.GetPayConfigs)
			payConfigs.POST("", configHandler.CreatePayConfig)
			payConfigs.PUT("/:id", configHandler.UpdatePayConfig)
			payConfigs.DELETE("/:id", configHandler.DeletePayConfig)
		}

		// UI配置路由
		uiConfigs := api.Group("/ui-configs")
		{
			uiConfigs.GET("", configHandler.GetUIConfigs)
			uiConfigs.POST("", configHandler.CreateUIConfig)
			uiConfigs.PUT("/:id", configHandler.UpdateUIConfig)
			uiConfigs.DELETE("/:id", configHandler.DeleteUIConfig)
		}

		// 小说配置路由
		novelConfigs := api.Group("/novel-configs")
		{
			novelConfigs.GET("", configHandler.GetNovelConfigs)
			novelConfigs.POST("", configHandler.CreateNovelConfig)
			novelConfigs.PUT("/:id", configHandler.UpdateNovelConfig)
			novelConfigs.DELETE("/:id", configHandler.DeleteNovelConfig)
		}

		// 网站创建路由
		api.POST("/create-website", websiteHandler.CreateWebsite)

		// 网站配置查询路由
		api.GET("/website-config/:clientId", configHandler.GetWebsiteConfig)
	}

	return r
}
