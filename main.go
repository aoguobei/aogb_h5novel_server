package main

import (
	"log"
	"os"

	"brand-config-api/database"
	"brand-config-api/routes"
)

func main() {
	// 设置默认环境变量
	if os.Getenv("DB_HOST") == "" {
		os.Setenv("DB_HOST", "localhost")
	}
	if os.Getenv("DB_PORT") == "" {
		os.Setenv("DB_PORT", "3306")
	}
	if os.Getenv("DB_USER") == "" {
		os.Setenv("DB_USER", "root")
	}
	if os.Getenv("DB_PASSWORD") == "" {
		os.Setenv("DB_PASSWORD", "aoguobei-otzf")
	}
	if os.Getenv("DB_NAME") == "" {
		os.Setenv("DB_NAME", "h5novel_config")
	}

	// 初始化数据库
	database.InitDB()

	// 设置路由
	r := routes.SetupRoutes()

	// 启动服务器
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
