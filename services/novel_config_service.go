package services

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"brand-config-api/config"
	"brand-config-api/database"
	"brand-config-api/models"
	"brand-config-api/utils"
	"brand-config-api/utils/rollback"

	"gorm.io/gorm"
)

// NovelConfigService 小说配置服务
type NovelConfigService struct {
	db     *gorm.DB
	config *config.Config
}

// NewNovelConfigService 创建小说配置服务实例
func NewNovelConfigService() *NovelConfigService {
	return &NovelConfigService{
		db:     database.DB,
		config: config.Load(),
	}
}

// GetNovelConfigs 获取所有小说配置
func (s *NovelConfigService) GetNovelConfigs() ([]models.NovelConfig, error) {
	var configs []models.NovelConfig
	err := s.db.Preload("Client.Brand").Find(&configs).Error
	return configs, err
}

// CreateNovelConfig 创建小说配置
func (s *NovelConfigService) CreateNovelConfig(config *models.NovelConfig) error {
	return s.db.Create(config).Error
}

// CreateNovelConfigWithFile 创建小说配置并生成配置文件（使用外部事务）
func (s *NovelConfigService) CreateNovelConfigWithFile(ctx *rollback.TransactionContext, novelConfigReq NovelConfigRequest, clientID int, brandCode, host string) (*models.NovelConfig, error) {
	// 1. 创建数据库记录
	novelConfig := &models.NovelConfig{
		ClientID:              clientID,
		TTJumpHomeUrl:         novelConfigReq.TTJumpHomeUrl,
		TTLoginCallbackDomain: novelConfigReq.TTLoginCallbackDomain,
	}

	if err := ctx.DB.Create(novelConfig).Error; err != nil {
		return nil, fmt.Errorf("failed to create novel config in database: %v", err)
	}

	// 2. 生成配置文件
	if err := s.generateConfigFile(ctx, novelConfig, brandCode, host); err != nil {
		return nil, fmt.Errorf("failed to generate novel config file: %v", err)
	}

	return novelConfig, nil
}

// generateConfigFile 生成小说配置文件（不管理事务）
func (s *NovelConfigService) generateConfigFile(ctx *rollback.TransactionContext, novelConfig *models.NovelConfig, brandCode, host string) error {
	configFile := filepath.Join(s.config.File.LocalConfigsDir, "novelConfig.js")

	// 检查文件是否存在，如果存在则备份
	if _, err := os.Stat(configFile); err == nil {
		// 文件存在，进行备份
		if err := ctx.Files.Backup(configFile, ""); err != nil {
			return fmt.Errorf("failed to backup file: %v", err)
		}
		log.Printf("📝 备份已存在的配置文件: %s", configFile)
	} else if os.IsNotExist(err) {
		// 文件不存在，标记为新创建文件
		if err := ctx.Files.Backup(configFile, ""); err != nil {
			return fmt.Errorf("failed to backup new file: %v", err)
		}
		log.Printf("📝 标记新创建的配置文件: %s", configFile)
	} else {
		// 其他错误
		return fmt.Errorf("failed to check file existence: %v", err)
	}

	// 读取现有配置文件或创建新的配置对象
	configfileManager := utils.NewConfigFileManager()
	configData, err := configfileManager.ReadConfigFile(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			// 文件不存在，创建新的配置对象
			configData = make(map[string]interface{})
			log.Printf("📄 配置文件不存在，创建新的配置: %s", configFile)
		} else {
			return fmt.Errorf("failed to read existing config file: %v", err)
		}
	}

	// 更新指定host的配置
	hostConfig := s.FormatNovelConfig(*novelConfig)

	// 确保品牌配置存在
	if configData[brandCode] == nil {
		configData[brandCode] = make(map[string]interface{})
	}

	// 更新指定host的配置
	configData[brandCode].(map[string]interface{})[host] = hostConfig

	// 写入文件
	if err := configfileManager.WriteConfigDataToFile(configData, configFile); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	log.Printf("✅ 小说配置文件生成成功: brand=%s, host=%s", brandCode, host)
	return nil
}

