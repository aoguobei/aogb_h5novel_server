package handlers

import (
	"strconv"

	"brand-config-api/models"
	"brand-config-api/services"
	"brand-config-api/utils"

	"github.com/gin-gonic/gin"
)

// ConfigHandler 配置控制器
type ConfigHandler struct {
	configService *services.ConfigService
}

// NewConfigHandler 创建配置控制器
func NewConfigHandler() *ConfigHandler {
	return &ConfigHandler{
		configService: services.NewConfigService(),
	}
}

// GetBaseConfigs 获取所有基础配置
func (h *ConfigHandler) GetBaseConfigs(c *gin.Context) {
	configs, err := h.configService.GetBaseConfigs()
	if err != nil {
		utils.InternalServerError(c, "获取基础配置列表失败")
		return
	}

	utils.Success(c, gin.H{
		"data":  configs,
		"total": len(configs),
	}, "获取基础配置列表成功")
}

// GetBaseConfig 获取单个基础配置
func (h *ConfigHandler) GetBaseConfig(c *gin.Context) {
	id := c.Param("id")
	configID, err := strconv.Atoi(id)
	if err != nil {
		utils.BadRequest(c, "无效的配置ID")
		return
	}

	config, err := h.configService.GetBaseConfigByID(configID)
	if err != nil {
		utils.NotFound(c, "基础配置不存在")
		return
	}

	utils.Success(c, gin.H{"data": config}, "获取基础配置成功")
}

// CreateBaseConfig 创建基础配置
func (h *ConfigHandler) CreateBaseConfig(c *gin.Context) {
	var config models.BaseConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}

	err := h.configService.CreateBaseConfig(&config)
	if err != nil {
		utils.InternalServerError(c, "创建基础配置失败")
		return
	}

	utils.Created(c, gin.H{"data": config}, "基础配置创建成功")
}

// UpdateBaseConfig 更新基础配置
func (h *ConfigHandler) UpdateBaseConfig(c *gin.Context) {
	id := c.Param("id")
	configID, err := strconv.Atoi(id)
	if err != nil {
		utils.BadRequest(c, "无效的配置ID")
		return
	}

	var config models.BaseConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}

	err = h.configService.UpdateBaseConfig(configID, &config)
	if err != nil {
		utils.InternalServerError(c, "更新基础配置失败")
		return
	}

	utils.Success(c, gin.H{"data": config}, "基础配置更新成功")
}

// DeleteBaseConfig 删除基础配置
func (h *ConfigHandler) DeleteBaseConfig(c *gin.Context) {
	id := c.Param("id")
	configID, err := strconv.Atoi(id)
	if err != nil {
		utils.BadRequest(c, "无效的配置ID")
		return
	}

	err = h.configService.DeleteBaseConfig(configID)
	if err != nil {
		utils.InternalServerError(c, "删除基础配置失败")
		return
	}

	utils.Success(c, nil, "基础配置删除成功")
}

// GetCommonConfigs 获取所有通用配置
func (h *ConfigHandler) GetCommonConfigs(c *gin.Context) {
	configs, err := h.configService.GetCommonConfigs()
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
func (h *ConfigHandler) CreateCommonConfig(c *gin.Context) {
	var config models.CommonConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}

	err := h.configService.CreateCommonConfig(&config)
	if err != nil {
		utils.InternalServerError(c, "创建通用配置失败")
		return
	}

	utils.Created(c, gin.H{"data": config}, "通用配置创建成功")
}

// UpdateCommonConfig 更新通用配置
func (h *ConfigHandler) UpdateCommonConfig(c *gin.Context) {
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

	err = h.configService.UpdateCommonConfig(configID, &config)
	if err != nil {
		utils.InternalServerError(c, "更新通用配置失败")
		return
	}

	utils.Success(c, gin.H{"data": config}, "通用配置更新成功")
}

// DeleteCommonConfig 删除通用配置
func (h *ConfigHandler) DeleteCommonConfig(c *gin.Context) {
	id := c.Param("id")
	configID, err := strconv.Atoi(id)
	if err != nil {
		utils.BadRequest(c, "无效的配置ID")
		return
	}

	err = h.configService.DeleteCommonConfig(configID)
	if err != nil {
		utils.InternalServerError(c, "删除通用配置失败")
		return
	}

	utils.Success(c, nil, "通用配置删除成功")
}

