package services

import (
	"fmt"
	"log"
	"path/filepath"

	"brand-config-api/config"
	"brand-config-api/database"
	"brand-config-api/models"
	"brand-config-api/utils"
	"brand-config-api/utils/rollback"

	"gorm.io/gorm"
)

// CommonConfigService 通用配置服务
type CommonConfigService struct {
	db     *gorm.DB
	config *config.Config
}

// NewCommonConfigService 创建通用配置服务实例
func NewCommonConfigService() *CommonConfigService {
	return &CommonConfigService{
		db:     database.DB,
		config: config.Load(),
	}
}

// GetCommonConfigs 获取所有通用配置
func (s *CommonConfigService) GetCommonConfigs() ([]models.CommonConfig, error) {
	var configs []models.CommonConfig
	err := s.db.Preload("Client.Brand").Find(&configs).Error
	return configs, err
}

// CreateCommonConfig 创建通用配置
func (s *CommonConfigService) CreateCommonConfig(config *models.CommonConfig) error {
	return s.db.Create(config).Error
}

// CreateCommonConfigWithFile 创建通用配置并生成配置文件（使用外部事务）
func (s *CommonConfigService) CreateCommonConfigWithFile(ctx *rollback.TransactionContext, commonConfigReq CommonConfigRequest, clientID int, brandCode, host string) (*models.CommonConfig, error) {
	// 1. 创建数据库记录
	commonConfig := &models.CommonConfig{
		ClientID:                clientID,
		DeliverBusinessIDEnable: commonConfigReq.DeliverBusinessIDEnable,
		DeliverBusinessID:       commonConfigReq.DeliverBusinessID,
		DeliverSwitchIDEnable:   commonConfigReq.DeliverSwitchIDEnable,
		DeliverSwitchID:         commonConfigReq.DeliverSwitchID,
		ProtocolCompany:         commonConfigReq.ProtocolCompany,
		ProtocolAbout:           commonConfigReq.ProtocolAbout,
		ProtocolPrivacy:         commonConfigReq.ProtocolPrivacy,
		ProtocolVod:             commonConfigReq.ProtocolVod,
		ProtocolUserCancel:      commonConfigReq.ProtocolUserCancel,
		ContactURL:              commonConfigReq.ContactURL,
		ScriptBase:              commonConfigReq.ScriptBase,
	}

	if err := ctx.DB.Create(commonConfig).Error; err != nil {
		return nil, fmt.Errorf("failed to create common config in database: %v", err)
	}

	// 2. 生成配置文件
	if err := s.generateConfigFile(ctx, commonConfig, brandCode, host); err != nil {
		return nil, fmt.Errorf("failed to generate common config file: %v", err)
	}

	return commonConfig, nil
}

// generateConfigFile 生成通用配置文件（不管理事务）
func (s *CommonConfigService) generateConfigFile(ctx *rollback.TransactionContext, commonConfig *models.CommonConfig, brandCode, host string) error {
	configFile := filepath.Join(s.config.File.CommonConfigsDir, brandCode+".js")
	hostConfig := s.FormatCommonConfig(*commonConfig)

	configFileUtils := utils.NewConfigFileUtils()
	if err := configFileUtils.GenerateConfigFile(ctx, configFile, hostConfig, host); err != nil {
		return fmt.Errorf("failed to generate config file: %v", err)
	}

	log.Printf("✅ 通用配置文件生成成功: brand=%s, host=%s", brandCode, host)
	return nil
}

// DeleteCommonConfigByClientID 根据client_id删除通用配置（独立事务）
func (s *CommonConfigService) DeleteCommonConfigByClientID(clientID int) error {
	// 创建事务管理器
	rollbackManager := rollback.NewRollbackManager(s.db, s.config)

	return rollbackManager.ExecuteWithTransaction(func(ctx *rollback.TransactionContext) error {
		return s.deleteCommonConfigInternal(ctx, clientID)
	}, nil)
}

