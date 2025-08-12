package services

import (
	"brand-config-api/database"
	"brand-config-api/models"

	"gorm.io/gorm"
)

// UIConfigService UI配置服务
type UIConfigService struct {
	db *gorm.DB
}

// NewUIConfigService 创建UI配置服务实例
func NewUIConfigService() *UIConfigService {
	return &UIConfigService{
		db: database.DB,
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

// CreateUIConfigFromRequestWithTx 从请求直接创建UI配置（在事务中）
func (s *UIConfigService) CreateUIConfigFromRequestWithTx(tx *gorm.DB, uiConfigReq UIConfigRequest, clientID int) (*models.UIConfig, error) {
	uiConfig := &models.UIConfig{
		ClientID:      clientID,
		ThemeBgMain:   uiConfigReq.ThemeBgMain,
		ThemeBgSecond: uiConfigReq.ThemeBgSecond,
		ThemeTextMain: uiConfigReq.ThemeTextMain,
	}

	return uiConfig, tx.Create(uiConfig).Error
}

// UpdateUIConfig 更新UI配置
func (s *UIConfigService) UpdateUIConfig(id int, config *models.UIConfig) error {
	return s.db.Model(&models.UIConfig{}).Where("id = ?", id).Updates(config).Error
}

// DeleteUIConfig 删除UI配置
func (s *UIConfigService) DeleteUIConfig(id int) error {
	return s.db.Delete(&models.UIConfig{}, id).Error
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
