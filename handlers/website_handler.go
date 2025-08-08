package handlers

import (
	"brand-config-api/services"
	"brand-config-api/utils"

	"github.com/gin-gonic/gin"
)

// WebsiteHandler 网站控制器
type WebsiteHandler struct {
	websiteService *services.WebsiteService
}

// NewWebsiteHandler 创建网站控制器
func NewWebsiteHandler() *WebsiteHandler {
	return &WebsiteHandler{
		websiteService: services.NewWebsiteService(),
	}
}

// CreateWebsite 创建网站
func (h *WebsiteHandler) CreateWebsite(c *gin.Context) {
	var req services.CreateWebsiteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}

	result, err := h.websiteService.CreateWebsite(&req)
	if err != nil {
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.Created(c, gin.H{
		"message": "Website created successfully",
		"data":    result,
	}, "网站创建成功")
}