// UpdateNovelConfigByClientID 根据client_id更新小说配置
func (s *NovelConfigService) UpdateNovelConfigByClientID(clientID int, novelConfig models.NovelConfig) error {
	// 创建事务管理器
	rollbackManager := rollback.NewRollbackManager(s.db, s.config)

	return rollbackManager.ExecuteWithTransaction(func(ctx *rollback.TransactionContext) error {
		// 获取客户端信息
		var client models.Client
		if err := ctx.DB.Where("id = ?", clientID).First(&client).Error; err != nil {
			return fmt.Errorf("failed to find client: %v", err)
		}

		// 获取品牌信息
		var brand models.Brand
		if err := ctx.DB.Where("id = ?", client.BrandID).First(&brand).Error; err != nil {
			return fmt.Errorf("failed to find brand: %v", err)
		}

		// 添加调试信息
		log.Printf("🔍 调试信息: clientID=%d, client.BrandID=%d, brand.Code=%s, client.Host=%s",
			clientID, client.BrandID, brand.Code, client.Host)

		log.Printf("🔄 开始更新小说配置: brand=%s, host=%s", brand.Code, client.Host)

		// 检查是否已存在小说配置记录
		var existingNovelConfig models.NovelConfig
		err := ctx.DB.Where("client_id = ?", clientID).First(&existingNovelConfig).Error

		if err != nil {
			if err == gorm.ErrRecordNotFound {
				// 记录不存在，创建新记录
				log.Printf("📝 小说配置记录不存在，创建新记录")
				novelConfig.ClientID = clientID
				if err := ctx.DB.Create(&novelConfig).Error; err != nil {
					return fmt.Errorf("failed to create novel config in database: %v", err)
				}
				log.Printf("✅ 数据库记录创建成功")
			} else {
				return fmt.Errorf("failed to check existing novel config: %v", err)
			}
		} else {
			// 记录存在，更新记录
			log.Printf("📝 小说配置记录已存在，更新记录")
			if err := ctx.DB.Model(&existingNovelConfig).Updates(novelConfig).Error; err != nil {
				return fmt.Errorf("failed to update novel config in database: %v", err)
			}
			log.Printf("✅ 数据库记录更新成功")
		}

		// 更新本地配置文件
		// 构建文件路径
		configFile := filepath.Join(s.config.File.LocalConfigsDir, "novelConfig.js")
		log.Printf("📁 准备更新文件: %s", configFile)

		// 备份文件
		log.Printf("📋 开始备份文件...")
		if err := ctx.Files.Backup(configFile, ""); err != nil {
			return fmt.Errorf("failed to backup file: %v", err)
		}
		log.Printf("✅ 文件备份成功")

		// 读取现有配置文件
		log.Printf("📖 开始读取现有配置文件...")
		configfileManager := utils.NewConfigFileManager()
		configData, err := configfileManager.ReadConfigFile(configFile)
		if err != nil {
			return fmt.Errorf("failed to read existing config file: %v", err)
		}
		log.Printf("✅ 配置文件读取成功，当前内容: %+v", configData)

		// 更新指定host的配置
		// 注意：文件结构是 {brandCode: {host: config}}
		hostConfig := s.FormatNovelConfig(novelConfig)
		log.Printf("📝 准备更新的配置: %+v", hostConfig)

		// 确保品牌配置存在
		if configData[brand.Code] == nil {
			log.Printf("🆕 品牌 %s 配置不存在，创建新的品牌配置", brand.Code)
			configData[brand.Code] = make(map[string]interface{})
		}

		// 更新指定host的配置
		configData[brand.Code].(map[string]interface{})[client.Host] = hostConfig
		log.Printf("✅ 配置数据更新完成，更新后的结构: %+v", configData)

		// 写入文件
		log.Printf("💾 开始写入配置文件...")
		if err := configfileManager.WriteConfigDataToFile(configData, configFile); err != nil {
			return fmt.Errorf("failed to write config file: %v", err)
		}
		log.Printf("✅ 配置文件写入成功")

		log.Printf("✅ 小说配置更新成功: brand=%s, host=%s", brand.Code, client.Host)
		log.Printf("📁 文件路径: %s", configFile)
		log.Printf("📝 更新的配置: %+v", hostConfig)
		return nil
	}, nil)
}

