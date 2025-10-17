package database

import (
	"fmt"
	"log"

	"brand-config-api/config"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDB 初始化数据库连接
func InitDB() {
	cfg := config.Load()

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name,
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	DB = db

	// 自动迁移数据库表（已禁用，使用 init.sql 手动建表）
	// err = DB.AutoMigrate(
	// 	&models.Brand{},
	// 	&models.Client{},
	// 	&models.BaseConfig{},
	// 	&models.CommonConfig{},
	// 	&models.PayConfig{},
	// 	&models.UIConfig{},
	// 	&models.NovelConfig{},
	// )

	// if err != nil {
	// 	log.Fatal("Failed to migrate database:", err)
	// }

	log.Println("Database connected and migrated successfully")
}
