package handlers

import (
	"strconv"

	"brand-config-api/models"
	"brand-config-api/services"
	"brand-config-api/utils"

	"github.com/gin-gonic/gin"
)

// NovelConfigHandler 小说配置处理器
type NovelConfigHandler struct {
	novelConfigService *services.NovelConfigService
}

// NewNovelConfigHandler 创建小说配置处理器实例
func NewNovelConfigHandler() *NovelConfigHandler {
	return &NovelConfigHandler{
		novelConfigService: services.NewNovelConfigService(),
	}
}

// GetNovelConfigs 获取所有小说配置
func (h *NovelConfigHandler) GetNovelConfigs(c *gin.Context) {
	configs, err := h.novelConfigService.GetNovelConfigs()
	if err != nil {
		utils.InternalServerError(c, "获取小说配置列表失败")
		return
	}

	utils.Success(c, gin.H{
		"data":  configs,
		"total": len(configs),
	}, "获取小说配置列表成功")
}

// CreateNovelConfig 创建小说配置
func (h *NovelConfigHandler) CreateNovelConfig(c *gin.Context) {
	var config models.NovelConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}

	if err := h.novelConfigService.CreateNovelConfig(&config); err != nil {
		utils.InternalServerError(c, "创建小说配置失败")
		return
	}

	utils.Created(c, gin.H{"data": config}, "小说配置创建成功")
}

// UpdateNovelConfig 更新小说配置
func (h *NovelConfigHandler) UpdateNovelConfig(c *gin.Context) {
	id := c.Param("id")
	configID, err := strconv.Atoi(id)
	if err != nil {
		utils.BadRequest(c, "无效的配置ID")
		return
	}

	var config models.NovelConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}

	if err := h.novelConfigService.UpdateNovelConfig(configID, &config); err != nil {
		utils.InternalServerError(c, "更新小说配置失败")
		return
	}

	utils.Success(c, gin.H{"data": config}, "小说配置更新成功")
}

// DeleteNovelConfig 删除小说配置
func (h *NovelConfigHandler) DeleteNovelConfig(c *gin.Context) {
	id := c.Param("id")
	configID, err := strconv.Atoi(id)
	if err != nil {
		utils.BadRequest(c, "无效的配置ID")
		return
	}

	if err := h.novelConfigService.DeleteNovelConfig(configID); err != nil {
		utils.InternalServerError(c, "删除小说配置失败")
		return
	}

	utils.Success(c, nil, "小说配置删除成功")
}