// deleteCommonConfigInternal 内部删除通用配置方法（不管理事务）
func (s *CommonConfigService) deleteCommonConfigInternal(ctx *rollback.TransactionContext, clientID int) error {
	// 先获取客户端和品牌信息
	var client models.Client
	if err := ctx.DB.Preload("Brand").Where("id = ?", clientID).First(&client).Error; err != nil {
		return fmt.Errorf("failed to find client: %v", err)
	}
	brand := client.Brand

	// 删除数据库记录（如果存在）
	if err := ctx.DB.Where("client_id = ?", clientID).Delete(&models.CommonConfig{}).Error; err != nil {
		return fmt.Errorf("failed to delete common config from database: %v", err)
	}

	// 处理配置文件
	configFile := filepath.Join(s.config.File.CommonConfigsDir, brand.Code+".js")

	configFileUtils := utils.NewConfigFileUtils()
	if err := configFileUtils.DeleteConfigFileHost(ctx, configFile, client.Host); err != nil {
		return fmt.Errorf("failed to delete config file host: %v", err)
	}

	return nil
}

// FormatCommonConfig 格式化通用配置
func (s *CommonConfigService) FormatCommonConfig(commonConfig models.CommonConfig) map[string]interface{} {
	return map[string]interface{}{
		"deliver": map[string]interface{}{
			"business_id": map[string]interface{}{
				"value":  commonConfig.DeliverBusinessID,
				"enable": commonConfig.DeliverBusinessIDEnable,
			},
			"switch_id": map[string]interface{}{
				"value":  commonConfig.DeliverSwitchID,
				"enable": commonConfig.DeliverSwitchIDEnable,
			},
		},
		"protocol": map[string]interface{}{
			"company":    commonConfig.ProtocolCompany,
			"about":      commonConfig.ProtocolAbout,
			"privacy":    commonConfig.ProtocolPrivacy,
			"vod":        commonConfig.ProtocolVod,
			"userCancel": commonConfig.ProtocolUserCancel,
		},
		"contact": commonConfig.ContactURL,
		"script": map[string]interface{}{
			"base": commonConfig.ScriptBase,
		},
	}
}

// UpdateCommonConfigByClientID 根据client_id更新通用配置
func (s *CommonConfigService) UpdateCommonConfigByClientID(clientID int, commonConfig models.CommonConfig) error {
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

		log.Printf("🔄 开始更新通用配置: brand=%s, host=%s", brand.Code, client.Host)

		// 验证必填字段
		if commonConfig.DeliverBusinessIDEnable && commonConfig.DeliverBusinessID == "" {
			return fmt.Errorf("deliver_business_id is required when deliver_business_id_enable is true")
		}
		if commonConfig.DeliverSwitchIDEnable && commonConfig.DeliverSwitchID == "" {
			return fmt.Errorf("deliver_switch_id is required when deliver_switch_id_enable is true")
		}

		// 检查是否已存在通用配置记录
		var existingCommonConfig models.CommonConfig
		err := ctx.DB.Where("client_id = ?", clientID).First(&existingCommonConfig).Error

		if err != nil {
			if err == gorm.ErrRecordNotFound {
				// 记录不存在，创建新记录
				log.Printf("📝 通用配置记录不存在，创建新记录")
				commonConfig.ClientID = clientID
				if err := ctx.DB.Create(&commonConfig).Error; err != nil {
					return fmt.Errorf("failed to create common config in database: %v", err)
				}
				log.Printf("✅ 数据库记录创建成功")
			} else {
				return fmt.Errorf("failed to check existing common config: %v", err)
			}
		} else {
			// 记录存在，更新记录
			log.Printf("📝 通用配置记录已存在，更新记录")
			if err := ctx.DB.Model(&existingCommonConfig).Updates(commonConfig).Error; err != nil {
				return fmt.Errorf("failed to update common config in database: %v", err)
			}
			log.Printf("✅ 数据库记录更新成功")
		}

		// 更新本地配置文件
		configFile := filepath.Join(s.config.File.CommonConfigsDir, brand.Code+".js")
		hostConfig := s.FormatCommonConfig(commonConfig)

		configFileUtils := utils.NewConfigFileUtils()
		if err := configFileUtils.UpdateConfigFileHost(ctx, configFile, hostConfig, client.Host); err != nil {
			return fmt.Errorf("failed to update config file host: %v", err)
		}

		log.Printf("✅ 通用配置更新成功: brand=%s, host=%s", brand.Code, client.Host)
		return nil
	}, nil)
}
