package handlers

import (
	"brand-config-api/services"
	"brand-config-api/types"
	"brand-config-api/utils"

	"github.com/gin-gonic/gin"
)

// GitHandler Git操作控制器
type GitHandler struct {
	gitService *services.GitService
}

// NewGitHandler 创建Git控制器
func NewGitHandler() *GitHandler {
	return &GitHandler{
		gitService: services.NewGitService(),
	}
}

// CommitCode 提交代码
func (h *GitHandler) CommitCode(c *gin.Context) {
	var req types.GitCommitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// base_path 现在是可选参数，为空时使用当前目录

	// 执行Git操作
	result := h.gitService.ExecuteGitCommit(&req)

	if result.Success {
		utils.Success(c, result, "代码提交成功")
	} else {
		utils.InternalServerError(c, result.Error)
	}
}

// GetGitStatus 获取Git状态
func (h *GitHandler) GetGitStatus(c *gin.Context) {
	basePath := c.Query("base_path") // 可选参数，为空时使用当前目录

	status, err := h.gitService.GetGitStatus(basePath)
	if err != nil {
		utils.InternalServerError(c, "获取Git状态失败: "+err.Error())
		return
	}

	utils.Success(c, status, "获取Git状态成功")
}

// SafeReset 安全重置Git工作区
func (h *GitHandler) SafeReset(c *gin.Context) {
	var req struct {
		BasePath string `json:"base_path"` // 可选参数，为空时使用当前目录
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// 执行安全重置
	result := h.gitService.SafeGitReset(req.BasePath)

	if result.Success {
		utils.Success(c, result, "Git重置操作成功")
	} else {
		utils.InternalServerError(c, result.Error)
	}
}

// ResetBranch 重置分支到远程分支状态
func (h *GitHandler) ResetBranch(c *gin.Context) {
	var req struct {
		RemoteName string `json:"remote_name"`
		BranchName string `json:"branch_name"`
		BasePath   string `json:"base_path"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// 执行分支重置
	result := h.gitService.ResetBranchToRemote(req.BasePath, req.RemoteName, req.BranchName)

	if result.Success {
		utils.Success(c, result, "分支重置操作成功")
	} else {
		utils.InternalServerError(c, result.Error)
	}
}

// PullCode 拉取代码
func (h *GitHandler) PullCode(c *gin.Context) {
	var req struct {
		RemoteName string `json:"remote_name"`
		BranchName string `json:"branch_name"`
		BasePath   string `json:"base_path"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// 设置默认值
	if req.RemoteName == "" {
		req.RemoteName = "origin"
	}
	if req.BranchName == "" {
		req.BranchName = "main"
	}

	// 执行Git拉取操作
	result := h.gitService.PullCode(req.BasePath, req.RemoteName, req.BranchName)

	if result.Success {
		utils.Success(c, result, "代码拉取成功")
	} else {
		utils.InternalServerError(c, result.Error)
	}
}

// PullBranchFromRepository 在指定代码库创建新分支
func (h *GitHandler) PullBranchFromRepository(c *gin.Context) {
	var req types.GitPullBranchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	// 执行在指定代码库创建分支操作
	result, err := h.gitService.PullBranch(req.RepositoryURL, req.BranchName, "master")

	if err != nil {
		utils.InternalServerError(c, err.Error())
		return
	}

	if result.Success {
		utils.Success(c, result, "分支创建成功")
	} else {
		utils.InternalServerError(c, result.Message)
	}
}
