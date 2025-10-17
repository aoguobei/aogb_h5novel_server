package services

import (
	"errors"
	"fmt"

	"brand-config-api/database"
	"brand-config-api/models"

	"gorm.io/gorm"
)

// BrandService 品牌服务
type BrandService struct {
	db *gorm.DB
}

// NewBrandService 创建品牌服务实例
func NewBrandService() *BrandService {
	return &BrandService{
		db: database.DB,
	}
}

// GetAllBrands 获取所有品牌
func (s *BrandService) GetAllBrands() ([]models.Brand, error) {
	var brands []models.Brand
	err := s.db.Preload("Type").Preload("Clients").Find(&brands).Error
	return brands, err
}

// GetBrandByID 根据ID获取品牌
func (s *BrandService) GetBrandByID(id int) (*models.Brand, error) {
	var brand models.Brand
	err := s.db.Preload("Type").First(&brand, id).Error
	if err != nil {
		return nil, err
	}
	return &brand, nil
}

// CreateBrand 创建品牌
func (s *BrandService) CreateBrand(code string, typeID uint) (*models.Brand, error) {
	// 检查品牌代码是否已存在
	var existingBrand models.Brand
	if err := s.db.Where("code = ?", code).First(&existingBrand).Error; err == nil {
		return nil, errors.New("brand code already exists")
	}

	// 检查类型是否存在
	var typeData models.Type
	if err := s.db.First(&typeData, typeID).Error; err != nil {
		return nil, errors.New("type not found")
	}

	brand := models.Brand{
		Code:   code,
		TypeID: typeID,
	}

	// 创建数据库记录
	if err := s.db.Create(&brand).Error; err != nil {
		return nil, err
	}

	return &brand, nil
}

// UpdateBrand 更新品牌
func (s *BrandService) UpdateBrand(id int, code string, typeID uint) (*models.Brand, error) {
	var brand models.Brand
	if err := s.db.First(&brand, id).Error; err != nil {
		return nil, err
	}

	// 检查新代码是否与其他品牌冲突
	var existingBrand models.Brand
	if err := s.db.Where("code = ? AND id != ?", code, id).First(&existingBrand).Error; err == nil {
		return nil, errors.New("brand code already exists")
	}

	// 检查类型是否存在
	var typeData models.Type
	if err := s.db.First(&typeData, typeID).Error; err != nil {
		return nil, errors.New("type not found")
	}

	brand.Code = code
	brand.TypeID = typeID

	// 更新数据库记录
	if err := s.db.Save(&brand).Error; err != nil {
		return nil, err
	}

	return &brand, nil
}

// DeleteBrand 删除品牌
func (s *BrandService) DeleteBrand(id int) error {
	// 检查是否有相关的客户端
	var count int64
	s.db.Model(&models.Client{}).Where("brand_id = ?", id).Count(&count)
	if count > 0 {
		return errors.New("cannot delete brand with existing clients")
	}

	// 获取品牌信息
	var brand models.Brand
	if err := s.db.First(&brand, id).Error; err != nil {
		return err
	}

	// 删除数据库记录
	if err := s.db.Delete(&brand).Error; err != nil {
		return fmt.Errorf("failed to delete brand from database: %v", err)
	}

	return nil
}
