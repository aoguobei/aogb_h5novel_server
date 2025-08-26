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

		// 安全重置
		git.POST("/reset", gitHandler.SafeReset)

		// 重置分支到远程状态
		git.POST("/reset-branch", gitHandler.ResetBranch)

		// 在指定代码库创建新分支
		git.POST("/pull-branch", gitHandler.PullBranchFromRepository)

	}
}
