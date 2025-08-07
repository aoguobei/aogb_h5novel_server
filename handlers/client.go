package handlers

import (
	"net/http"
	"strconv"

	"brand-config-api/database"
	"brand-config-api/models"

	"github.com/gin-gonic/gin"
)

// GetClients 获取所有客户端
func GetClients(c *gin.Context) {
	var clients []models.Client

	result := database.DB.Preload("Brand").Find(&clients)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  clients,
		"total": len(clients),
	})
}

// GetClient 获取单个客户端
func GetClient(c *gin.Context) {
	id := c.Param("id")
	clientID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid client ID"})
		return
	}

	var client models.Client
	result := database.DB.Preload("Brand").First(&client, clientID)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Client not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": client})
}

// CreateClient 创建客户端
func CreateClient(c *gin.Context) {
	var client models.Client
	if err := c.ShouldBindJSON(&client); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证brand_id是否存在
	var brand models.Brand
	if err := database.DB.First(&brand, client.BrandID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Brand not found"})
		return
	}

	// 验证host格式
	validHosts := []string{"h5", "tth5", "ksh5"}
	hostValid := false
	for _, validHost := range validHosts {
		if client.Host == validHost {
			hostValid = true
			break
		}
	}
	if !hostValid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid host. Must be one of: h5, tth5, ksh5"})
		return
	}

	// 检查是否已存在相同的brand_id + host组合
	var existingClient models.Client
	if err := database.DB.Where("brand_id = ? AND host = ?", client.BrandID, client.Host).First(&existingClient).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Client with this brand_id and host already exists"})
		return
	}

	result := database.DB.Create(&client)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	// 重新查询以获取关联的Brand信息
	database.DB.Preload("Brand").First(&client, client.ID)

	c.JSON(http.StatusCreated, gin.H{"data": client})
}

// UpdateClient 更新客户端
func UpdateClient(c *gin.Context) {
	id := c.Param("id")
	clientID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid client ID"})
		return
	}

	var client models.Client
	if err := database.DB.First(&client, clientID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Client not found"})
		return
	}

	var updateData models.Client
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证brand_id是否存在（如果提供了新的brand_id）
	if updateData.BrandID != 0 {
		var brand models.Brand
		if err := database.DB.First(&brand, updateData.BrandID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Brand not found"})
			return
		}
		client.BrandID = updateData.BrandID
	}

	// 验证host格式（如果提供了新的host）
	if updateData.Host != "" {
		validHosts := []string{"h5", "tth5", "ksh5"}
		hostValid := false
		for _, validHost := range validHosts {
			if updateData.Host == validHost {
				hostValid = true
				break
			}
		}
		if !hostValid {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid host. Must be one of: h5, tth5, ksh5"})
			return
		}
		client.Host = updateData.Host
	}

	// 检查是否已存在相同的brand_id + host组合（排除当前记录）
	var existingClient models.Client
	if err := database.DB.Where("brand_id = ? AND host = ? AND id != ?", client.BrandID, client.Host, clientID).First(&existingClient).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Client with this brand_id and host already exists"})
		return
	}

	result := database.DB.Save(&client)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	// 重新查询以获取关联的Brand信息
	database.DB.Preload("Brand").First(&client, client.ID)

	c.JSON(http.StatusOK, gin.H{"data": client})
}

// DeleteClient 删除客户端
func DeleteClient(c *gin.Context) {
	id := c.Param("id")
	clientID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid client ID"})
		return
	}

	result := database.DB.Delete(&models.Client{}, clientID)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Client not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Client deleted successfully"})
}