// DeleteNovelConfigByClientID 根据client_id删除小说配置（独立事务）
func (s *NovelConfigService) DeleteNovelConfigByClientID(clientID int) error {
	// 创建事务管理器
	rollbackManager := rollback.NewRollbackManager(s.db, s.config)

	return rollbackManager.ExecuteWithTransaction(func(ctx *rollback.TransactionContext) error {
		return s.deleteNovelConfigInternal(ctx, clientID)
	}, nil)
}

// deleteNovelConfigInternal 内部删除小说配置方法（不管理事务）
func (s *NovelConfigService) deleteNovelConfigInternal(ctx *rollback.TransactionContext, clientID int) error {
	// 先获取客户端和品牌信息
	var client models.Client
	if err := ctx.DB.Preload("Brand").Where("id = ?", clientID).First(&client).Error; err != nil {
		return fmt.Errorf("failed to find client: %v", err)
	}
	brand := client.Brand

	// 删除数据库记录（如果存在）
	if err := ctx.DB.Where("client_id = ?", clientID).Delete(&models.NovelConfig{}).Error; err != nil {
		return fmt.Errorf("failed to delete novel config from database: %v", err)
	}

	// 处理配置文件
	configFile := filepath.Join(s.config.File.LocalConfigsDir, "novelConfig.js")

	// 检查文件是否存在
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return nil
	}

	// 备份文件
	if err := ctx.Files.Backup(configFile, ""); err != nil {
		return fmt.Errorf("failed to backup file: %v", err)
	}

	// 读取现有配置文件
	configfileManager := utils.NewConfigFileManager()
	configData, err := configfileManager.ReadConfigFile(configFile)
	if err != nil {
		return fmt.Errorf("failed to read existing config file: %v", err)
	}

	// 删除指定host的配置
	if brandConfig, exists := configData[brand.Code]; exists {
		if hostConfig, ok := brandConfig.(map[string]interface{}); ok {
			if _, hostExists := hostConfig[client.Host]; hostExists {
				delete(hostConfig, client.Host)

				// 如果品牌配置为空，删除整个品牌配置
				if len(hostConfig) == 0 {
					delete(configData, brand.Code)
				}
			}
		}
	}

	// 写入更新后的配置文件
	if err := configfileManager.WriteConfigDataToFile(configData, configFile); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	return nil
}

// FormatNovelConfig 格式化小说配置
func (s *NovelConfigService) FormatNovelConfig(novelConfig models.NovelConfig) map[string]interface{} {
	return map[string]interface{}{
		"tt_jump_home_url":         novelConfig.TTJumpHomeUrl,
		"tt_login_callback_domain": novelConfig.TTLoginCallbackDomain,
	}
}

// RemoveNovelConfigEntries 删除novelconfig.js中对应品牌的host配置
func (s *NovelConfigService) RemoveNovelConfigEntries(ctx *rollback.TransactionContext, brandCode, host string) error {
	novelConfigFile := filepath.Join(s.config.File.LocalConfigsDir, "novelConfig.js")

	// 检查文件是否存在
	if _, err := os.Stat(novelConfigFile); os.IsNotExist(err) {
		return nil
	}

	// 读取现有配置文件
	configfileManager := utils.NewConfigFileManager()
	configData, err := configfileManager.ReadConfigFile(novelConfigFile)
	if err != nil {
		return fmt.Errorf("failed to read novelconfig.js: %v", err)
	}

	// 检查是否存在该品牌的配置
	if brandConfig, exists := configData[brandCode]; !exists {
		return nil
	} else {
		// 检查品牌配置是否为map类型
		if hostConfig, ok := brandConfig.(map[string]interface{}); ok {
			// 删除指定host的配置
			if _, hostExists := hostConfig[host]; hostExists {
				delete(hostConfig, host)

				// 如果品牌配置为空，删除整个品牌配置
				if len(hostConfig) == 0 {
					delete(configData, brandCode)
				}
			}
		}
	}

	// 写入更新后的配置文件
	if err := configfileManager.WriteConfigDataToFile(configData, novelConfigFile); err != nil {
		return fmt.Errorf("failed to write novelconfig.js: %v", err)
	}

	return nil
}
