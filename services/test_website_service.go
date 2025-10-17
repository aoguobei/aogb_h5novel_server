package services

import (
	"errors"

	"brand-config-api/database"
	"brand-config-api/models"

	"gorm.io/gorm"
)

// TestWebsiteService 测试网站服务
type TestWebsiteService struct {
	db *gorm.DB
}

// NewTestWebsiteService 创建测试网站服务实例
func NewTestWebsiteService() *TestWebsiteService {
	return &TestWebsiteService{
		db: database.DB,
	}
}

// GetAllTestWebsites 获取所有测试网站
func (s *TestWebsiteService) GetAllTestWebsites() ([]models.TestWebsite, error) {
	var websites []models.TestWebsite
	if err := s.db.Find(&websites).Error; err != nil {
		return nil, err
	}
	// 统计各网站关联的测试链接数量
	for i := range websites {
		var count int64
		s.db.Model(&models.TestLink{}).Where("website_id = ?", websites[i].ID).Count(&count)
		websites[i].TestLinksCount = int(count)
	}
	return websites, nil
}

// GetTestWebsiteByID 根据ID获取测试网站
func (s *TestWebsiteService) GetTestWebsiteByID(id uint) (*models.TestWebsite, error) {
	var website models.TestWebsite
	if err := s.db.Preload("TestLinks").First(&website, id).Error; err != nil {
		return nil, err
	}
	var count int64
	s.db.Model(&models.TestLink{}).Where("website_id = ?", website.ID).Count(&count)
	website.TestLinksCount = int(count)
	return &website, nil
}

// CreateTestWebsite 创建测试网站
func (s *TestWebsiteService) CreateTestWebsite(name, websiteType, scriptBase, prodDomain, testDomain string) (*models.TestWebsite, error) {
	// 唯一性校验：同 type + script_base + prod_domain + test_domain 不能重复
	var existing models.TestWebsite
	err := s.db.Where("type = ? AND script_base = ? AND prod_domain = ? AND test_domain = ?",
		websiteType, scriptBase, prodDomain, testDomain).First(&existing).Error
	if err == nil {
		return nil, errors.New("同类型与域名/基础路径重复，禁止创建重复网站")
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	website := &models.TestWebsite{
		Name:       name,
		Type:       websiteType,
		ScriptBase: scriptBase,
		ProdDomain: prodDomain,
		TestDomain: testDomain,
	}

	if err := s.db.Create(website).Error; err != nil {
		return nil, err
	}

	return s.GetTestWebsiteByID(website.ID)
}

// UpdateTestWebsite 更新测试网站
func (s *TestWebsiteService) UpdateTestWebsite(id uint, name, websiteType, scriptBase, prodDomain, testDomain string) (*models.TestWebsite, error) {
	var website models.TestWebsite
	if err := s.db.First(&website, id).Error; err != nil {
		return nil, errors.New("test website not found")
	}

	website.Name = name
	website.Type = websiteType
	website.ScriptBase = scriptBase
	website.ProdDomain = prodDomain
	website.TestDomain = testDomain

	if err := s.db.Save(&website).Error; err != nil {
		return nil, err
	}

	return s.GetTestWebsiteByID(website.ID)
}

// DeleteTestWebsite 删除测试网站
func (s *TestWebsiteService) DeleteTestWebsite(id uint) error {
	// 检查是否有关联测试链接
	var count int64
	s.db.Model(&models.TestLink{}).Where("website_id = ?", id).Count(&count)
	if count > 0 {
		return errors.New("该测试网站存在关联的测试链接，无法删除")
	}

	var website models.TestWebsite
	if err := s.db.First(&website, id).Error; err != nil {
		return errors.New("test website not found")
	}

	return s.db.Delete(&website).Error
}

// GetTestLinksByWebsiteID 根据测试网站ID获取测试链接
func (s *TestWebsiteService) GetTestLinksByWebsiteID(websiteID uint) ([]models.TestLink, error) {
	var testLinks []models.TestLink
	err := s.db.Preload("Website").
		Where("website_id = ?", websiteID).
		Find(&testLinks).Error
	return testLinks, err
}
