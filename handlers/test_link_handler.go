package handlers

import (
	"strconv"

	"brand-config-api/services"
	"brand-config-api/utils"

	"github.com/gin-gonic/gin"
)

// TestLinkHandler 测试链接控制器
type TestLinkHandler struct {
	testLinkService *services.TestLinkService
}

// NewTestLinkHandler 创建测试链接控制器
func NewTestLinkHandler() *TestLinkHandler {
	return &TestLinkHandler{
		testLinkService: services.NewTestLinkService(),
	}
}

// CreateTestLink 创建测试链接
func (h *TestLinkHandler) CreateTestLink(c *gin.Context) {
	var requestData struct {
		WebsiteID uint   `json:"website_id" binding:"required"`
		TestURL   string `json:"test_url" binding:"required"`
		TestTitle string `json:"test_title" binding:"required"`
	}

	if err := c.ShouldBindJSON(&requestData); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	testLink, err := h.testLinkService.CreateTestLink(
		requestData.WebsiteID,
		requestData.TestURL,
		requestData.TestTitle,
	)

	if err != nil {
		utils.Error(c, 500, "创建失败: "+err.Error())
		return
	}

	utils.Success(c, testLink, "测试链接创建成功")
}

// BatchCreateTestLinks 批量创建测试链接
func (h *TestLinkHandler) BatchCreateTestLinks(c *gin.Context) {
	var requestData struct {
		WebsiteIDs []uint                  `json:"website_ids" binding:"required"`
		TestLinks  []services.TestLinkData `json:"test_links" binding:"required"`
	}

	if err := c.ShouldBindJSON(&requestData); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	createdTestLinks, err := h.testLinkService.BatchCreateTestLinks(requestData.WebsiteIDs, requestData.TestLinks)
	if err != nil {
		utils.Error(c, 500, "批量创建失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{
		"list":  createdTestLinks,
		"total": len(createdTestLinks),
	}, "批量创建成功")
}

// UpdateTestLink 更新测试链接
func (h *TestLinkHandler) UpdateTestLink(c *gin.Context) {
	id := c.Param("id")
	testLinkID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		utils.BadRequest(c, "无效的测试链接ID")
		return
	}

	var requestData struct {
		WebsiteID uint   `json:"website_id" binding:"required"`
		TestURL   string `json:"test_url" binding:"required"`
		TestTitle string `json:"test_title" binding:"required"`
	}

	if err := c.ShouldBindJSON(&requestData); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	testLink, err := h.testLinkService.UpdateTestLink(
		testLinkID,
		requestData.WebsiteID,
		requestData.TestURL,
		requestData.TestTitle,
	)

	if err != nil {
		utils.Error(c, 500, "更新失败: "+err.Error())
		return
	}

	utils.Success(c, testLink, "测试链接更新成功")
}

// DeleteTestLink 删除测试链接
func (h *TestLinkHandler) DeleteTestLink(c *gin.Context) {
	id := c.Param("id")
	testLinkID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		utils.BadRequest(c, "无效的测试链接ID")
		return
	}

	err = h.testLinkService.DeleteTestLink(testLinkID)
	if err != nil {
		utils.Conflict(c, err.Error())
		return
	}

	utils.Success(c, nil, "测试链接删除成功")
}
