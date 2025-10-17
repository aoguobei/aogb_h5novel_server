package handlers

import (
	"brand-config-api/services"
	"brand-config-api/utils"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

// BuildHandler 构建处理器
type BuildHandler struct {
	wsManager   *utils.WebSocketManager
	taskManager *utils.TaskManager
}

// NewBuildHandler 创建构建处理器
func NewBuildHandler(wsManager *utils.WebSocketManager, taskManager *utils.TaskManager) *BuildHandler {
	return &BuildHandler{
		wsManager:   wsManager,
		taskManager: taskManager,
	}
}

// H5BuildConfig H5项目构建配置
type H5BuildConfig struct {
	Branch          string   `json:"branch" binding:"required"`
	Version         string   `json:"version" binding:"required"`
	Environments    []string `json:"environments" binding:"required"`
	Projects        []string `json:"projects" binding:"required"`
	ForceForeignNet bool     `json:"forceForeignNet"`
	SSHHost         string   `json:"sshHost" binding:"required"`
	SSHUser         string   `json:"sshUser" binding:"required"`
	SSHPassword     string   `json:"sshPassword" binding:"required"`
}

// BuildH5 构建H5项目
func (h *BuildHandler) BuildH5(c *gin.Context) {
	var config H5BuildConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		log.Printf("❌ 绑定构建配置失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的构建配置: " + err.Error(),
		})
		return
	}

	// 验证配置
	if err := h.validateBuildConfig(config); err != nil {
		log.Printf("❌ 构建配置验证失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// 创建任务
	task, err := h.taskManager.CreateTask()
	if err != nil {
		log.Printf("❌ 创建任务失败: %v", err)
		c.JSON(http.StatusTooManyRequests, gin.H{
			"success": false,
			"error":   "任务创建失败: " + err.Error(),
		})
		return
	}
	taskID := task.ID
	log.Printf("🏗️ 创建H5构建任务: %s", taskID)

	// 立即返回任务ID
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"taskId":  taskID,
		"message": "构建任务已创建，请通过WebSocket连接获取实时进度",
	})

	// 异步执行构建
	go func() {
		defer func() {
			h.taskManager.CompleteTask(taskID, "构建任务结束")
		}()

		// 开始任务
		h.taskManager.StartTask(taskID, "开始H5项目构建...")

		buildService := services.NewBuildService()

		// 构建批量构建请求
		batchReq := &services.BatchBuildRequest{
			Projects:     config.Projects,
			Branch:       config.Branch,
			Version:      config.Version,
			Environment:  strings.Join(config.Environments, ","), // 将环境数组转为逗号分隔字符串
			ForceForeign: config.ForceForeignNet,
			SSHHost:      config.SSHHost,
			SSHUser:      config.SSHUser,
			SSHPassword:  config.SSHPassword,
		}

		// 执行批量构建
		_, err := buildService.ExecuteBatchBuild(batchReq, func(progress services.BuildProgress) {
			// 发送进度到WebSocket
			h.wsManager.SendMessage(taskID, map[string]interface{}{
				"type": "deploy_output",
				"data": map[string]interface{}{
					"type":    progress.Status,
					"message": progress.Output,
				},
			})

			// 发送任务进度
			h.wsManager.SendMessage(taskID, map[string]interface{}{
				"type": "task_status",
				"data": map[string]interface{}{
					"status":   progress.Status,
					"progress": progress.Percentage,
					"message":  progress.Text,
					"details":  progress.Detail,
					"project":  progress.Project,
				},
			})
		})

		// 发送任务完成状态
		if err != nil {
			h.taskManager.FailTask(taskID, err.Error())
			h.wsManager.SendMessage(taskID, map[string]interface{}{
				"type": "task_status",
				"data": map[string]interface{}{
					"status":  "failed",
					"message": "构建失败",
					"error":   err.Error(),
				},
			})
		} else {
			h.taskManager.CompleteTask(taskID, "H5项目构建成功完成")
			h.wsManager.SendMessage(taskID, map[string]interface{}{
				"type": "task_status",
				"data": map[string]interface{}{
					"status":  "completed",
					"message": "H5项目构建成功完成",
				},
			})
		}
	}()
}

// validateBuildConfig 验证构建配置
func (h *BuildHandler) validateBuildConfig(config H5BuildConfig) error {
	// 验证分支名称
	if strings.TrimSpace(config.Branch) == "" {
		return fmt.Errorf("分支名称不能为空")
	}

	// 验证版本号格式
	if !regexp.MustCompile(`^\d+\.\d+\.\d+$`).MatchString(config.Version) {
		return fmt.Errorf("版本号格式错误，应为 x.x.x")
	}

	// 验证环境选择
	validEnvs := map[string]bool{"master": true, "release": true, "local": true}
	if len(config.Environments) == 0 {
		return fmt.Errorf("必须选择至少一个环境")
	}
	for _, env := range config.Environments {
		if !validEnvs[env] {
			return fmt.Errorf("无效的环境: %s", env)
		}
	}

	// 验证项目选择 - 只检查是否为空，不限制具体项目
	if len(config.Projects) == 0 {
		return fmt.Errorf("必须选择至少一个项目")
	}

	// 验证SSH配置
	if strings.TrimSpace(config.SSHHost) == "" {
		return fmt.Errorf("SSH主机不能为空")
	}
	if strings.TrimSpace(config.SSHUser) == "" {
		return fmt.Errorf("SSH用户名不能为空")
	}
	if strings.TrimSpace(config.SSHPassword) == "" {
		return fmt.Errorf("SSH密码不能为空")
	}

	return nil
}

// GetTaskStatus 获取任务状态
func (h *BuildHandler) GetTaskStatus(c *gin.Context) {
	taskID := c.Param("taskId")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "任务ID不能为空",
		})
		return
	}

	task, exists := h.taskManager.GetTask(taskID)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "任务不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    task,
	})
}

// handleBuildOutput 处理构建输出
func (h *BuildHandler) handleBuildOutput(taskID string, outputChan <-chan services.OutputMessage) {
	for output := range outputChan {
		// 发送构建输出到WebSocket
		h.wsManager.SendMessage(taskID, map[string]interface{}{
			"type": "deploy_output",
			"data": output,
		})
	}
}
