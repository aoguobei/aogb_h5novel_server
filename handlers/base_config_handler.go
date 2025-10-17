package handlers

import (
	"net/http"
	"strconv"

	"brand-config-api/models"
	"brand-config-api/services"
	"brand-config-api/utils"

	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
)

// BaseConfigHandler 基础配置控制器
type BaseConfigHandler struct {
	baseConfigService *services.BaseConfigService
}

// NewBaseConfigHandler 创建基础配置控制器
func NewBaseConfigHandler() *BaseConfigHandler {
	return &BaseConfigHandler{
		baseConfigService: services.NewBaseConfigService(),
	}
}

// GetBaseConfigs 获取所有基础配置
func (h *BaseConfigHandler) GetBaseConfigs(c *gin.Context) {
	configs, err := h.baseConfigService.GetBaseConfigs()
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "获取基础配置失败")
		return
	}

	utils.Success(c, configs, "获取基础配置成功")
}

// GetBaseConfigByID 根据ID获取基础配置
func (h *BaseConfigHandler) GetBaseConfigByID(c *gin.Context) {
	idStr := c.Param("id")
	configID, err := strconv.Atoi(idStr)
	if err != nil {
		utils.BadRequest(c, "无效的配置ID")
		return
	}

	config, err := h.baseConfigService.GetBaseConfigByID(configID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.NotFound(c, "配置不存在")
		} else {
			utils.Error(c, http.StatusInternalServerError, "获取配置失败")
		}
		return
	}

	utils.Success(c, config, "获取配置成功")
}

// CreateBaseConfig 创建基础配置
func (h *BaseConfigHandler) CreateBaseConfig(c *gin.Context) {
	var config models.BaseConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}

	if err := h.baseConfigService.CreateBaseConfig(&config); err != nil {
		utils.InternalServerError(c, "创建基础配置失败")
		return
	}

	utils.Created(c, gin.H{"data": config}, "基础配置创建成功")
}

// DeleteBaseConfigByClientID 根据client_id删除基础配置
func (h *BaseConfigHandler) DeleteBaseConfigByClientID(c *gin.Context) {
	clientIDStr := c.Param("clientId")
	clientID, err := strconv.Atoi(clientIDStr)
	if err != nil {
		utils.BadRequest(c, "无效的客户端ID")
		return
	}

	err = h.baseConfigService.DeleteBaseConfigByClientID(clientID)
	if err != nil {
		utils.InternalServerError(c, "删除配置失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{
		"message":   "基础配置删除成功",
		"client_id": clientID,
	}, "基础配置删除成功")
}

// UpdateBaseConfigByClientID 根据client_id更新基础配置
func (h *BaseConfigHandler) UpdateBaseConfigByClientID(c *gin.Context) {
	// 获取client_id参数
	clientIDStr := c.Param("clientId")
	clientID, err := strconv.Atoi(clientIDStr)
	if err != nil {
		utils.BadRequest(c, "无效的客户端ID")
		return
	}

	// 解析请求体
	var baseConfig models.BaseConfig
	if err := c.ShouldBindJSON(&baseConfig); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// 更新配置（使用新的回滚服务）
	err = h.baseConfigService.UpdateBaseConfigByClientID(clientID, baseConfig)
	if err != nil {
		utils.InternalServerError(c, "更新基础配置失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{
		"message":   "基础配置更新成功",
		"client_id": clientID,
	}, "基础配置更新成功")
}

// GetBaseConfigByClientID 根据client_id获取基础配置
func (h *BaseConfigHandler) GetBaseConfigByClientID(c *gin.Context) {
	// 获取client_id参数
	clientIDStr := c.Param("clientId")
	clientID, err := strconv.Atoi(clientIDStr)
	if err != nil {
		utils.BadRequest(c, "无效的客户端ID")
		return
	}

	// 获取配置
	config, err := h.baseConfigService.GetBaseConfigByClientID(clientID)
	if err != nil {
		utils.NotFound(c, "获取基础配置失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{
		"data":      config,
		"client_id": clientID,
	}, "获取基础配置成功")
}
