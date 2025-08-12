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

// UpdateCommonConfig 更新通用配置
func (h *CommonConfigHandler) UpdateCommonConfig(c *gin.Context) {
	id := c.Param("id")
	configID, err := strconv.Atoi(id)
	if err != nil {
		utils.BadRequest(c, "无效的配置ID")
		return
	}

	var config models.CommonConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}

	if err := h.commonConfigService.UpdateCommonConfig(configID, &config); err != nil {
		utils.InternalServerError(c, "更新通用配置失败")
		return
	}

	utils.Success(c, gin.H{"data": config}, "通用配置更新成功")
}

// DeleteCommonConfig 删除通用配置
func (h *CommonConfigHandler) DeleteCommonConfig(c *gin.Context) {
	id := c.Param("id")
	configID, err := strconv.Atoi(id)
	if err != nil {
		utils.BadRequest(c, "无效的配置ID")
		return
	}

	if err := h.commonConfigService.DeleteCommonConfig(configID); err != nil {
		utils.InternalServerError(c, "删除通用配置失败")
		return
	}

	utils.Success(c, nil, "通用配置删除成功")
}
