package handlers

import (
	"brand-config-api/services"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// DeployHandler 部署处理器
type DeployHandler struct {
	deployService *services.DeployService
}

// NewDeployHandler 创建部署处理器
func NewDeployHandler() *DeployHandler {
	return &DeployHandler{
		deployService: services.NewDeployService(),
	}
}

// DeployNginx 部署nginx配置
func (h *DeployHandler) DeployNginx(c *gin.Context) {
	var config services.NginxDeployConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的请求数据: " + err.Error(),
		})
		return
	}

	// 设置流式响应
	c.Header("Content-Type", "application/json")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Status(http.StatusOK)

	// 创建输出通道
	outputChan := make(chan services.OutputMessage, 100)

	// 获取响应写入器
	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "流式响应不支持",
		})
		return
	}

	// 启动远程部署
	go func() {
		defer close(outputChan)
		if err := h.deployService.ExecuteDeployScriptWithStream(config, outputChan); err != nil {
			outputChan <- services.OutputMessage{
				Type:    "failed",
				Message: fmt.Sprintf("远程部署失败: %v", err),
			}
		}
	}()

	// 流式输出结果
	for msg := range outputChan {
		data, _ := json.Marshal(msg)
		c.Writer.Write(data)
		c.Writer.Write([]byte("\n"))
		flusher.Flush()
	}
}

// DeployLocal 本地部署
func (h *DeployHandler) DeployLocal(c *gin.Context) {
	var config services.LocalDeployConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的请求数据: " + err.Error(),
		})
		return
	}

	// 设置流式响应
	c.Header("Content-Type", "application/json")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Status(http.StatusOK)

	// 创建输出通道
	outputChan := make(chan services.OutputMessage, 100)

	// 获取响应写入器
	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "流式响应不支持",
		})
		return
	}

	// 启动本地部署
	go func() {
		defer close(outputChan)
		if err := h.deployService.ExecuteLocalScript(config, outputChan); err != nil {
			outputChan <- services.OutputMessage{
				Type:    "failed",
				Message: fmt.Sprintf("本地部署失败: %v", err),
			}
		}
	}()

	// 流式输出结果
	for msg := range outputChan {
		data, _ := json.Marshal(msg)
		c.Writer.Write(data)
		c.Writer.Write([]byte("\n"))
		flusher.Flush()
	}
}
