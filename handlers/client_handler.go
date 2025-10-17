package handlers

import (
	"strconv"

	"brand-config-api/models"
	"brand-config-api/services"
	"brand-config-api/utils"

	"github.com/gin-gonic/gin"
)

// ClientHandler 客户端控制器
type ClientHandler struct {
	clientService *services.ClientService
}

// NewClientHandler 创建客户端控制器
func NewClientHandler() *ClientHandler {
	return &ClientHandler{
		clientService: services.NewClientService(),
	}
}

// GetClients 获取所有客户端
func (h *ClientHandler) GetClients(c *gin.Context) {
	clients, err := h.clientService.GetAllClients()
	if err != nil {
		utils.InternalServerError(c, "获取客户端列表失败")
		return
	}

	utils.Success(c, gin.H{
		"data":  clients,
		"total": len(clients),
	}, "获取客户端列表成功")
}

// GetClient 获取单个客户端
func (h *ClientHandler) GetClient(c *gin.Context) {
	id := c.Param("id")
	clientID, err := strconv.Atoi(id)
	if err != nil {
		utils.BadRequest(c, "无效的客户端ID")
		return
	}

	client, err := h.clientService.GetClientByID(clientID)
	if err != nil {
		utils.NotFound(c, "客户端不存在")
		return
	}

	utils.Success(c, gin.H{"data": client}, "获取客户端成功")
}

// CreateClient 创建客户端
func (h *ClientHandler) CreateClient(c *gin.Context) {
	var client models.Client
	if err := c.ShouldBindJSON(&client); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}

	createdClient, err := h.clientService.CreateClient(client.BrandID, client.Host)
	if err != nil {
		utils.Conflict(c, err.Error())
		return
	}

	utils.Created(c, gin.H{"data": createdClient}, "客户端创建成功")
}

// UpdateClient 更新客户端
func (h *ClientHandler) UpdateClient(c *gin.Context) {
	id := c.Param("id")
	clientID, err := strconv.Atoi(id)
	if err != nil {
		utils.BadRequest(c, "无效的客户端ID")
		return
	}

	var client models.Client
	if err := c.ShouldBindJSON(&client); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}

	updatedClient, err := h.clientService.UpdateClient(clientID, client.BrandID, client.Host)
	if err != nil {
		utils.Conflict(c, err.Error())
		return
	}

	utils.Success(c, gin.H{"data": updatedClient}, "客户端更新成功")
}

// DeleteClient 删除客户端
func (h *ClientHandler) DeleteClient(c *gin.Context) {
	id := c.Param("id")
	clientID, err := strconv.Atoi(id)
	if err != nil {
		utils.BadRequest(c, "无效的客户端ID")
		return
	}

	err = h.clientService.DeleteClient(clientID)
	if err != nil {
		utils.Conflict(c, err.Error())
		return
	}

	utils.Success(c, nil, "客户端删除成功")
}
