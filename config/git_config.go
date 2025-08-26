package config

import (
	"os"
	"path/filepath"
)

// GitConfig Git操作配置
type GitConfig struct {
	DefaultBasePath  string // 默认Git仓库路径
	DefaultRemote    string // 默认远程仓库名称
	DefaultTargetRef string // 默认目标引用
}

// GetGitConfig 获取Git配置
func GetGitConfig() *GitConfig {
	// 使用config.go中的basePath/funNovel
	appConfig := Load()
	projectRoot := appConfig.File.ProjectRoot

	// 检查项目根目录是否为Git仓库
	gitPath := filepath.Join(projectRoot, ".git")
	if _, err := os.Stat(gitPath); err == nil {
		return &GitConfig{
			DefaultBasePath:  projectRoot,
			DefaultRemote:    "origin",
			DefaultTargetRef: "HEAD:refs/for/uni/funNovel/devNew",
		}
	}

	// 如果项目根目录不是Git仓库，尝试在basePath目录
	basePath := appConfig.File.BasePath
	gitPath = filepath.Join(basePath, ".git")
	if _, err := os.Stat(gitPath); err == nil {
		return &GitConfig{
			DefaultBasePath:  basePath,
			DefaultRemote:    "origin",
			DefaultTargetRef: "HEAD:refs/for/uni/funNovel/devNew",
		}
	}

	// 如果都找不到，返回项目根目录
	return &GitConfig{
		DefaultBasePath:  projectRoot,
		DefaultRemote:    "origin",
		DefaultTargetRef: "HEAD:refs/for/uni/funNovel/devNew",
	}
}
