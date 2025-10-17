package handlers

import (
	"net/http"
	"strconv"

	"brand-config-api/services"

	"github.com/gin-gonic/gin"
)

// TestWebsiteHandler 测试网站处理器
type TestWebsiteHandler struct {
	service *services.TestWebsiteService
}

// NewTestWebsiteHandler 创建测试网站处理器
func NewTestWebsiteHandler() *TestWebsiteHandler {
	return &TestWebsiteHandler{
		service: services.NewTestWebsiteService(),
	}
}

// GetTestWebsites 获取所有测试网站
func (h *TestWebsiteHandler) GetTestWebsites(c *gin.Context) {
	websites, err := h.service.GetAllTestWebsites()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取测试网站列表失败",
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "获取测试网站列表成功",
		"data": gin.H{
			"list":  websites,
			"total": len(websites),
		},
	})
}

// GetTestWebsite 获取单个测试网站
func (h *TestWebsiteHandler) GetTestWebsite(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的ID",
			"data":    nil,
		})
		return
	}

	website, err := h.service.GetTestWebsiteByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "测试网站不存在",
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "获取测试网站成功",
		"data":    website,
	})
}

// CreateTestWebsite 创建测试网站
func (h *TestWebsiteHandler) CreateTestWebsite(c *gin.Context) {
	var req struct {
		Name       string `json:"name" binding:"required"`
		Type       string `json:"type" binding:"required"`
		ScriptBase string `json:"script_base"`
		ProdDomain string `json:"prod_domain"`
		TestDomain string `json:"test_domain"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "请求参数错误",
			"data":    nil,
		})
		return
	}

	website, err := h.service.CreateTestWebsite(
		req.Name,
		req.Type,
		req.ScriptBase,
		req.ProdDomain,
		req.TestDomain,
	)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{
			"success": false,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "测试网站创建成功",
		"data":    website,
	})
}

// UpdateTestWebsite 更新测试网站
func (h *TestWebsiteHandler) UpdateTestWebsite(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的ID",
			"data":    nil,
		})
		return
	}

	var req struct {
		Name       string `json:"name" binding:"required"`
		Type       string `json:"type" binding:"required"`
		ScriptBase string `json:"script_base"`
		ProdDomain string `json:"prod_domain"`
		TestDomain string `json:"test_domain"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "请求参数错误",
			"data":    nil,
		})
		return
	}

	website, err := h.service.UpdateTestWebsite(
		uint(id),
		req.Name,
		req.Type,
		req.ScriptBase,
		req.ProdDomain,
		req.TestDomain,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "更新测试网站失败",
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "测试网站更新成功",
		"data":    website,
	})
}

// DeleteTestWebsite 删除测试网站
func (h *TestWebsiteHandler) DeleteTestWebsite(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的ID",
			"data":    nil,
		})
		return
	}

	if err := h.service.DeleteTestWebsite(uint(id)); err != nil {
		c.JSON(http.StatusConflict, gin.H{
			"success": false,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "测试网站删除成功",
		"data":    nil,
	})
}

// GetTestLinksByWebsiteID 获取测试网站的测试链接
func (h *TestWebsiteHandler) GetTestLinksByWebsiteID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的ID",
			"data":    nil,
		})
		return
	}

	testLinks, err := h.service.GetTestLinksByWebsiteID(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取测试链接失败",
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "获取测试链接成功",
		"data": gin.H{
			"list":  testLinks,
			"total": len(testLinks),
		},
	})
}
