package main

import (
	"brand-config-api/config"
	"fmt"
)

func main() {
	// 加载配置
	cfg := config.Load()

	fmt.Println("=== 配置测试 ===")
	fmt.Printf("基础路径: %s\n", cfg.File.BasePath)
	fmt.Printf("项目根目录: %s\n", cfg.File.ProjectRoot)
	fmt.Printf("配置文件目录: %s\n", cfg.File.ConfigDir)
	fmt.Printf("预构建目录: %s\n", cfg.File.PrebuildDir)
	fmt.Printf("静态资源目录: %s\n", cfg.File.StaticDir)
	fmt.Printf("Host文件: %s\n", cfg.File.HostFile)
	fmt.Printf("Index文件: %s\n", cfg.File.IndexFile)
	fmt.Printf("Vite配置文件: %s\n", cfg.File.ViteConfigFile)
	fmt.Printf("Package文件: %s\n", cfg.File.PackageFile)
	fmt.Printf("Base配置目录: %s\n", cfg.File.BaseConfigsDir)
	fmt.Printf("Common配置目录: %s\n", cfg.File.CommonConfigsDir)
	fmt.Printf("Pay配置目录: %s\n", cfg.File.PayConfigsDir)
	fmt.Printf("UI配置目录: %s\n", cfg.File.UIConfigsDir)
	fmt.Printf("Local配置目录: %s\n", cfg.File.LocalConfigsDir)
	fmt.Printf("Novel配置文件: %s\n", cfg.File.NovelConfigFile)

	fmt.Println("\n=== 数据库配置 ===")
	fmt.Printf("数据库主机: %s\n", cfg.Database.Host)
	fmt.Printf("数据库端口: %s\n", cfg.Database.Port)
	fmt.Printf("数据库用户: %s\n", cfg.Database.User)
	fmt.Printf("数据库名称: %s\n", cfg.Database.Name)

	fmt.Println("\n=== 服务器配置 ===")
	fmt.Printf("服务器端口: %s\n", cfg.Server.Port)
	fmt.Printf("运行模式: %s\n", cfg.Server.Mode)
}
