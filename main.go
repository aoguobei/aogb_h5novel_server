package main

import (
	"fmt"
	"log"

	"brand-config-api/config"
	"brand-config-api/database"
	"brand-config-api/middleware"
	"brand-config-api/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg := config.Load()

	// 设置Gin模式
	gin.SetMode(cfg.Server.Mode)

	// 初始化数据库
	database.InitDB()

	// 设置路由
	r := routes.SetupRoutes()

	// 添加中间件
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())
	r.Use(middleware.ErrorHandler())
	r.Use(middleware.CORS())
	r.Use(middleware.RequestID())

	log.Printf("Server starting on port %s", cfg.Server.Port)
	if err := r.Run(":" + cfg.Server.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}

	// 启动服务器
	port := cfg.Server.Port
	fmt.Printf("🚀 服务启动成功！\n")
	fmt.Printf("📧 邮件发送API(用户认证): http://localhost%s/api/email/send-user\n", port)
	fmt.Printf("🌐 其他API: http://localhost%s/api/*\n", port)
	fmt.Printf("按 Ctrl+C 停止服务\n\n")

	log.Fatal(r.Run(port))
}
