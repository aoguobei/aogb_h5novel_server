package services

import (
	"brand-config-api/database"
	"brand-config-api/models"

	"gorm.io/gorm"
)

// CommonConfigService 通用配置服务
type CommonConfigService struct {
	db *gorm.DB
}

// NewCommonConfigService 创建通用配置服务实例
func NewCommonConfigService() *CommonConfigService {
	return &CommonConfigService{
		db: database.DB,
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

// CreateCommonConfigFromRequestWithTx 从请求直接创建通用配置（在事务中）
func (s *CommonConfigService) CreateCommonConfigFromRequestWithTx(tx *gorm.DB, commonConfigReq CommonConfigRequest, clientID int) (*models.CommonConfig, error) {
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

	return commonConfig, tx.Create(commonConfig).Error
}

// UpdateCommonConfig 更新通用配置
func (s *CommonConfigService) UpdateCommonConfig(id int, config *models.CommonConfig) error {
	return s.db.Model(&models.CommonConfig{}).Where("id = ?", id).Updates(config).Error
}

// DeleteCommonConfig 删除通用配置
func (s *CommonConfigService) DeleteCommonConfig(id int) error {
	return s.db.Delete(&models.CommonConfig{}, id).Error
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
