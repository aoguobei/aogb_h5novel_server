package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"brand-config-api/database"
	"brand-config-api/models"

	"github.com/gin-gonic/gin"
)

// writeConfigToFile 将配置写入本地文件
func writeConfigToFile(configType string, config interface{}, brandCode string, host string) error {
	// 根据配置类型确定文件路径
	var configDir string
	var fileName string

	switch configType {
	case "common":
		configDir = "C:/F_explorer/h5projects/jianruiH5/novel_h5config/funNovel/src/appConfig/commonConfigs"
		fileName = brandCode + ".js"
	case "pay":
		configDir = "C:/F_explorer/h5projects/jianruiH5/novel_h5config/funNovel/src/appConfig/payConfigs"
		fileName = brandCode + ".js"
	case "ui":
		configDir = "C:/F_explorer/h5projects/jianruiH5/novel_h5config/funNovel/src/appConfig/uiConfigs"
		fileName = brandCode + ".js"
	default:
		return fmt.Errorf("unknown config type: %s", configType)
	}

	// 确保目录存在
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %v", configDir, err)
	}

	// 构建文件路径
	configFile := filepath.Join(configDir, fileName)

	// 检查文件是否存在
	_, err := os.Stat(configFile)
	fileExists := err == nil

	var configData map[string]interface{}

	if fileExists {
		// 读取现有文件
		content, err := os.ReadFile(configFile)
		if err != nil {
			return fmt.Errorf("failed to read existing config file: %v", err)
		}

		// 解析现有配置（移除export default前缀）
		contentStr := string(content)
		if strings.HasPrefix(contentStr, "export default ") {
			contentStr = strings.TrimPrefix(contentStr, "export default ")
		}

		// 将单引号替换为双引号，使其符合JSON格式
		contentStr = strings.ReplaceAll(contentStr, "'", `"`)

		if err := json.Unmarshal([]byte(contentStr), &configData); err != nil {
			return fmt.Errorf("failed to parse existing config file: %v", err)
		}
	} else {
		// 创建新的配置对象
		configData = make(map[string]interface{})
	}

	// 使用host作为key，添加配置
	configData[host] = config

	// 转换为JSON并添加export default前缀
	configJSON, err := json.MarshalIndent(configData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config to JSON: %v", err)
	}

	content := fmt.Sprintf("export default %s\n", string(configJSON))

	if err := os.WriteFile(configFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write config file %s: %v", configFile, err)
	}

	fmt.Printf("✅ %s config written to: %s with host key: %s\n", configType, configFile, host)
	return nil
}

// formatCommonConfig 格式化CommonConfig为标准格式
func formatCommonConfig(config models.CommonConfig) map[string]interface{} {
	return map[string]interface{}{
		"deliver": map[string]interface{}{
			"business_id": map[string]interface{}{
				"value":  config.DeliverBusinessID,
				"enable": config.DeliverBusinessIDEnable,
			},
			"switch_id": map[string]interface{}{
				"value":  config.DeliverSwitchID,
				"enable": config.DeliverSwitchIDEnable,
			},
		},
		"protocol": map[string]interface{}{
			"company":    config.ProtocolCompany,
			"about":      config.ProtocolAbout,
			"privacy":    config.ProtocolPrivacy,
			"vod":        config.ProtocolVod,
			"userCancel": config.ProtocolUserCancel,
		},
		"contact": config.ContactURL,
		"script": map[string]interface{}{
			"base": config.ScriptBase,
		},
	}
}

// formatPayConfig 格式化PayConfig为标准格式
func formatPayConfig(config models.PayConfig) map[string]interface{} {
	return map[string]interface{}{
		"normal_pay": map[string]interface{}{
			"enable": config.NormalPayEnable,
			"gateway_id": map[string]interface{}{
				"android": config.NormalPayGatewayAndroid,
				"ios":     config.NormalPayGatewayIOS,
			},
		},
		"renew_pay": map[string]interface{}{
			"enable": config.RenewPayEnable,
			"gateway_id": map[string]interface{}{
				"android": config.RenewPayGatewayAndroid,
				"ios":     config.RenewPayGatewayIOS,
			},
		},
	}
}

