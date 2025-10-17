package services

import (
	"errors"

	"brand-config-api/database"
	"brand-config-api/models"

	"gorm.io/gorm"
)

// ClientService 客户端服务
type ClientService struct {
	db *gorm.DB
}

// NewClientService 创建客户端服务实例
func NewClientService() *ClientService {
	return &ClientService{
		db: database.DB,
	}
}

// GetAllClients 获取所有客户端
func (s *ClientService) GetAllClients() ([]models.Client, error) {
	var clients []models.Client
	err := s.db.Preload("Brand").
		Preload("Brand.Type").
		Preload("BaseConfigs").
		Preload("CommonConfigs").
		Preload("PayConfigs").
		Preload("UIConfigs").
		Preload("NovelConfigs").
		Find(&clients).Error
	return clients, err
}

// GetClientByID 根据ID获取客户端
func (s *ClientService) GetClientByID(id int) (*models.Client, error) {
	var client models.Client
	err := s.db.Preload("Brand").
		Preload("Brand.Type").
		Preload("BaseConfigs").
		Preload("CommonConfigs").
		Preload("PayConfigs").
		Preload("UIConfigs").
		Preload("NovelConfigs").
		First(&client, id).Error
	if err != nil {
		return nil, err
	}
	return &client, nil
}

// CreateClient 创建客户端
func (s *ClientService) CreateClient(brandID int, host string) (*models.Client, error) {
	// 验证品牌是否存在
	var brand models.Brand
	if err := s.db.First(&brand, brandID).Error; err != nil {
		return nil, errors.New("brand not found")
	}

	// 验证host格式
	validHosts := []string{"h5", "tth5", "ksh5"}
	hostValid := false
	for _, validHost := range validHosts {
		if host == validHost {
			hostValid = true
			break
		}
	}
	if !hostValid {
		return nil, errors.New("invalid host. Must be one of: h5, tth5, ksh5")
	}

	// 检查是否已存在相同的brand_id + host组合
	var existingClient models.Client
	if err := s.db.Where("brand_id = ? AND host = ?", brandID, host).First(&existingClient).Error; err == nil {
		return nil, errors.New("client with this brand_id and host already exists")
	}

	client := models.Client{
		BrandID: brandID,
		Host:    host,
	}

	err := s.db.Create(&client).Error
	if err != nil {
		return nil, err
	}

	// 重新查询以获取关联的Brand信息
	s.db.Preload("Brand").First(&client, client.ID)

	return &client, nil
}

// CreateClientWithTx 在事务中创建客户端
func (s *ClientService) CreateClientWithTx(tx *gorm.DB, brandID int, host string) (*models.Client, error) {
	// 验证品牌是否存在
	var brand models.Brand
	if err := tx.First(&brand, brandID).Error; err != nil {
		return nil, errors.New("brand not found")
	}

	// 验证host格式
	validHosts := []string{"h5", "tth5", "ksh5"}
	hostValid := false
	for _, validHost := range validHosts {
		if host == validHost {
			hostValid = true
			break
		}
	}
	if !hostValid {
		return nil, errors.New("invalid host. Must be one of: h5, tth5, ksh5")
	}

	// 检查是否已存在相同的brand_id + host组合
	var existingClient models.Client
	if err := tx.Where("brand_id = ? AND host = ?", brandID, host).First(&existingClient).Error; err == nil {
		return nil, errors.New("client with this brand_id and host already exists")
	}

	client := models.Client{
		BrandID: brandID,
		Host:    host,
	}

	err := tx.Create(&client).Error
	if err != nil {
		return nil, err
	}

	// 手动设置Brand信息，因为GORM不会自动填充
	client.Brand = brand

	return &client, nil
}

// UpdateClient 更新客户端
func (s *ClientService) UpdateClient(id int, brandID int, host string) (*models.Client, error) {
	var client models.Client
	if err := s.db.First(&client, id).Error; err != nil {
		return nil, err
	}

	// 验证品牌是否存在
	var brand models.Brand
	if err := s.db.First(&brand, brandID).Error; err != nil {
		return nil, errors.New("brand not found")
	}

	// 验证host格式
	validHosts := []string{"h5", "tth5", "ksh5"}
	hostValid := false
	for _, validHost := range validHosts {
		if host == validHost {
			hostValid = true
			break
		}
	}
	if !hostValid {
		return nil, errors.New("invalid host. Must be one of: h5, tth5, ksh5")
	}

	// 检查是否与其他客户端冲突
	var existingClient models.Client
	if err := s.db.Where("brand_id = ? AND host = ? AND id != ?", brandID, host, id).First(&existingClient).Error; err == nil {
		return nil, errors.New("client with this brand_id and host already exists")
	}

	client.BrandID = brandID
	client.Host = host

	err := s.db.Save(&client).Error
	if err != nil {
		return nil, err
	}

	// 重新查询以获取关联的Brand信息
	s.db.Preload("Brand").First(&client, client.ID)

	return &client, nil
}

// DeleteClient 删除客户端
func (s *ClientService) DeleteClient(id int) error {
	// 检查是否有相关的配置
	var count int64
	s.db.Model(&models.BaseConfig{}).Where("client_id = ?", id).Count(&count)
	if count > 0 {
		return errors.New("cannot delete client with existing configurations")
	}

	return s.db.Delete(&models.Client{}, id).Error
}
