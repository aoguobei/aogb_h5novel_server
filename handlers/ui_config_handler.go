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

// UpdateUIConfigByClientID 根据client_id更新UI配置
func (h *UIConfigHandler) UpdateUIConfigByClientID(c *gin.Context) {
	clientID := c.Param("client_id")
	clientIDInt, err := strconv.Atoi(clientID)
	if err != nil {
		utils.BadRequest(c, "无效的客户端ID")
		return
	}

	var config models.UIConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}

	if err := h.uiConfigService.UpdateUIConfigByClientID(clientIDInt, config); err != nil {
		utils.InternalServerError(c, "更新UI配置失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{"data": config}, "UI配置更新成功")
}

// DeleteUIConfigByClientID 根据client_id删除UI配置
func (h *UIConfigHandler) DeleteUIConfigByClientID(c *gin.Context) {
	clientIDStr := c.Param("client_id")
	clientID, err := strconv.Atoi(clientIDStr)
	if err != nil {
		utils.BadRequest(c, "无效的客户端ID")
		return
	}

	err = h.uiConfigService.DeleteUIConfigByClientID(clientID)
	if err != nil {
		utils.InternalServerError(c, "删除UI配置失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{
		"message":   "UI配置删除成功",
		"client_id": clientID,
	}, "UI配置删除成功")
}
