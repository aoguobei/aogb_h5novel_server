package handlers

import (
	"brand-config-api/services"
	"brand-config-api/types"
	"brand-config-api/utils"
	"log"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// WebsiteHandler 网站控制器
type WebsiteHandler struct {
	websiteService *services.WebsiteService
	wsManager      *utils.WebSocketManager
	taskManager    *utils.TaskManager
}

// NewWebsiteHandler 创建网站控制器
func NewWebsiteHandler(wsManager *utils.WebSocketManager, taskManager *utils.TaskManager) *WebsiteHandler {
	return &WebsiteHandler{
		websiteService: services.NewWebsiteService(),
		wsManager:      wsManager,
		taskManager:    taskManager,
	}
}

// CreateWebsite 创建网站
func (h *WebsiteHandler) CreateWebsite(c *gin.Context) {
	var req services.CreateWebsiteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}

	// 创建任务
	task, err := h.taskManager.CreateTask()
	if err != nil {
		utils.InternalServerError(c, "创建任务失败: "+err.Error())
		return
	}

	// 立即返回任务ID
	utils.Created(c, gin.H{
		"message": "任务已创建，正在处理...",
		"data": gin.H{
			"task_id": task.ID,
		},
	}, "任务已创建")

	// 异步执行网站创建
	go h.executeWebsiteCreation(task.ID, &req)
}

// executeWebsiteCreation 异步执行网站创建
func (h *WebsiteHandler) executeWebsiteCreation(taskID string, req *services.CreateWebsiteRequest) {
	// 开始任务
	h.taskManager.StartTask(taskID)

	// 等待一小段时间，确保前端有时间建立WebSocket连接
	time.Sleep(500 * time.Millisecond)

	h.wsManager.SendMessage(taskID, gin.H{
		"type": "progress",
		"data": gin.H{
			"percentage": 0,
			"status":     "running",
			"text":       "开始创建网站...",
			"details":    []gin.H{},
		},
	})

	// 创建进度回调函数
	progressCallback := types.ProgressCallback(func(percentage int, text string, detail string) {
		h.updateProgress(taskID, percentage, text, detail)
	})

	// 调用实际的网站创建服务（带真实进度）
	result, err := h.websiteService.CreateWebsite(req, progressCallback)

	if err != nil {
		// 任务失败
		h.taskManager.FailTask(taskID, err.Error())
		h.wsManager.SendMessage(taskID, gin.H{
			"type": "progress",
			"data": gin.H{
				"percentage": 0,
				"status":     "exception",
				"text":       "创建失败",
				"details": []gin.H{
					{"status": "error", "message": "创建失败: " + err.Error()},
				},
			},
		})
		return
	}

	// 任务完成
	h.taskManager.CompleteTask(taskID, "网站创建完成！")
	h.wsManager.SendMessage(taskID, gin.H{
		"type": "progress",
		"data": gin.H{
			"percentage": 100,
			"status":     "success",
			"text":       "网站创建完成！",
			"details": []gin.H{
				{"status": "success", "message": "恭喜！您的网站已成功创建"},
			},
		},
	})

	// 发送最终结果
	h.wsManager.SendMessage(taskID, gin.H{
		"type": "result",
		"data": result,
	})
}

// updateProgress 更新进度
func (h *WebsiteHandler) updateProgress(taskID string, percentage int, text string, detail string) {
	log.Printf("🔄 更新进度 [%s]: %d%% - %s: %s", taskID, percentage, text, detail)
	h.taskManager.UpdateTaskProgress(taskID, percentage, text)
	h.wsManager.SendMessage(taskID, gin.H{
		"type": "progress",
		"data": gin.H{
			"percentage": percentage,
			"status":     "running",
			"text":       text,
			"details": []gin.H{
				{"status": "success", "message": detail},
			},
		},
	})
}

// GetWebsiteConfig 获取网站配置
func (h *WebsiteHandler) GetWebsiteConfig(c *gin.Context) {
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

	config, err := h.websiteService.GetWebsiteConfig(clientIDInt)
	if err != nil {
		utils.NotFound(c, "网站配置不存在")
		return
	}

	utils.Success(c, gin.H{
		"message": "Website config retrieved successfully",
		"data":    config,
	}, "获取网站配置成功")
}

// DeleteWebsite 删除网站
func (h *WebsiteHandler) DeleteWebsite(c *gin.Context) {
	clientIDStr := c.Param("clientId")
	clientID, err := strconv.Atoi(clientIDStr)
	if err != nil {
		utils.BadRequest(c, "无效的客户端ID")
		return
	}

	err = h.websiteService.DeleteWebsite(clientID)
	if err != nil {
		utils.InternalServerError(c, "删除网站失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{
		"message":   "网站删除成功",
		"client_id": clientID,
	}, "网站删除成功")
}
