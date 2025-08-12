package handlers

import (
	"strconv"

	"brand-config-api/models"
	"brand-config-api/services"
	"brand-config-api/utils"

	"github.com/gin-gonic/gin"
)

// UIConfigHandler UI配置处理器
type UIConfigHandler struct {
	uiConfigService *services.UIConfigService
}

// NewUIConfigHandler 创建UI配置处理器实例
func NewUIConfigHandler() *UIConfigHandler {
	return &UIConfigHandler{
		uiConfigService: services.NewUIConfigService(),
	}
}

// GetUIConfigs 获取所有UI配置
func (h *UIConfigHandler) GetUIConfigs(c *gin.Context) {
	configs, err := h.uiConfigService.GetUIConfigs()
	if err != nil {
		utils.InternalServerError(c, "获取UI配置列表失败")
		return
	}

	utils.Success(c, gin.H{
		"data":  configs,
		"total": len(configs),
	}, "获取UI配置列表成功")
}

// CreateUIConfig 创建UI配置
func (h *UIConfigHandler) CreateUIConfig(c *gin.Context) {
	var config models.UIConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}

	if err := h.uiConfigService.CreateUIConfig(&config); err != nil {
		utils.InternalServerError(c, "创建UI配置失败")
		return
	}

	utils.Created(c, gin.H{"data": config}, "UI配置创建成功")
}

// UpdateUIConfig 更新UI配置
func (h *UIConfigHandler) UpdateUIConfig(c *gin.Context) {
	id := c.Param("id")
	configID, err := strconv.Atoi(id)
	if err != nil {
		utils.BadRequest(c, "无效的配置ID")
		return
	}

	var config models.UIConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}

	if err := h.uiConfigService.UpdateUIConfig(configID, &config); err != nil {
		utils.InternalServerError(c, "更新UI配置失败")
		return
	}

	utils.Success(c, gin.H{"data": config}, "UI配置更新成功")
}

// DeleteUIConfig 删除UI配置
func (h *UIConfigHandler) DeleteUIConfig(c *gin.Context) {
	id := c.Param("id")
	configID, err := strconv.Atoi(id)
	if err != nil {
		utils.BadRequest(c, "无效的配置ID")
		return
	}

	if err := h.uiConfigService.DeleteUIConfig(configID); err != nil {
		utils.InternalServerError(c, "删除UI配置失败")
		return
	}

	utils.Success(c, nil, "UI配置删除成功")
}
