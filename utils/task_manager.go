package utils

import (
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

// TaskStatus 任务状态
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
)

// Task 任务信息
type Task struct {
	ID        string     `json:"id"`
	Status    TaskStatus `json:"status"`
	Progress  int        `json:"progress"`
	Message   string     `json:"message"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	Error     string     `json:"error,omitempty"`
}

// TaskManager 任务管理器
type TaskManager struct {
	tasks map[string]*Task
	mutex sync.RWMutex
	// 限制同时运行的任务数量
	maxConcurrentTasks int32
	currentTasks       int32
}

// NewTaskManager 创建新的任务管理器
func NewTaskManager(maxConcurrent int32) *TaskManager {
	return &TaskManager{
		tasks:              make(map[string]*Task),
		maxConcurrentTasks: maxConcurrent,
	}
}

// CreateTask 创建新任务
func (tm *TaskManager) CreateTask() (*Task, error) {
	// 检查是否超过最大并发数
	if atomic.LoadInt32(&tm.currentTasks) >= tm.maxConcurrentTasks {
		return nil, ErrTooManyTasks
	}

	taskID := uuid.New().String()
	task := &Task{
		ID:        taskID,
		Status:    TaskStatusPending,
		Progress:  0,
		Message:   "任务已创建，等待执行...",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	tm.mutex.Lock()
	tm.tasks[taskID] = task
	tm.mutex.Unlock()

	atomic.AddInt32(&tm.currentTasks, 1)
	log.Printf("任务已创建: %s", taskID)

	return task, nil
}

// GetTask 获取任务信息
func (tm *TaskManager) GetTask(taskID string) (*Task, bool) {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()
	task, exists := tm.tasks[taskID]
	return task, exists
}

// UpdateTaskProgress 更新任务进度
func (tm *TaskManager) UpdateTaskProgress(taskID string, progress int, message string) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	if task, exists := tm.tasks[taskID]; exists {
		task.Progress = progress
		task.Message = message
		task.UpdatedAt = time.Now()
		log.Printf("任务进度更新 [%s]: %d%% - %s", taskID, progress, message)
	}
}

// CompleteTask 完成任务
func (tm *TaskManager) CompleteTask(taskID string, message string) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	if task, exists := tm.tasks[taskID]; exists {
		task.Status = TaskStatusCompleted
		task.Progress = 100
		task.Message = message
		task.UpdatedAt = time.Now()
		atomic.AddInt32(&tm.currentTasks, -1)
		log.Printf("任务已完成: %s", taskID)
	}
}

// FailTask 任务失败
func (tm *TaskManager) FailTask(taskID string, error string) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	if task, exists := tm.tasks[taskID]; exists {
		task.Status = TaskStatusFailed
		task.Message = "任务执行失败"
		task.Error = error
		task.UpdatedAt = time.Now()
		atomic.AddInt32(&tm.currentTasks, -1)
		log.Printf("任务失败: %s - %s", taskID, error)
	}
}

// StartTask 开始任务
func (tm *TaskManager) StartTask(taskID string, message ...string) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	if task, exists := tm.tasks[taskID]; exists {
		task.Status = TaskStatusRunning

		// 如果提供了消息参数，使用它；否则使用默认消息
		if len(message) > 0 && message[0] != "" {
			task.Message = message[0]
		} else {
			task.Message = "任务正在执行..."
		}

		task.UpdatedAt = time.Now()
		log.Printf("任务已开始: %s - %s", taskID, task.Message)
	}
}

// CleanupOldTasks 清理旧任务（可选，用于内存管理）
func (tm *TaskManager) CleanupOldTasks(maxAge time.Duration) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	now := time.Now()
	for taskID, task := range tm.tasks {
		if now.Sub(task.UpdatedAt) > maxAge {
			delete(tm.tasks, taskID)
			log.Printf("清理旧任务: %s", taskID)
		}
	}
}

// GetCurrentTaskCount 获取当前任务数量
func (tm *TaskManager) GetCurrentTaskCount() int32 {
	return atomic.LoadInt32(&tm.currentTasks)
}

// GetMaxConcurrentTasks 获取最大并发任务数
func (tm *TaskManager) GetMaxConcurrentTasks() int32 {
	return tm.maxConcurrentTasks
}
