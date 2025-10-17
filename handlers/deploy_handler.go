package handlers

import (
	"brand-config-api/services"
	"brand-config-api/utils"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// DeployHandler 部署处理器
type DeployHandler struct {
	deployService *services.DeployService
	wsManager     *utils.WebSocketManager
	taskManager   *utils.TaskManager
}

// NewDeployHandler 创建部署处理器
func NewDeployHandler(wsManager *utils.WebSocketManager, taskManager *utils.TaskManager) *DeployHandler {
	return &DeployHandler{
		deployService: services.NewDeployService(),
		wsManager:     wsManager,
		taskManager:   taskManager,
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

	// 创建任务
	task, err := h.taskManager.CreateTask()
	if err != nil {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"error": "系统繁忙，请稍后重试",
		})
		return
	}

	// 返回任务ID给前端
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"taskId":  task.ID,
		"message": "部署任务已创建，请通过WebSocket连接获取实时进度",
	})

	// 启动远程部署（异步）
	go func() {
		defer func() {
			h.taskManager.CompleteTask(task.ID, "部署任务结束")
		}()

		// 更新任务状态为运行中
		h.taskManager.StartTask(task.ID, "开始远程部署...")

		// 创建WebSocket输出适配器
		wsOutputChan := make(chan services.OutputMessage, 100)

		// 启动WebSocket消息转发器
		go h.forwardMessagesToWebSocket(task.ID, wsOutputChan)

		// 执行部署脚本
		if err := h.deployService.ExecuteDeployScriptWithStream(config, wsOutputChan); err != nil {
			h.taskManager.FailTask(task.ID, fmt.Sprintf("远程部署失败: %v", err))
			wsOutputChan <- services.OutputMessage{
				Type:    "failed",
				Message: fmt.Sprintf("远程部署失败: %v", err),
			}
		} else {
			h.taskManager.CompleteTask(task.ID, "远程部署成功完成")
		}

		close(wsOutputChan)
	}()
}

// forwardMessagesToWebSocket 将输出消息转发到WebSocket
func (h *DeployHandler) forwardMessagesToWebSocket(taskID string, outputChan <-chan services.OutputMessage) {
	for msg := range outputChan {
		// 发送到WebSocket
		h.wsManager.SendMessage(taskID, gin.H{
			"type": "deploy_output",
			"data": msg,
		})

		// 同时更新任务进度（根据消息类型）
		switch msg.Type {
		case "output":
			// 普通输出，可以根据内容判断进度
			h.taskManager.UpdateTaskProgress(taskID, 50, msg.Message) // 简单的进度估算
		case "success":
			h.taskManager.UpdateTaskProgress(taskID, 100, msg.Message)
		case "failed":
			// 失败消息会在主流程中处理
		case "error":
			// 错误消息，但不一定是失败
			h.taskManager.UpdateTaskProgress(taskID, 50, "遇到错误: "+msg.Message)
		}
	}
}
