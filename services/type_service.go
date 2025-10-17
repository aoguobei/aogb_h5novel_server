package services

import (
	"errors"

	"brand-config-api/database"
	"brand-config-api/models"

	"gorm.io/gorm"
)

// TypeService 类型服务
type TypeService struct {
	db *gorm.DB
}

// NewTypeService 创建类型服务实例
func NewTypeService() *TypeService {
	return &TypeService{
		db: database.DB,
	}
}

// GetAllTypes 获取所有类型
func (s *TypeService) GetAllTypes() ([]models.Type, error) {
	var types []models.Type
	err := s.db.Find(&types).Error
	return types, err
}

// GetTypeByID 根据ID获取类型
func (s *TypeService) GetTypeByID(id uint) (*models.Type, error) {
	var typeData models.Type
	err := s.db.First(&typeData, id).Error
	if err != nil {
		return nil, err
	}
	return &typeData, nil
}

// CreateType 创建类型
func (s *TypeService) CreateType(name, code string) (*models.Type, error) {
	// 检查类型代码是否已存在
	var existingType models.Type
	if err := s.db.Where("code = ?", code).First(&existingType).Error; err == nil {
		return nil, errors.New("type code already exists")
	}

	typeData := models.Type{
		Name: name,
		Code: code,
	}

	if err := s.db.Create(&typeData).Error; err != nil {
		return nil, err
	}

	return &typeData, nil
}

// UpdateType 更新类型
func (s *TypeService) UpdateType(id uint, name, code string) (*models.Type, error) {
	var typeData models.Type
	if err := s.db.First(&typeData, id).Error; err != nil {
		return nil, err
	}

	// 检查新代码是否与其他类型冲突
	var existingType models.Type
	if err := s.db.Where("code = ? AND id != ?", code, id).First(&existingType).Error; err == nil {
		return nil, errors.New("type code already exists")
	}

	typeData.Name = name
	typeData.Code = code

	if err := s.db.Save(&typeData).Error; err != nil {
		return nil, err
	}

	return &typeData, nil
}

// DeleteType 删除类型
func (s *TypeService) DeleteType(id uint) error {
	// 检查是否有相关的品牌
	var count int64
	s.db.Model(&models.Brand{}).Where("type_id = ?", id).Count(&count)
	if count > 0 {
		return errors.New("cannot delete type with existing brands")
	}

	var typeData models.Type
	if err := s.db.First(&typeData, id).Error; err != nil {
		return err
	}

	return s.db.Delete(&typeData).Error
}
