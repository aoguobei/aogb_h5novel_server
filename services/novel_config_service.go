package services

import (
	"brand-config-api/database"
	"brand-config-api/models"

	"gorm.io/gorm"
)

// NovelConfigService 小说配置服务
type NovelConfigService struct {
	db *gorm.DB
}

// NewNovelConfigService 创建小说配置服务实例
func NewNovelConfigService() *NovelConfigService {
	return &NovelConfigService{
		db: database.DB,
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

// CreateNovelConfigFromRequestWithTx 从请求直接创建小说配置（在事务中）
func (s *NovelConfigService) CreateNovelConfigFromRequestWithTx(tx *gorm.DB, novelConfigReq NovelConfigRequest, clientID int) (*models.NovelConfig, error) {
	novelConfig := &models.NovelConfig{
		ClientID:              clientID,
		TTJumpHomeUrl:         novelConfigReq.TTJumpHomeUrl,
		TTLoginCallbackDomain: novelConfigReq.TTLoginCallbackDomain,
	}

	return novelConfig, tx.Create(novelConfig).Error
}

// UpdateNovelConfig 更新小说配置
func (s *NovelConfigService) UpdateNovelConfig(id int, config *models.NovelConfig) error {
	return s.db.Model(&models.NovelConfig{}).Where("id = ?", id).Updates(config).Error
}

// DeleteNovelConfig 删除小说配置
func (s *NovelConfigService) DeleteNovelConfig(id int) error {
	return s.db.Delete(&models.NovelConfig{}, id).Error
}

// FormatNovelConfig 格式化小说配置
func (s *NovelConfigService) FormatNovelConfig(novelConfig models.NovelConfig) map[string]interface{} {
	return map[string]interface{}{
		"tt_jump_home_url":         novelConfig.TTJumpHomeUrl,
		"tt_login_callback_domain": novelConfig.TTLoginCallbackDomain,
	}
}
