package handlers

import (
	"strconv"

	"brand-config-api/models"
	"brand-config-api/services"
	"brand-config-api/utils"

	"github.com/gin-gonic/gin"
)

// PayConfigHandler 支付配置处理器
type PayConfigHandler struct {
	payConfigService *services.PayConfigService
}

// NewPayConfigHandler 创建支付配置处理器实例
func NewPayConfigHandler() *PayConfigHandler {
	return &PayConfigHandler{
		payConfigService: services.NewPayConfigService(),
	}
}

// GetPayConfigs 获取所有支付配置
func (h *PayConfigHandler) GetPayConfigs(c *gin.Context) {
	configs, err := h.payConfigService.GetPayConfigs()
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
func (h *PayConfigHandler) CreatePayConfig(c *gin.Context) {
	var config models.PayConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}

	if err := h.payConfigService.CreatePayConfig(&config); err != nil {
		utils.InternalServerError(c, "创建支付配置失败")
		return
	}

	utils.Created(c, gin.H{"data": config}, "支付配置创建成功")
}

// UpdatePayConfigByClientID 根据client_id更新支付配置
func (h *PayConfigHandler) UpdatePayConfigByClientID(c *gin.Context) {
	clientID := c.Param("client_id")
	clientIDInt, err := strconv.Atoi(clientID)
	if err != nil {
		utils.BadRequest(c, "无效的客户端ID")
		return
	}

	var config models.PayConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}

	if err := h.payConfigService.UpdatePayConfigByClientID(clientIDInt, config); err != nil {
		utils.InternalServerError(c, "更新支付配置失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{"data": config}, "支付配置更新成功")
}

// DeletePayConfigByClientID 根据client_id删除支付配置
func (h *PayConfigHandler) DeletePayConfigByClientID(c *gin.Context) {
	clientIDStr := c.Param("client_id")
	clientID, err := strconv.Atoi(clientIDStr)
	if err != nil {
		utils.BadRequest(c, "无效的客户端ID")
		return
	}

	err = h.payConfigService.DeletePayConfigByClientID(clientID)
	if err != nil {
		utils.InternalServerError(c, "删除支付配置失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{
		"message":   "支付配置删除成功",
		"client_id": clientID,
	}, "支付配置删除成功")
}
