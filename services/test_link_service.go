package services

import (
	"errors"

	"brand-config-api/database"
	"brand-config-api/models"

	"gorm.io/gorm"
)

// TestLinkService 测试链接服务
type TestLinkService struct {
	db *gorm.DB
}

// NewTestLinkService 创建测试链接服务实例
func NewTestLinkService() *TestLinkService {
	return &TestLinkService{
		db: database.DB,
	}
}

// GetAllTestLinks 获取所有测试链接
func (s *TestLinkService) GetAllTestLinks() ([]models.TestLink, error) {
	var testLinks []models.TestLink
	err := s.db.Preload("Website").Find(&testLinks).Error
	return testLinks, err
}

// GetTestLinksByWebsiteID 根据测试网站ID获取测试链接
func (s *TestLinkService) GetTestLinksByWebsiteID(websiteID uint) ([]models.TestLink, error) {
	var testLinks []models.TestLink
	err := s.db.Preload("Website").
		Where("website_id = ?", websiteID).
		Find(&testLinks).Error
	return testLinks, err
}

// GetTestLinkByID 根据ID获取测试链接
func (s *TestLinkService) GetTestLinkByID(id int64) (*models.TestLink, error) {
	var testLink models.TestLink
	err := s.db.Preload("Website").First(&testLink, id).Error
	if err != nil {
		return nil, err
	}
	return &testLink, nil
}

// CreateTestLink 创建测试链接
func (s *TestLinkService) CreateTestLink(websiteID uint, testURL, testTitle string) (*models.TestLink, error) {
	// 验证测试网站是否存在
	var website models.TestWebsite
	if err := s.db.First(&website, websiteID).Error; err != nil {
		return nil, errors.New("test website not found")
	}

	testLink := &models.TestLink{
		WebsiteID: websiteID,
		TestURL:   testURL,
		TestTitle: testTitle,
	}

	if err := s.db.Create(testLink).Error; err != nil {
		return nil, err
	}

	// 重新加载以获取关联数据
	return s.GetTestLinkByID(int64(testLink.ID))
}

// TestLinkData 测试链接数据
type TestLinkData struct {
	TestTitle string `json:"test_title"`
	TestURL   string `json:"test_url"`
}

// getTestLinkIDs 获取测试链接ID列表
func getTestLinkIDs(testLinks []models.TestLink) []int64 {
	ids := make([]int64, len(testLinks))
	for i, testLink := range testLinks {
		ids[i] = int64(testLink.ID)
	}
	return ids
}

// BatchCreateTestLinks 批量创建测试链接
func (s *TestLinkService) BatchCreateTestLinks(websiteIDs []uint, testLinks []TestLinkData) ([]models.TestLink, error) {
	// 验证所有测试网站是否存在
	for _, websiteID := range websiteIDs {
		var website models.TestWebsite
		if err := s.db.First(&website, websiteID).Error; err != nil {
			return nil, errors.New("test website not found")
		}
	}

	// 准备批量插入数据
	var testLinksToCreate []models.TestLink
	for _, websiteID := range websiteIDs {
		for _, testLinkData := range testLinks {
			testLink := models.TestLink{
				WebsiteID: websiteID,
				TestURL:   testLinkData.TestURL,
				TestTitle: testLinkData.TestTitle,
			}
			testLinksToCreate = append(testLinksToCreate, testLink)
		}
	}

	// 批量创建
	if err := s.db.Create(&testLinksToCreate).Error; err != nil {
		return nil, err
	}

	// 重新加载以获取关联数据
	var createdTestLinks []models.TestLink
	err := s.db.Preload("Website").
		Where("id IN ?", getTestLinkIDs(testLinksToCreate)).
		Find(&createdTestLinks).Error

	return createdTestLinks, err
}

// UpdateTestLink 更新测试链接
func (s *TestLinkService) UpdateTestLink(id int64, websiteID uint, testURL, testTitle string) (*models.TestLink, error) {
	// 验证测试链接是否存在
	var testLink models.TestLink
	if err := s.db.First(&testLink, id).Error; err != nil {
		return nil, errors.New("test link not found")
	}

	// 验证测试网站是否存在
	var website models.TestWebsite
	if err := s.db.First(&website, websiteID).Error; err != nil {
		return nil, errors.New("test website not found")
	}

	// 更新字段
	testLink.WebsiteID = websiteID
	testLink.TestURL = testURL
	testLink.TestTitle = testTitle

	if err := s.db.Save(&testLink).Error; err != nil {
		return nil, err
	}

	// 重新加载以获取关联数据
	return s.GetTestLinkByID(int64(testLink.ID))
}

// DeleteTestLink 删除测试链接
func (s *TestLinkService) DeleteTestLink(id int64) error {
	// 验证测试链接是否存在
	var testLink models.TestLink
	if err := s.db.First(&testLink, id).Error; err != nil {
		return errors.New("test link not found")
	}

	return s.db.Delete(&testLink).Error
}

// GetTestLinksByClientID 根据客户端ID获取测试链接 - 这个方法需要重新设计
// 由于TestWebsite不再有ClientID，这个方法可能需要通过其他方式实现
// 或者需要从路由中移除这个功能
func (s *TestLinkService) GetTestLinksByClientID(clientID int64) ([]models.TestLink, error) {
	// 这个方法现在无法实现，因为TestWebsite没有ClientID
	// 可以考虑移除这个方法，或者通过其他方式实现
	return nil, errors.New("this method is no longer supported as TestWebsite does not have ClientID")
}

// BatchUpdateDomains 批量更新域名 - 这个方法也需要重新设计
// 由于TestWebsite不再有ClientID，这个方法可能需要通过其他方式实现
func (s *TestLinkService) BatchUpdateDomains(clientID int64, domains map[string]string) error {
	// 这个方法现在无法实现，因为TestWebsite没有ClientID
	// 可以考虑移除这个方法，或者通过其他方式实现
	return errors.New("this method is no longer supported as TestWebsite does not have ClientID")
}
