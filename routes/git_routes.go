package routes

import (
	"brand-config-api/handlers"

	"github.com/gin-gonic/gin"
)

// SetupGitRoutes 设置Git操作路由
func SetupGitRoutes(r *gin.Engine) {
	gitHandler := handlers.NewGitHandler()

	// Git操作路由组
	git := r.Group("/api/git")
	{
		// 代码提交
		git.POST("/commit", gitHandler.CommitCode)

		// 获取Git状态
		git.GET("/status", gitHandler.GetGitStatus)

		// 代码拉取
		git.POST("/pull", gitHandler.PullCode)

		// 重置到远程分支
		git.POST("/reset-to-remote", gitHandler.ResetToRemote)

		// 在指定代码库创建新分支
		git.POST("/pull-branch", gitHandler.PullBranchFromRepository)

		// 获取Git日志
		git.GET("/log", gitHandler.GetGitLog)

	}
}
