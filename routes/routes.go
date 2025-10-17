package routes

import (
	"brand-config-api/database"
	"brand-config-api/handlers"
	"brand-config-api/utils"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// 全局管理器实例
var (
	wsManager   *utils.WebSocketManager
	taskManager *utils.TaskManager
)

// SetupRoutes 设置路由
func SetupRoutes() *gin.Engine {
	r := gin.Default()

	// 初始化全局管理器
	wsManager = utils.NewWebSocketManager()
	taskManager = utils.NewTaskManager(10) // 最多10个并发任务

	// CORS配置
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	r.Use(cors.New(config))

	// 创建控制器实例
	typeHandler := handlers.NewTypeHandler()
	brandHandler := handlers.NewBrandHandler()
	clientHandler := handlers.NewClientHandler()
	websiteHandler := handlers.NewWebsiteHandler(wsManager, taskManager)
	baseConfigHandler := handlers.NewBaseConfigHandler()
	commonConfigHandler := handlers.NewCommonConfigHandler()
	payConfigHandler := handlers.NewPayConfigHandler()
	uiConfigHandler := handlers.NewUIConfigHandler()
	novelConfigHandler := handlers.NewNovelConfigHandler()
	websocketHandler := handlers.NewWebSocketHandler(wsManager, taskManager)

	// WebSocket路由
	r.GET("/ws", websocketHandler.HandleWebSocket)

	// API路由组
	api := r.Group("/api")
	{
		// 类型相关路由
		types := api.Group("/types")
		{
			types.GET("", typeHandler.GetTypes)
			types.GET("/:id", typeHandler.GetType)
			types.POST("", typeHandler.CreateType)
			types.PUT("/:id", typeHandler.UpdateType)
			types.DELETE("/:id", typeHandler.DeleteType)
		}

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
			baseConfigs.GET("", baseConfigHandler.GetBaseConfigs)
			// baseConfigs.GET("/:id", baseConfigHandler.GetBaseConfigByID)
			baseConfigs.POST("", baseConfigHandler.CreateBaseConfig)
			// baseConfigs.PUT("/:id", baseConfigHandler.UpdateBaseConfig)
			baseConfigs.PUT("/:clientId", baseConfigHandler.UpdateBaseConfigByClientID)
			baseConfigs.GET("/:clientId", baseConfigHandler.GetBaseConfigByClientID)
			baseConfigs.DELETE("/client/:clientId", baseConfigHandler.DeleteBaseConfigByClientID)
		}

		// 通用配置路由
		commonConfigs := api.Group("/common-configs")
		{
			commonConfigs.GET("", commonConfigHandler.GetCommonConfigs)
			commonConfigs.POST("", commonConfigHandler.CreateCommonConfig)
			// commonConfigs.PUT("/:id", commonConfigHandler.UpdateCommonConfig)
			commonConfigs.PUT("/:client_id", commonConfigHandler.UpdateCommonConfigByClientID)
			commonConfigs.DELETE("/client/:client_id", commonConfigHandler.DeleteCommonConfigByClientID)
		}

		// 支付配置路由
		payConfigs := api.Group("/pay-configs")
		{
			payConfigs.GET("", payConfigHandler.GetPayConfigs)
			payConfigs.POST("", payConfigHandler.CreatePayConfig)
			// payConfigs.PUT("/:id", payConfigHandler.UpdatePayConfig)
			payConfigs.PUT("/:client_id", payConfigHandler.UpdatePayConfigByClientID)
			payConfigs.DELETE("/client/:client_id", payConfigHandler.DeletePayConfigByClientID)
		}

		// UI配置路由
		uiConfigs := api.Group("/ui-configs")
		{
			uiConfigs.GET("", uiConfigHandler.GetUIConfigs)
			uiConfigs.POST("", uiConfigHandler.CreateUIConfig)
			// uiConfigs.PUT("/:id", uiConfigHandler.UpdateUIConfig)
			uiConfigs.PUT("/:client_id", uiConfigHandler.UpdateUIConfigByClientID)
			uiConfigs.DELETE("/client/:client_id", uiConfigHandler.DeleteUIConfigByClientID)
		}

		// 小说配置路由
		novelConfigs := api.Group("/novel-configs")
		{
			novelConfigs.GET("", novelConfigHandler.GetNovelConfigs)
			novelConfigs.POST("", novelConfigHandler.CreateNovelConfig)
			// novelConfigs.PUT("/:id", novelConfigHandler.UpdateNovelConfig)
			novelConfigs.PUT("/:client_id", novelConfigHandler.UpdateNovelConfigByClientID)
			novelConfigs.DELETE("/client/:client_id", novelConfigHandler.DeleteNovelConfigByClientID)
		}

		// 网站创建路由
		api.POST("/create-website", websiteHandler.CreateWebsite)

		// 网站配置查询路由
		api.GET("/website-config/:clientId", websiteHandler.GetWebsiteConfig)

		// 网站删除路由
		api.DELETE("/website/:clientId", websiteHandler.DeleteWebsite)
	}

	// 设置测试管理路由
	SetupTestRoutes(api)

	// 设置邮件路由
	SetupEmailRoutes(r)

	// 设置Git操作路由
	SetupGitRoutes(r)

	// 设置部署路由
	SetupDeployRoutes(r, wsManager, taskManager)

	// 设置构建路由
	SetupBuildRoutes(r, wsManager, taskManager)

	// 设置数据库路由
	SetupDatabaseRoutes(r, database.DB)

	return r
}
