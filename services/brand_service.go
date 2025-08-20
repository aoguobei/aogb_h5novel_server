package services

import (
	"errors"
	"fmt"
	"log"

	"brand-config-api/database"
	"brand-config-api/models"
	"brand-config-api/utils/rollback"

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
	err := s.db.Preload("Clients").Find(&brands).Error
	return brands, err
}

// GetBrandByID 根据ID获取品牌
func (s *BrandService) GetBrandByID(id int) (*models.Brand, error) {
	var brand models.Brand
	err := s.db.First(&brand, id).Error
	if err != nil {
		return nil, err
	}
	return &brand, nil
}

// CreateBrand 创建品牌
func (s *BrandService) CreateBrand(code string) (*models.Brand, error) {
	// 检查品牌代码是否已存在
	var existingBrand models.Brand
	if err := s.db.Where("code = ?", code).First(&existingBrand).Error; err == nil {
		return nil, errors.New("brand code already exists")
	}

	brand := models.Brand{
		Code: code,
	}

	// 先创建数据库记录
	if err := s.db.Create(&brand).Error; err != nil {
		return nil, err
	}

	// 数据库创建成功，再处理文件操作
	fileService := NewFileService()
	fileManager := rollback.NewFileRollback(fileService.config)

	// 更新 _host.js 文件，添加新的 brand
	if err := fileService.UpdateHostFileForBrand(code, fileManager); err != nil {
		// 文件操作失败，记录错误但不阻止创建
		// 因为数据库已经创建了，文件操作失败不影响数据一致性
		log.Printf("Warning: failed to update _host.js for brand '%s': %v", code, err)
		log.Printf("Note: brand '%s' has been created in database, but file update failed", code)
	}

	// 清理回滚数据
	if err := fileManager.Clear(); err != nil {
		log.Printf("Warning: failed to clear rollback data: %v", err)
	}

	return &brand, nil
}

// UpdateBrand 更新品牌
func (s *BrandService) UpdateBrand(id int, code string) (*models.Brand, error) {
	var brand models.Brand
	if err := s.db.First(&brand, id).Error; err != nil {
		return nil, err
	}

	// 检查新代码是否与其他品牌冲突
	var existingBrand models.Brand
	if err := s.db.Where("code = ? AND id != ?", code, id).First(&existingBrand).Error; err == nil {
		return nil, errors.New("brand code already exists")
	}

	oldCode := brand.Code
	brand.Code = code

	// 先更新数据库记录
	if err := s.db.Save(&brand).Error; err != nil {
		return nil, err
	}

	// 数据库更新成功，再处理文件操作
	fileService := NewFileService()
	fileManager := rollback.NewFileRollback(fileService.config)

	// 如果品牌代码发生变化，需要更新 _host.js 文件
	if oldCode != code {
		// 先删除旧的品牌配置
		if err := fileService.removeHostFileBrandConfig(oldCode, fileManager); err != nil {
			log.Printf("Warning: failed to remove old brand config from _host.js: %v", err)
		}

		// 再添加新的品牌配置
		if err := fileService.UpdateHostFileForBrand(code, fileManager); err != nil {
			log.Printf("Warning: failed to add new brand config to _host.js: %v", err)
		}
	}

	// 清理回滚数据
	if err := fileManager.Clear(); err != nil {
		log.Printf("Warning: failed to clear rollback data: %v", err)
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

	// 先删除数据库记录
	if err := s.db.Delete(&brand).Error; err != nil {
		return fmt.Errorf("failed to delete brand from database: %v", err)
	}

	// 数据库删除成功，再处理文件删除
	fileService := NewFileService()
	fileManager := rollback.NewFileRollback(fileService.config)

	// 删除_host.js中对应的品牌配置
	if err := fileService.removeHostFileBrandConfig(brand.Code, fileManager); err != nil {
		// 文件操作失败，记录错误但不阻止删除
		// 因为数据库已经删除了，文件操作失败不影响数据一致性
		log.Printf("Warning: failed to remove brand config from _host.js: %v", err)
		log.Printf("Note: brand '%s' has been deleted from database, but file cleanup failed", brand.Code)
	}

	// 清理回滚数据
	if err := fileManager.Clear(); err != nil {
		log.Printf("Warning: failed to clear rollback data: %v", err)
	}

	return nil
}
