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

// UIConfigService UI配置服务
type UIConfigService struct {
	db     *gorm.DB
	config *config.Config
}

// NewUIConfigService 创建UI配置服务实例
func NewUIConfigService() *UIConfigService {
	return &UIConfigService{
		db:     database.DB,
		config: config.Load(),
	}
}

// GetUIConfigs 获取所有UI配置
func (s *UIConfigService) GetUIConfigs() ([]models.UIConfig, error) {
	var configs []models.UIConfig
	err := s.db.Preload("Client.Brand").Find(&configs).Error
	return configs, err
}

// CreateUIConfig 创建UI配置
func (s *UIConfigService) CreateUIConfig(config *models.UIConfig) error {
	return s.db.Create(config).Error
}

// CreateUIConfigWithFile 创建UI配置并生成配置文件（使用外部事务）
func (s *UIConfigService) CreateUIConfigWithFile(ctx *rollback.TransactionContext, uiConfigReq UIConfigRequest, clientID int, brandCode, host string) (*models.UIConfig, error) {
	// 1. 创建数据库记录
	uiConfig := &models.UIConfig{
		ClientID:      clientID,
		ThemeBgMain:   uiConfigReq.ThemeBgMain,
		ThemeBgSecond: uiConfigReq.ThemeBgSecond,
		ThemeTextMain: uiConfigReq.ThemeTextMain,
	}

	if err := ctx.DB.Create(uiConfig).Error; err != nil {
		return nil, fmt.Errorf("failed to create UI config in database: %v", err)
	}

	// 2. 生成配置文件
	if err := s.generateConfigFile(ctx, uiConfig, brandCode, host); err != nil {
		return nil, fmt.Errorf("failed to generate UI config file: %v", err)
	}

	return uiConfig, nil
}

// generateConfigFile 生成UI配置文件（不管理事务）
func (s *UIConfigService) generateConfigFile(ctx *rollback.TransactionContext, uiConfig *models.UIConfig, brandCode, host string) error {
	configFile := filepath.Join(s.config.File.UIConfigsDir, brandCode+".js")
	hostConfig := s.FormatUIConfig(*uiConfig)

	configFileUtils := utils.NewConfigFileUtils()
	if err := configFileUtils.GenerateConfigFile(ctx, configFile, hostConfig, host); err != nil {
		return fmt.Errorf("failed to generate config file: %v", err)
	}

	log.Printf("✅ UI配置文件生成成功: brand=%s, host=%s", brandCode, host)
	return nil
}

// UpdateUIConfigByClientID 根据client_id更新UI配置
func (s *UIConfigService) UpdateUIConfigByClientID(clientID int, uiConfig models.UIConfig) error {
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

		log.Printf("🔄 开始更新UI配置: brand=%s, host=%s", brand.Code, client.Host)

		// 检查是否已存在UI配置记录
		var existingUIConfig models.UIConfig
		err := ctx.DB.Where("client_id = ?", clientID).First(&existingUIConfig).Error

		if err != nil {
			if err == gorm.ErrRecordNotFound {
				// 记录不存在，创建新记录
				log.Printf("📝 UI配置记录不存在，创建新记录")
				uiConfig.ClientID = clientID
				if err := ctx.DB.Create(&uiConfig).Error; err != nil {
					return fmt.Errorf("failed to create ui config in database: %v", err)
				}
				log.Printf("✅ 数据库记录创建成功")
			} else {
				return fmt.Errorf("failed to check existing ui config: %v", err)
			}
		} else {
			// 记录存在，更新记录
			log.Printf("📝 UI配置记录已存在，更新记录")
			if err := ctx.DB.Model(&existingUIConfig).Updates(uiConfig).Error; err != nil {
				return fmt.Errorf("failed to update ui config in database: %v", err)
			}
			log.Printf("✅ 数据库记录更新成功")
		}

		// 更新本地配置文件
		// 构建文件路径
		configFile := filepath.Join(s.config.File.UIConfigsDir, brand.Code+".js")

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

		// 更新指定host的配置
		hostConfig := s.FormatUIConfig(uiConfig)
		configData[client.Host] = hostConfig

		// 写入文件
		if err := configfileManager.WriteConfigDataToFile(configData, configFile); err != nil {
			return fmt.Errorf("failed to write config file: %v", err)
		}

		log.Printf("✅ UI配置更新成功: brand=%s, host=%s", brand.Code, client.Host)
		return nil
	}, nil)
}

// DeleteUIConfigByClientID 根据client_id删除UI配置（独立事务）
func (s *UIConfigService) DeleteUIConfigByClientID(clientID int) error {
	// 创建事务管理器
	rollbackManager := rollback.NewRollbackManager(s.db, s.config)

	return rollbackManager.ExecuteWithTransaction(func(ctx *rollback.TransactionContext) error {
		return s.deleteUIConfigInternal(ctx, clientID)
	}, nil)
}

// deleteUIConfigInternal 内部删除UI配置方法（不管理事务）
func (s *UIConfigService) deleteUIConfigInternal(ctx *rollback.TransactionContext, clientID int) error {
	// 先获取客户端和品牌信息
	var client models.Client
	if err := ctx.DB.Preload("Brand").Where("id = ?", clientID).First(&client).Error; err != nil {
		return fmt.Errorf("failed to find client: %v", err)
	}
	brand := client.Brand

	// 删除数据库记录（如果存在）
	if err := ctx.DB.Where("client_id = ?", clientID).Delete(&models.UIConfig{}).Error; err != nil {
		return fmt.Errorf("failed to delete ui config from database: %v", err)
	}

	// 处理配置文件
	configFile := filepath.Join(s.config.File.UIConfigsDir, brand.Code+".js")

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
	if _, exists := configData[client.Host]; exists {
		delete(configData, client.Host)
	}

	// 写入更新后的配置文件
	if err := configfileManager.WriteConfigDataToFile(configData, configFile); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	return nil
}

// FormatUIConfig 格式化UI配置
func (s *UIConfigService) FormatUIConfig(uiConfig models.UIConfig) map[string]interface{} {
	result := map[string]interface{}{
		"bgStyle": map[string]interface{}{
			"main":   uiConfig.ThemeBgMain,
			"second": uiConfig.ThemeBgSecond,
		},
	}

	if uiConfig.ThemeTextMain != nil && *uiConfig.ThemeTextMain != "" {
		result["textColor"] = map[string]interface{}{
			"main": *uiConfig.ThemeTextMain,
		}
	}

	return result
}
