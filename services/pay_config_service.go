package services

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"brand-config-api/config"
	"brand-config-api/database"
	"brand-config-api/models"
	"brand-config-api/utils"
	"brand-config-api/utils/rollback"

	"gorm.io/gorm"
)

// PayConfigService 支付配置服务
type PayConfigService struct {
	db     *gorm.DB
	config *config.Config
}

// NewPayConfigService 创建支付配置服务实例
func NewPayConfigService() *PayConfigService {
	return &PayConfigService{
		db:     database.DB,
		config: config.Load(),
	}
}

// GetPayConfigs 获取所有支付配置
func (s *PayConfigService) GetPayConfigs() ([]models.PayConfig, error) {
	var configs []models.PayConfig
	err := s.db.Preload("Client.Brand").Find(&configs).Error
	return configs, err
}

// CreatePayConfig 创建支付配置
func (s *PayConfigService) CreatePayConfig(config *models.PayConfig) error {
	return s.db.Create(config).Error
}

// CreatePayConfigWithFile 创建支付配置并生成配置文件（使用外部事务）
func (s *PayConfigService) CreatePayConfigWithFile(ctx *rollback.TransactionContext, payConfigReq PayConfigRequest, clientID int, brandCode, host string) (*models.PayConfig, error) {
	// 1. 创建数据库记录
	payConfig := &models.PayConfig{
		ClientID:                clientID,
		NormalPayEnable:         payConfigReq.NormalPayEnable,
		NormalPayGatewayAndroid: payConfigReq.NormalPayGatewayAndroid,
		NormalPayGatewayIOS:     payConfigReq.NormalPayGatewayIOS,
		RenewPayEnable:          payConfigReq.RenewPayEnable,
		RenewPayGatewayAndroid:  payConfigReq.RenewPayGatewayAndroid,
		RenewPayGatewayIOS:      payConfigReq.RenewPayGatewayIOS,
	}

	if err := ctx.DB.Create(payConfig).Error; err != nil {
		return nil, fmt.Errorf("failed to create pay config in database: %v", err)
	}

	// 2. 生成配置文件
	if err := s.generateConfigFile(ctx, payConfig, brandCode, host); err != nil {
		return nil, fmt.Errorf("failed to generate pay config file: %v", err)
	}

	return payConfig, nil
}

// generateConfigFile 生成支付配置文件（不管理事务）
func (s *PayConfigService) generateConfigFile(ctx *rollback.TransactionContext, payConfig *models.PayConfig, brandCode, host string) error {
	configFile := filepath.Join(s.config.File.PayConfigsDir, brandCode+".js")
	hostConfig := s.FormatPayConfig(*payConfig)

	configFileUtils := utils.NewConfigFileUtils()
	if err := configFileUtils.GenerateConfigFile(ctx, configFile, hostConfig, host); err != nil {
		return fmt.Errorf("failed to generate config file: %v", err)
	}

	log.Printf("✅ 支付配置文件生成成功: brand=%s, host=%s", brandCode, host)
	return nil
}

// UpdatePayConfigByClientID 根据client_id更新支付配置
func (s *PayConfigService) UpdatePayConfigByClientID(clientID int, payConfig models.PayConfig) error {
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

		log.Printf("🔄 开始更新支付配置: brand=%s, host=%s", brand.Code, client.Host)

		// 检查是否已存在支付配置记录
		var existingPayConfig models.PayConfig
		err := ctx.DB.Where("client_id = ?", clientID).First(&existingPayConfig).Error

		if err != nil {
			if err == gorm.ErrRecordNotFound {
				// 记录不存在，创建新记录
				log.Printf("📝 支付配置记录不存在，创建新记录")
				payConfig.ClientID = clientID
				if err := ctx.DB.Create(&payConfig).Error; err != nil {
					return fmt.Errorf("failed to create pay config in database: %v", err)
				}
				log.Printf("✅ 数据库记录创建成功")
			} else {
				return fmt.Errorf("failed to check existing pay config: %v", err)
			}
		} else {
			// 记录存在，更新记录
			log.Printf("📝 支付配置记录已存在，更新记录")
			log.Printf("🔍 更新前的配置: %+v", existingPayConfig)
			log.Printf("🔍 要更新的配置: %+v", payConfig)

			// 保留原有的 ID 和 created_at，设置其他字段
			payConfig.ID = existingPayConfig.ID
			payConfig.ClientID = clientID
			payConfig.CreatedAt = existingPayConfig.CreatedAt
			payConfig.UpdatedAt = time.Now()

			// 使用 Select 明确指定要更新的字段，这样可以更新零值（如 false）
			if err := ctx.DB.Model(&existingPayConfig).Select(
				"client_id",
				"normal_pay_enable",
				"normal_pay_gateway_android",
				"normal_pay_gateway_ios",
				"renew_pay_enable",
				"renew_pay_gateway_android",
				"renew_pay_gateway_ios",
				"updated_at",
			).Updates(&payConfig).Error; err != nil {
				return fmt.Errorf("failed to update pay config in database: %v", err)
			}
			log.Printf("✅ 数据库记录更新成功")
		}

		// 更新本地配置文件
		// 构建文件路径
		configFile := filepath.Join(s.config.File.PayConfigsDir, brand.Code+".js")

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
		hostConfig := s.FormatPayConfig(payConfig)
		configData[client.Host] = hostConfig

		// 写入文件
		if err := configfileManager.WriteConfigDataToFile(configData, configFile); err != nil {
			return fmt.Errorf("failed to write config file: %v", err)
		}

		log.Printf("✅ 支付配置更新成功: brand=%s, host=%s", brand.Code, client.Host)
		return nil
	}, nil)
}

// DeletePayConfigByClientID 根据client_id删除支付配置（独立事务）
func (s *PayConfigService) DeletePayConfigByClientID(clientID int) error {
	// 创建事务管理器
	rollbackManager := rollback.NewRollbackManager(s.db, s.config)

	return rollbackManager.ExecuteWithTransaction(func(ctx *rollback.TransactionContext) error {
		return s.deletePayConfigInternal(ctx, clientID)
	}, nil)
}

// deletePayConfigInternal 内部删除支付配置方法（不管理事务）
func (s *PayConfigService) deletePayConfigInternal(ctx *rollback.TransactionContext, clientID int) error {
	// 先获取客户端和品牌信息
	var client models.Client
	if err := ctx.DB.Preload("Brand").Where("id = ?", clientID).First(&client).Error; err != nil {
		return fmt.Errorf("failed to find client: %v", err)
	}
	brand := client.Brand

	// 删除数据库记录（如果存在）
	if err := ctx.DB.Where("client_id = ?", clientID).Delete(&models.PayConfig{}).Error; err != nil {
		return fmt.Errorf("failed to delete pay config from database: %v", err)
	}

	// 处理配置文件
	configFile := filepath.Join(s.config.File.PayConfigsDir, brand.Code+".js")

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

// FormatPayConfig 格式化支付配置
func (s *PayConfigService) FormatPayConfig(payConfig models.PayConfig) map[string]interface{} {
	return map[string]interface{}{
		"normal_pay": map[string]interface{}{
			"enable": payConfig.NormalPayEnable,
			"gateway_id": map[string]interface{}{
				"android": payConfig.NormalPayGatewayAndroid,
				"ios":     payConfig.NormalPayGatewayIOS,
			},
		},
		"renew_pay": map[string]interface{}{
			"enable": payConfig.RenewPayEnable,
			"gateway_id": map[string]interface{}{
				"android": payConfig.RenewPayGatewayAndroid,
				"ios":     payConfig.RenewPayGatewayIOS,
			},
		},
	}
}
