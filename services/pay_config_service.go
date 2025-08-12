package services

import (
	"brand-config-api/database"
	"brand-config-api/models"

	"gorm.io/gorm"
)

// PayConfigService 支付配置服务
type PayConfigService struct {
	db *gorm.DB
}

// NewPayConfigService 创建支付配置服务实例
func NewPayConfigService() *PayConfigService {
	return &PayConfigService{
		db: database.DB,
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

// CreatePayConfigFromRequestWithTx 从请求直接创建支付配置（在事务中）
func (s *PayConfigService) CreatePayConfigFromRequestWithTx(tx *gorm.DB, payConfigReq PayConfigRequest, clientID int) (*models.PayConfig, error) {
	payConfig := &models.PayConfig{
		ClientID:                clientID,
		NormalPayEnable:         payConfigReq.NormalPayEnable,
		NormalPayGatewayAndroid: payConfigReq.NormalPayGatewayAndroid,
		NormalPayGatewayIOS:     payConfigReq.NormalPayGatewayIOS,
		RenewPayEnable:          payConfigReq.RenewPayEnable,
		RenewPayGatewayAndroid:  payConfigReq.RenewPayGatewayAndroid,
		RenewPayGatewayIOS:      payConfigReq.RenewPayGatewayIOS,
	}

	return payConfig, tx.Create(payConfig).Error
}

// UpdatePayConfig 更新支付配置
func (s *PayConfigService) UpdatePayConfig(id int, config *models.PayConfig) error {
	return s.db.Model(&models.PayConfig{}).Where("id = ?", id).Updates(config).Error
}

// DeletePayConfig 删除支付配置
func (s *PayConfigService) DeletePayConfig(id int) error {
	return s.db.Delete(&models.PayConfig{}, id).Error
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
