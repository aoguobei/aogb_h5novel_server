package services

import (
	"brand-config-api/database"
	"brand-config-api/models"

	"gorm.io/gorm"
)

// ConfigService 配置服务
type ConfigService struct {
	db *gorm.DB
}

// NewConfigService 创建配置服务实例
func NewConfigService() *ConfigService {
	return &ConfigService{
		db: database.DB,
	}
}

// GetBaseConfigs 获取所有基础配置
func (s *ConfigService) GetBaseConfigs() ([]models.BaseConfig, error) {
	var configs []models.BaseConfig
	err := s.db.Preload("Client.Brand").Find(&configs).Error
	return configs, err
}

// GetBaseConfigByID 根据ID获取基础配置
func (s *ConfigService) GetBaseConfigByID(id int) (*models.BaseConfig, error) {
	var config models.BaseConfig
	err := s.db.Preload("Client.Brand").First(&config, id).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// CreateBaseConfig 创建基础配置
func (s *ConfigService) CreateBaseConfig(config *models.BaseConfig) error {
	return s.db.Create(config).Error
}

// UpdateBaseConfig 更新基础配置
func (s *ConfigService) UpdateBaseConfig(id int, config *models.BaseConfig) error {
	return s.db.Model(&models.BaseConfig{}).Where("id = ?", id).Updates(config).Error
}

// DeleteBaseConfig 删除基础配置
func (s *ConfigService) DeleteBaseConfig(id int) error {
	return s.db.Delete(&models.BaseConfig{}, id).Error
}

// GetCommonConfigs 获取所有通用配置
func (s *ConfigService) GetCommonConfigs() ([]models.CommonConfig, error) {
	var configs []models.CommonConfig
	err := s.db.Preload("Client.Brand").Find(&configs).Error
	return configs, err
}

// CreateCommonConfig 创建通用配置
func (s *ConfigService) CreateCommonConfig(config *models.CommonConfig) error {
	return s.db.Create(config).Error
}

// UpdateCommonConfig 更新通用配置
func (s *ConfigService) UpdateCommonConfig(id int, config *models.CommonConfig) error {
	return s.db.Model(&models.CommonConfig{}).Where("id = ?", id).Updates(config).Error
}

// DeleteCommonConfig 删除通用配置
func (s *ConfigService) DeleteCommonConfig(id int) error {
	return s.db.Delete(&models.CommonConfig{}, id).Error
}

// GetPayConfigs 获取所有支付配置
func (s *ConfigService) GetPayConfigs() ([]models.PayConfig, error) {
	var configs []models.PayConfig
	err := s.db.Preload("Client.Brand").Find(&configs).Error
	return configs, err
}

// CreatePayConfig 创建支付配置
func (s *ConfigService) CreatePayConfig(config *models.PayConfig) error {
	return s.db.Create(config).Error
}

// UpdatePayConfig 更新支付配置
func (s *ConfigService) UpdatePayConfig(id int, config *models.PayConfig) error {
	return s.db.Model(&models.PayConfig{}).Where("id = ?", id).Updates(config).Error
}

// DeletePayConfig 删除支付配置
func (s *ConfigService) DeletePayConfig(id int) error {
	return s.db.Delete(&models.PayConfig{}, id).Error
}

// GetUIConfigs 获取所有UI配置
func (s *ConfigService) GetUIConfigs() ([]models.UIConfig, error) {
	var configs []models.UIConfig
	err := s.db.Preload("Client.Brand").Find(&configs).Error
	return configs, err
}

// CreateUIConfig 创建UI配置
func (s *ConfigService) CreateUIConfig(config *models.UIConfig) error {
	return s.db.Create(config).Error
}

// UpdateUIConfig 更新UI配置
func (s *ConfigService) UpdateUIConfig(id int, config *models.UIConfig) error {
	return s.db.Model(&models.UIConfig{}).Where("id = ?", id).Updates(config).Error
}

// DeleteUIConfig 删除UI配置
func (s *ConfigService) DeleteUIConfig(id int) error {
	return s.db.Delete(&models.UIConfig{}, id).Error
}

// GetNovelConfigs 获取所有小说配置
func (s *ConfigService) GetNovelConfigs() ([]models.NovelConfig, error) {
	var configs []models.NovelConfig
	err := s.db.Preload("Client.Brand").Find(&configs).Error
	return configs, err
}

// CreateNovelConfig 创建小说配置
func (s *ConfigService) CreateNovelConfig(config *models.NovelConfig) error {
	return s.db.Create(config).Error
}

// UpdateNovelConfig 更新小说配置
func (s *ConfigService) UpdateNovelConfig(id int, config *models.NovelConfig) error {
	return s.db.Model(&models.NovelConfig{}).Where("id = ?", id).Updates(config).Error
}

// DeleteNovelConfig 删除小说配置
func (s *ConfigService) DeleteNovelConfig(id int) error {
	return s.db.Delete(&models.NovelConfig{}, id).Error
}

// GetWebsiteConfig 获取网站完整配置
func (s *ConfigService) GetWebsiteConfig(clientID int) (map[string]interface{}, error) {
	// 查询Client信息
	var client models.Client
	if err := s.db.Preload("Brand").First(&client, clientID).Error; err != nil {
		return nil, err
	}

	// 查询BaseConfig
	var baseConfig models.BaseConfig
	s.db.Where("client_id = ?", clientID).First(&baseConfig)

	// 查询CommonConfig
	var commonConfig models.CommonConfig
	s.db.Where("client_id = ?", clientID).First(&commonConfig)

	// 查询PayConfig
	var payConfig models.PayConfig
	s.db.Where("client_id = ?", clientID).First(&payConfig)

	// 查询UIConfig
	var uiConfig models.UIConfig
	s.db.Where("client_id = ?", clientID).First(&uiConfig)

	// 查询NovelConfig
	var novelConfig models.NovelConfig
	s.db.Where("client_id = ?", clientID).First(&novelConfig)

	// 构建响应数据
	response := map[string]interface{}{
		"client": map[string]interface{}{
			"id":         client.ID,
			"host":       client.Host,
			"created_at": client.CreatedAt,
			"updated_at": client.UpdatedAt,
			"brand": map[string]interface{}{
				"id":   client.Brand.ID,
				"code": client.Brand.Code,
			},
		},
		"base_config":   baseConfig,
		"common_config": commonConfig,
		"pay_config":    payConfig,
		"ui_config":     uiConfig,
		"novel_config":  novelConfig,
	}

	return response, nil
}
