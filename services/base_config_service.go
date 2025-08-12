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

// BaseConfigService 基础配置服务
type BaseConfigService struct {
	db     *gorm.DB
	config *config.Config
}

// NewBaseConfigService 创建基础配置服务实例
func NewBaseConfigService() *BaseConfigService {
	return &BaseConfigService{
		db:     database.DB,
		config: config.Load(),
	}
}

// getConfigFilePath 获取配置文件路径
func (s *BaseConfigService) getConfigFilePath(brandCode string) string {
	return filepath.Join(s.config.File.BaseConfigsDir, brandCode+".js")
}

// GetBaseConfigByClientID 根据client_id获取基础配置
func (s *BaseConfigService) GetBaseConfigByClientID(clientID int) (map[string]interface{}, error) {
	// 通过client_id查询Client和Brand信息
	var client models.Client
	if err := database.DB.Preload("Brand").First(&client, clientID).Error; err != nil {
		return nil, fmt.Errorf("failed to find client with ID %d: %v", clientID, err)
	}

	// 获取brand_code和host
	brandCode := client.Brand.Code
	host := client.Host

	fmt.Printf("📋 Getting base config for client %d: brand=%s, host=%s\n", clientID, brandCode, host)

	// 直接实现获取配置的逻辑
	configFile := s.getConfigFilePath(brandCode)

	// 检查文件是否存在
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file does not exist: %s", configFile)
	}

	// 读取配置
	configfileManager := utils.NewConfigFileManager()
	configData, err := configfileManager.ReadConfigFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	// 获取指定host的配置
	hostConfig, exists := configData[host]
	if !exists {
		return nil, fmt.Errorf("host %s not found in config", host)
	}

	// 转换为map[string]interface{}
	if hostConfigMap, ok := hostConfig.(map[string]interface{}); ok {
		return hostConfigMap, nil
	}

	return nil, fmt.Errorf("invalid host config format")
}

// FormatBaseConfig 格式化基础配置
func (s *BaseConfigService) FormatBaseConfig(baseConfig models.BaseConfig) map[string]interface{} {
	return map[string]interface{}{
		"app_name": baseConfig.AppName,
		"platform": baseConfig.Platform,
		"app_code": baseConfig.AppCode,
		"product":  baseConfig.Product,
		"customer": baseConfig.Customer,
		"appid":    baseConfig.AppID,
		"version":  baseConfig.Version,
		"cl":       baseConfig.CL,
		"uc":       baseConfig.UC,
	}
}

// GetBaseConfigs 获取所有基础配置
func (s *BaseConfigService) GetBaseConfigs() ([]models.BaseConfig, error) {
	var configs []models.BaseConfig
	err := s.db.Preload("Client.Brand").Find(&configs).Error
	return configs, err
}

// GetBaseConfigByID 根据ID获取基础配置
func (s *BaseConfigService) GetBaseConfigByID(id int) (*models.BaseConfig, error) {
	var config models.BaseConfig
	err := s.db.Preload("Client.Brand").First(&config, id).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// CreateBaseConfig 创建基础配置
func (s *BaseConfigService) CreateBaseConfig(config *models.BaseConfig) error {
	return s.db.Create(config).Error
}

// CreateBaseConfigFromRequestWithTx 从请求直接创建基础配置（在事务中）
func (s *BaseConfigService) CreateBaseConfigFromRequestWithTx(tx *gorm.DB, baseConfigReq BaseConfigRequest, clientID int) (*models.BaseConfig, error) {
	baseConfig := &models.BaseConfig{
		ClientID: clientID,
		Platform: baseConfigReq.Platform,
		AppName:  baseConfigReq.AppName,
		AppCode:  baseConfigReq.AppCode,
		Product:  baseConfigReq.Product,
		Customer: baseConfigReq.Customer,
		AppID:    baseConfigReq.AppID,
		Version:  baseConfigReq.Version,
		CL:       baseConfigReq.CL,
		UC:       baseConfigReq.UC,
	}

	return baseConfig, tx.Create(baseConfig).Error
}

// UpdateBaseConfig 更新基础配置
func (s *BaseConfigService) UpdateBaseConfig(id int, config *models.BaseConfig) error {
	return s.db.Model(&models.BaseConfig{}).Where("id = ?", id).Updates(config).Error
}

// DeleteBaseConfig 删除基础配置
func (s *BaseConfigService) DeleteBaseConfig(id int) error {
	return s.db.Delete(&models.BaseConfig{}, id).Error
}

// BaseConfigFieldUpdate 基础配置字段更新结构
func (s *BaseConfigService) UpdateBaseConfigByClientID(clientID int, baseConfig models.BaseConfig) error {
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

		log.Printf("🔄 开始更新基础配置: brand=%s, host=%s", brand.Code, client.Host)

		// 内联验证和更新数据库逻辑
		if baseConfig.AppName == "" {
			return fmt.Errorf("app_name is required")
		}
		if baseConfig.Platform == "" {
			return fmt.Errorf("platform is required")
		}
		if baseConfig.AppCode == "" {
			return fmt.Errorf("app_code is required")
		}
		if baseConfig.Product == "" {
			return fmt.Errorf("product is required")
		}
		if baseConfig.Customer == "" {
			return fmt.Errorf("customer is required")
		}
		if baseConfig.CL == "" {
			return fmt.Errorf("cl is required")
		}

		// 更新数据库记录
		if err := ctx.DB.Model(&models.BaseConfig{}).Where("client_id = ?", clientID).Updates(baseConfig).Error; err != nil {
			return fmt.Errorf("failed to update base config in database: %v", err)
		}

		fmt.Printf("💾 数据库记录更新成功\n")

		// 更新本地配置文件
		// 构建文件路径
		configFile := filepath.Join(s.config.File.BaseConfigsDir, brand.Code+".js")

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
		hostConfig := s.FormatBaseConfig(baseConfig)

		configData[client.Host] = hostConfig

		// 写入文件
		if err := configfileManager.WriteConfigDataToFile(configData, configFile); err != nil {
			return fmt.Errorf("failed to write config file: %v", err)
		}

		log.Printf("📝 调用writeBaseConfigToFile...")
		log.Printf("📝 准备写入配置文件: %s", configFile)
		log.Printf("📖 读取现有配置文件...")

		log.Printf("✅ 基础配置更新成功: brand=%s, host=%s", brand.Code, client.Host)
		return nil
	})
}
