package utils

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocketManager WebSocket连接管理器
type WebSocketManager struct {
	connections  map[string]*websocket.Conn
	messageQueue map[string][]interface{} // 消息队列，存储未发送的消息
	mutex        sync.RWMutex
}

// NewWebSocketManager 创建新的WebSocket管理器
func NewWebSocketManager() *WebSocketManager {
	return &WebSocketManager{
		connections:  make(map[string]*websocket.Conn),
		messageQueue: make(map[string][]interface{}),
	}
}

// AddConnection 添加连接
func (wm *WebSocketManager) AddConnection(taskID string, conn *websocket.Conn) {
	wm.mutex.Lock()
	defer wm.mutex.Unlock()
	wm.connections[taskID] = conn
	log.Printf("WebSocket连接已添加: %s", taskID)

	// 发送队列中的消息
	if messages, exists := wm.messageQueue[taskID]; exists {
		log.Printf("发送队列中的 %d 条消息给任务 %s", len(messages), taskID)
		for _, msg := range messages {
			wm.sendMessageToConnection(taskID, conn, msg)
		}
		// 清空队列
		delete(wm.messageQueue, taskID)
	}
}

// RemoveConnection 移除连接
func (wm *WebSocketManager) RemoveConnection(taskID string) {
	wm.mutex.Lock()
	defer wm.mutex.Unlock()
	if conn, exists := wm.connections[taskID]; exists {
		conn.Close()
		delete(wm.connections, taskID)
		log.Printf("WebSocket连接已移除: %s", taskID)
	}
}

// SendMessage 发送消息到指定任务
func (wm *WebSocketManager) SendMessage(taskID string, message interface{}) {
	wm.mutex.RLock()
	conn, exists := wm.connections[taskID]
	wm.mutex.RUnlock()

	if !exists {
		// 连接不存在，将消息加入队列
		wm.mutex.Lock()
		if wm.messageQueue[taskID] == nil {
			wm.messageQueue[taskID] = make([]interface{}, 0)
		}
		wm.messageQueue[taskID] = append(wm.messageQueue[taskID], message)
		wm.mutex.Unlock()
		log.Printf("📥 消息已加入队列 [%s]，等待连接建立", taskID)
		return
	}

	wm.sendMessageToConnection(taskID, conn, message)
}

// sendMessageToConnection 发送消息到指定连接
func (wm *WebSocketManager) sendMessageToConnection(taskID string, conn *websocket.Conn, message interface{}) {
	// 设置写入超时
	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

	// 序列化消息
	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("消息序列化失败: %v", err)
		return
	}

	// 发送消息
	log.Printf("📤 发送WebSocket消息 [%s]: %s", taskID, string(data))
	err = conn.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		log.Printf("发送WebSocket消息失败: %v", err)
		wm.RemoveConnection(taskID)
	} else {
		log.Printf("✅ WebSocket消息发送成功 [%s]", taskID)
	}
}

// BroadcastMessage 广播消息到所有连接
func (wm *WebSocketManager) BroadcastMessage(message interface{}) {
	wm.mutex.RLock()
	defer wm.mutex.RUnlock()

	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("消息序列化失败: %v", err)
		return
	}

	for taskID, conn := range wm.connections {
		conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
		err := conn.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			log.Printf("广播消息失败 [%s]: %v", taskID, err)
			conn.Close()
			delete(wm.connections, taskID)
		}
	}
}

// GetConnectionCount 获取连接数量
func (wm *WebSocketManager) GetConnectionCount() int {
	wm.mutex.RLock()
	defer wm.mutex.RUnlock()
	return len(wm.connections)
}
