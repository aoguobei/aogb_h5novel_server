package handlers

import (
	"strconv"

	"brand-config-api/models"
	"brand-config-api/services"
	"brand-config-api/utils"

	"github.com/gin-gonic/gin"
)

// CommonConfigHandler 通用配置处理器
type CommonConfigHandler struct {
	commonConfigService *services.CommonConfigService
}

// NewCommonConfigHandler 创建通用配置处理器实例
func NewCommonConfigHandler() *CommonConfigHandler {
	return &CommonConfigHandler{
		commonConfigService: services.NewCommonConfigService(),
	}
}

// GetCommonConfigs 获取所有通用配置
func (h *CommonConfigHandler) GetCommonConfigs(c *gin.Context) {
	configs, err := h.commonConfigService.GetCommonConfigs()
	if err != nil {
		utils.InternalServerError(c, "获取通用配置列表失败")
		return
	}

	utils.Success(c, gin.H{
		"data":  configs,
		"total": len(configs),
	}, "获取通用配置列表成功")
}

// CreateCommonConfig 创建通用配置
func (h *CommonConfigHandler) CreateCommonConfig(c *gin.Context) {
	var config models.CommonConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}

	if err := h.commonConfigService.CreateCommonConfig(&config); err != nil {
		utils.InternalServerError(c, "创建通用配置失败")
		return
	}

	utils.Created(c, gin.H{"data": config}, "通用配置创建成功")
}

// UpdateCommonConfigByClientID 根据client_id更新通用配置
func (h *CommonConfigHandler) UpdateCommonConfigByClientID(c *gin.Context) {
	clientID := c.Param("client_id")
	clientIDInt, err := strconv.Atoi(clientID)
	if err != nil {
		utils.BadRequest(c, "无效的客户端ID")
		return
	}

	var config models.CommonConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}

	if err := h.commonConfigService.UpdateCommonConfigByClientID(clientIDInt, config); err != nil {
		utils.InternalServerError(c, "更新通用配置失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{"data": config}, "通用配置更新成功")
}

// DeleteCommonConfigByClientID 根据client_id删除通用配置
func (h *CommonConfigHandler) DeleteCommonConfigByClientID(c *gin.Context) {
	clientIDStr := c.Param("client_id")
	clientID, err := strconv.Atoi(clientIDStr)
	if err != nil {
		utils.BadRequest(c, "无效的客户端ID")
		return
	}

	err = h.commonConfigService.DeleteCommonConfigByClientID(clientID)
	if err != nil {
		utils.InternalServerError(c, "删除通用配置失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{
		"message":   "通用配置删除成功",
		"client_id": clientID,
	}, "通用配置删除成功")
}
