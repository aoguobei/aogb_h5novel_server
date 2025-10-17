package handlers

import (
	"brand-config-api/utils"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有来源，生产环境应该更严格
	},
}

// WebSocketHandler WebSocket处理器
type WebSocketHandler struct {
	wsManager   *utils.WebSocketManager
	taskManager *utils.TaskManager
}

// NewWebSocketHandler 创建新的WebSocket处理器
func NewWebSocketHandler(wsManager *utils.WebSocketManager, taskManager *utils.TaskManager) *WebSocketHandler {
	return &WebSocketHandler{
		wsManager:   wsManager,
		taskManager: taskManager,
	}
}

// HandleWebSocket 处理WebSocket连接
func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	taskID := c.Query("taskId")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "taskId is required"})
		return
	}

	// 检查任务是否存在
	task, exists := h.taskManager.GetTask(taskID)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	// 升级HTTP连接为WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket升级失败: %v", err)
		return
	}

	// 添加连接到管理器
	h.wsManager.AddConnection(taskID, conn)

	// 发送初始任务状态
	h.wsManager.SendMessage(taskID, gin.H{
		"type": "task_status",
		"data": task,
	})

	log.Printf("WebSocket连接已建立: %s", taskID)

	// 处理连接关闭
	defer func() {
		h.wsManager.RemoveConnection(taskID)
		log.Printf("WebSocket连接已关闭: %s", taskID)
	}()

	// 保持连接活跃
	for {
		// 读取消息（这里主要是为了检测连接是否断开）
		_, _, err := conn.ReadMessage()
		if err != nil {
			log.Printf("WebSocket读取错误: %v", err)
			break
		}
	}
}