// formatUIConfig 格式化UIConfig为标准格式
func formatUIConfig(config models.UIConfig) map[string]interface{} {
	result := map[string]interface{}{
		"bgStyle": map[string]interface{}{
			"main":   config.ThemeBgMain,
			"second": config.ThemeBgSecond,
		},
	}

	if config.ThemeTextMain != nil && *config.ThemeTextMain != "" {
		result["textColor"] = map[string]interface{}{
			"main": *config.ThemeTextMain,
		}
	}

	return result
}

// BaseConfig 相关处理器

// GetBaseConfigs 获取基础配置列表
func GetBaseConfigs(c *gin.Context) {
	var configs []models.BaseConfig

	result := database.DB.Preload("Brand").Find(&configs)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  configs,
		"total": len(configs),
	})
}

// GetBaseConfig 获取单个基础配置
func GetBaseConfig(c *gin.Context) {
	id := c.Param("id")
	configID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid config ID"})
		return
	}

	var config models.BaseConfig
	result := database.DB.Preload("Brand").First(&config, configID)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Config not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": config})
}

// CreateBaseConfig 创建基础配置
func CreateBaseConfig(c *gin.Context) {
	var config models.BaseConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查是否已存在相同client_id的配置
	var existingConfig models.BaseConfig
	if err := database.DB.Where("client_id = ?", config.ClientID).First(&existingConfig).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Configuration for this client already exists"})
		return
	}

	result := database.DB.Create(&config)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	// 获取关联的Client和Brand信息
	var client models.Client
	if err := database.DB.Preload("Brand").First(&client, config.ClientID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get client info"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": config})
}

// UpdateBaseConfig 更新基础配置
func UpdateBaseConfig(c *gin.Context) {
	id := c.Param("id")
	configID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid config ID"})
		return
	}

	var config models.BaseConfig
	if err := database.DB.First(&config, configID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Config not found"})
		return
	}

	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result := database.DB.Save(&config)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	// 获取关联的Client和Brand信息
	var client models.Client
	if err := database.DB.Preload("Brand").First(&client, config.ClientID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get client info"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": config})
}

// DeleteBaseConfig 删除基础配置
func DeleteBaseConfig(c *gin.Context) {
	id := c.Param("id")
	configID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid config ID"})
		return
	}

	result := database.DB.Delete(&models.BaseConfig{}, configID)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Config not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Config deleted successfully"})
}

// CommonConfig 相关处理器

// GetCommonConfigs 获取通用配置列表
func GetCommonConfigs(c *gin.Context) {
	var configs []models.CommonConfig

	result := database.DB.Preload("Brand").Find(&configs)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  configs,
		"total": len(configs),
	})
}

// CreateCommonConfig 创建通用配置
func CreateCommonConfig(c *gin.Context) {
	var config models.CommonConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查是否已存在相同client_id的配置
	var existingConfig models.CommonConfig
	if err := database.DB.Where("client_id = ?", config.ClientID).First(&existingConfig).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Configuration for this client already exists"})
		return
	}

	result := database.DB.Create(&config)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	// 获取关联的Client和Brand信息
	var client models.Client
	if err := database.DB.Preload("Brand").First(&client, config.ClientID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get client info"})
		return
	}

	// 格式化配置并写入本地文件
	formattedConfig := formatCommonConfig(config)
	if err := writeConfigToFile("common", formattedConfig, client.Brand.Code, client.Host); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write config file: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": config})
}

// UpdateCommonConfig 更新通用配置
func UpdateCommonConfig(c *gin.Context) {
	id := c.Param("id")
	configID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid config ID"})
		return
	}

	var config models.CommonConfig
	if err := database.DB.First(&config, configID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Config not found"})
		return
	}

	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result := database.DB.Save(&config)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	// 获取关联的Client和Brand信息
	var client models.Client
	if err := database.DB.Preload("Brand").First(&client, config.ClientID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get client info"})
		return
	}

	// 格式化配置并写入本地文件
	formattedConfig := formatCommonConfig(config)
	if err := writeConfigToFile("common", formattedConfig, client.Brand.Code, client.Host); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write config file: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": config})
}

// DeleteCommonConfig 删除通用配置
func DeleteCommonConfig(c *gin.Context) {
	id := c.Param("id")
	configID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid config ID"})
		return
	}

	result := database.DB.Delete(&models.CommonConfig{}, configID)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Config not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Config deleted successfully"})
}

// PayConfig 相关处理器