// GetPayConfigs 获取所有支付配置
func (h *ConfigHandler) GetPayConfigs(c *gin.Context) {
	configs, err := h.configService.GetPayConfigs()
	if err != nil {
		utils.InternalServerError(c, "获取支付配置列表失败")
		return
	}

	utils.Success(c, gin.H{
		"data":  configs,
		"total": len(configs),
	}, "获取支付配置列表成功")
}

// CreatePayConfig 创建支付配置
func (h *ConfigHandler) CreatePayConfig(c *gin.Context) {
	var config models.PayConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}

	err := h.configService.CreatePayConfig(&config)
	if err != nil {
		utils.InternalServerError(c, "创建支付配置失败")
		return
	}

	utils.Created(c, gin.H{"data": config}, "支付配置创建成功")
}

// UpdatePayConfig 更新支付配置
func (h *ConfigHandler) UpdatePayConfig(c *gin.Context) {
	id := c.Param("id")
	configID, err := strconv.Atoi(id)
	if err != nil {
		utils.BadRequest(c, "无效的配置ID")
		return
	}

	var config models.PayConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}

	err = h.configService.UpdatePayConfig(configID, &config)
	if err != nil {
		utils.InternalServerError(c, "更新支付配置失败")
		return
	}

	utils.Success(c, gin.H{"data": config}, "支付配置更新成功")
}

// DeletePayConfig 删除支付配置
func (h *ConfigHandler) DeletePayConfig(c *gin.Context) {
	id := c.Param("id")
	configID, err := strconv.Atoi(id)
	if err != nil {
		utils.BadRequest(c, "无效的配置ID")
		return
	}

	err = h.configService.DeletePayConfig(configID)
	if err != nil {
		utils.InternalServerError(c, "删除支付配置失败")
		return
	}

	utils.Success(c, nil, "支付配置删除成功")
}

// GetUIConfigs 获取所有UI配置
func (h *ConfigHandler) GetUIConfigs(c *gin.Context) {
	configs, err := h.configService.GetUIConfigs()
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
func (h *ConfigHandler) CreateUIConfig(c *gin.Context) {
	var config models.UIConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}

	err := h.configService.CreateUIConfig(&config)
	if err != nil {
		utils.InternalServerError(c, "创建UI配置失败")
		return
	}

	utils.Created(c, gin.H{"data": config}, "UI配置创建成功")
}

// UpdateUIConfig 更新UI配置
func (h *ConfigHandler) UpdateUIConfig(c *gin.Context) {
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

	err = h.configService.UpdateUIConfig(configID, &config)
	if err != nil {
		utils.InternalServerError(c, "更新UI配置失败")
		return
	}

	utils.Success(c, gin.H{"data": config}, "UI配置更新成功")
}

// DeleteUIConfig 删除UI配置
func (h *ConfigHandler) DeleteUIConfig(c *gin.Context) {
	id := c.Param("id")
	configID, err := strconv.Atoi(id)
	if err != nil {
		utils.BadRequest(c, "无效的配置ID")
		return
	}

	err = h.configService.DeleteUIConfig(configID)
	if err != nil {
		utils.InternalServerError(c, "删除UI配置失败")
		return
	}

	utils.Success(c, nil, "UI配置删除成功")
}

// GetWebsiteConfig 获取网站配置
func (h *ConfigHandler) GetWebsiteConfig(c *gin.Context) {
	clientID := c.Param("clientId")
	if clientID == "" {
		utils.BadRequest(c, "客户端ID不能为空")
		return
	}

	clientIDInt, err := strconv.Atoi(clientID)
	if err != nil {
		utils.BadRequest(c, "无效的客户端ID")
		return
	}

	config, err := h.configService.GetWebsiteConfig(clientIDInt)
	if err != nil {
		utils.NotFound(c, "网站配置不存在")
		return
	}

	utils.Success(c, gin.H{
		"message": "Website config retrieved successfully",
		"data":    config,
	}, "获取网站配置成功")
}