// GetPayConfigs 获取支付配置列表
func GetPayConfigs(c *gin.Context) {
	var configs []models.PayConfig

	query := database.DB.Preload("Client")

	// 如果提供了client_id参数，则按client_id过滤
	if clientID := c.Query("client_id"); clientID != "" {
		query = query.Where("client_id = ?", clientID)
	}

	result := query.Find(&configs)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  configs,
		"total": len(configs),
	})
}

// CreatePayConfig 创建支付配置
func CreatePayConfig(c *gin.Context) {
	var config models.PayConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查是否已存在相同client_id的配置
	var existingConfig models.PayConfig
	if err := database.DB.Where("client_id = ?", config.ClientID).First(&existingConfig).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Configuration for this client already exists"})
		return
	}

	result := database.DB.Create(&config)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	// 获取关联的Client和Brand信息
	var client models.Client
	if err := database.DB.Preload("Brand").First(&client, config.ClientID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get client info"})
		return
	}

	// 格式化配置并写入本地文件
	formattedConfig := formatPayConfig(config)
	if err := writeConfigToFile("pay", formattedConfig, client.Brand.Code, client.Host); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write config file: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": config})
}

// UpdatePayConfig 更新支付配置
func UpdatePayConfig(c *gin.Context) {
	id := c.Param("id")
	configID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid config ID"})
		return
	}

	var config models.PayConfig
	if err := database.DB.First(&config, configID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Config not found"})
		return
	}

	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result := database.DB.Save(&config)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	// 获取关联的Client和Brand信息
	var client models.Client
	if err := database.DB.Preload("Brand").First(&client, config.ClientID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get client info"})
		return
	}

	// 格式化配置并写入本地文件
	formattedConfig := formatPayConfig(config)
	if err := writeConfigToFile("pay", formattedConfig, client.Brand.Code, client.Host); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write config file: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": config})
}

// DeletePayConfig 删除支付配置
func DeletePayConfig(c *gin.Context) {
	id := c.Param("id")
	configID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid config ID"})
		return
	}

	result := database.DB.Delete(&models.PayConfig{}, configID)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Config not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Config deleted successfully"})
}

// UIConfig 相关处理器

// GetUIConfigs 获取UI配置列表
func GetUIConfigs(c *gin.Context) {
	var configs []models.UIConfig

	query := database.DB.Preload("Client")

	// 如果提供了client_id参数，则按client_id过滤
	if clientID := c.Query("client_id"); clientID != "" {
		query = query.Where("client_id = ?", clientID)
	}

	result := query.Find(&configs)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  configs,
		"total": len(configs),
	})
}

// CreateUIConfig 创建UI配置
func CreateUIConfig(c *gin.Context) {
	var config models.UIConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查是否已存在相同client_id的配置
	var existingConfig models.UIConfig
	if err := database.DB.Where("client_id = ?", config.ClientID).First(&existingConfig).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Configuration for this client already exists"})
		return
	}

	result := database.DB.Create(&config)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	// 获取关联的Client和Brand信息
	var client models.Client
	if err := database.DB.Preload("Brand").First(&client, config.ClientID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get client info"})
		return
	}

	// 格式化配置并写入本地文件
	formattedConfig := formatUIConfig(config)
	if err := writeConfigToFile("ui", formattedConfig, client.Brand.Code, client.Host); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write config file: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": config})
}

// UpdateUIConfig 更新UI配置
func UpdateUIConfig(c *gin.Context) {
	id := c.Param("id")
	configID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid config ID"})
		return
	}

	var config models.UIConfig
	if err := database.DB.First(&config, configID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Config not found"})
		return
	}

	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result := database.DB.Save(&config)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	// 获取关联的Client和Brand信息
	var client models.Client
	if err := database.DB.Preload("Brand").First(&client, config.ClientID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get client info"})
		return
	}

	// 格式化配置并写入本地文件
	formattedConfig := formatUIConfig(config)
	if err := writeConfigToFile("ui", formattedConfig, client.Brand.Code, client.Host); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write config file: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": config})
}

// DeleteUIConfig 删除UI配置
func DeleteUIConfig(c *gin.Context) {
	id := c.Param("id")
	configID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid config ID"})
		return
	}

	result := database.DB.Delete(&models.UIConfig{}, configID)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Config not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Config deleted successfully"})
}
