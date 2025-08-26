package utils

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// GitStatus Git状态信息
type GitStatus struct {
	Branch     string   `json:"branch"`
	Status     string   `json:"status"`
	Staged     []string `json:"staged"`
	Modified   []string `json:"modified"`
	Untracked  []string `json:"untracked"`
	StashCount int      `json:"stash_count"`
	Ahead      int      `json:"ahead"`
	Behind     int      `json:"behind"`
}

// GitOperationDetail Git操作详情
type GitOperationDetail struct {
	Operation string `json:"operation"`
	Status    string `json:"status"`
	Message   string `json:"message"`
	Output    string `json:"output"`
	Duration  int64  `json:"duration"`
}

// ValidateGitEnvironment 统一验证Git环境
// 整合了：validateGitPath + checkGitAvailability + checkNetworkConnectivity + checkDirectoryPermissions
func ValidateGitEnvironment(gitReposDir string) error {
	// 检查Git仓库目录是否存在
	if gitReposDir == "" {
		return fmt.Errorf("Git仓库目录未配置")
	}

	// 检查目录是否存在，如果不存在则创建
	if _, err := os.Stat(gitReposDir); os.IsNotExist(err) {
		if err := os.MkdirAll(gitReposDir, 0755); err != nil {
			return fmt.Errorf("无法创建Git仓库目录 %s: %v", gitReposDir, err)
		}
		log.Printf("创建Git仓库目录: %s", gitReposDir)
	}

	// 检查目录权限
	if err := checkDirectoryPermissions(gitReposDir); err != nil {
		return fmt.Errorf("Git仓库目录权限检查失败: %v", err)
	}

	// 检查Git是否可用
	if err := checkGitAvailability(); err != nil {
		return fmt.Errorf("Git命令不可用: %v", err)
	}

	// 检查网络连接
	if err := checkNetworkConnectivity(); err != nil {
		return fmt.Errorf("网络连接检查失败: %v", err)
	}

	return nil
}

// GetGitStatus 统一获取Git状态信息
// 整合了：getCurrentBranch + getStashCount + getBranchStatus + getWorkingDirectoryStatus
func GetGitStatus(basePath string) (*GitStatus, error) {
	status := &GitStatus{}

	// 获取当前分支
	if branch, err := getCurrentBranch(basePath); err == nil {
		status.Branch = branch
	}

	// 获取工作区状态
	if err := getWorkingDirectoryStatus(basePath, status); err != nil {
		return nil, err
	}

	// 获取暂存数量
	if stashCount, err := getStashCount(basePath); err == nil {
		status.StashCount = stashCount
	}

	// 获取分支领先/落后信息
	if ahead, behind, err := getBranchStatus(basePath); err == nil {
		status.Ahead = ahead
		status.Behind = behind
	}

	return status, nil
}

// ExecuteGitCommand 统一执行Git命令
// 整合了：executeGitCommand + executeGitCommandWithDetail
func ExecuteGitCommand(basePath, operationName, command string, args []string, description string) GitOperationDetail {
	startTime := time.Now()
	detail := GitOperationDetail{
		Operation: operationName,
		Status:    "success",
		Message:   description,
	}

	// 创建命令
	cmd := exec.Command(command, args...)
	cmd.Dir = basePath
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROGRESS=0")

	// 执行命令
	output, err := cmd.CombinedOutput()
	duration := time.Since(startTime).Milliseconds()
	detail.Duration = duration

	if err != nil {
		detail.Status = "error"
		detail.Message = fmt.Sprintf("%s 失败: %v %s", description, err, string(output))
		detail.Output = string(output)
		log.Printf("Git命令执行失败: %s %v, 输出: %s", command, args, string(output))
	} else {
		detail.Status = "success"
		detail.Message = fmt.Sprintf("%s 成功", description)
		detail.Output = strings.TrimSpace(string(output))
		log.Printf("Git命令执行成功: %s %v, 输出: %s", command, args, string(output))
	}

	return detail
}

// IsGitRepository 检查指定路径是否为Git仓库
func IsGitRepository(path string) bool {
	gitDir := filepath.Join(path, ".git")
	info, err := os.Stat(gitDir)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// GenerateRepoName 根据仓库URL生成唯一的目录名
func GenerateRepoName(repositoryURL string) string {
	// 移除协议前缀
	url := strings.TrimPrefix(repositoryURL, "http://")
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "git://")

	// 移除.git后缀
	url = strings.TrimSuffix(url, ".git")

	// 替换特殊字符
	url = strings.ReplaceAll(url, "/", "_")
	url = strings.ReplaceAll(url, ":", "_")
	url = strings.ReplaceAll(url, ".", "_")

	// 限制长度
	if len(url) > 50 {
		url = url[:50]
	}

	return url
}

// ParseGitStatusOutput 解析git status --porcelain的输出
func ParseGitStatusOutput(output string) (*GitStatus, error) {
	status := &GitStatus{}

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// 解析状态行
		if len(line) >= 2 {
			statusCode := line[:2]
			fileName := strings.TrimSpace(line[2:])

			switch {
			case strings.Contains(statusCode, "A"), strings.Contains(statusCode, "M"):
				status.Staged = append(status.Staged, fileName)
			case strings.Contains(statusCode, "M"), strings.Contains(statusCode, "D"):
				status.Modified = append(status.Modified, fileName)
			case strings.Contains(statusCode, "??"):
				status.Untracked = append(status.Untracked, fileName)
			}
		}
	}

	// 设置总体状态
	if len(status.Staged) > 0 || len(status.Modified) > 0 || len(status.Untracked) > 0 {
		status.Status = "dirty"
	} else {
		status.Status = "clean"
	}

	return status, nil
}

// ==================== 私有辅助函数 ====================

// checkDirectoryPermissions 检查目录权限
func checkDirectoryPermissions(dir string) error {
	// 尝试创建测试文件
	testFile := filepath.Join(dir, ".test_permission")
	file, err := os.Create(testFile)
	if err != nil {
		return fmt.Errorf("无法在目录 %s 中创建文件: %v", dir, err)
	}
	file.Close()

	// 删除测试文件
	os.Remove(testFile)
	return nil
}

// checkGitAvailability 检查Git是否可用
func checkGitAvailability() error {
	cmd := exec.Command("git", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Git命令不可用: %v", err)
	}
	return nil
}

// checkNetworkConnectivity 检查网络连接
func checkNetworkConnectivity() error {
	// 简单检查网络连接，使用通用的DNS服务器
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 使用通用的DNS服务器进行网络检查
	// 简化处理：直接使用Windows格式，在Linux/Mac上也能工作
	cmd := exec.CommandContext(ctx, "ping", "-n", "1", "8.8.8.8")

	if err := cmd.Run(); err != nil {
		log.Printf("网络连接检查失败: 无法ping通8.8.8.8")
		return fmt.Errorf("网络连接检查失败")
	}

	log.Printf("网络连接检查通过")
	return nil
}

// getCurrentBranch 获取当前分支
func getCurrentBranch(basePath string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = basePath
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// getWorkingDirectoryStatus 获取工作区状态
func getWorkingDirectoryStatus(basePath string, status *GitStatus) error {
	// 获取状态
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = basePath
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// 解析状态行
		if len(line) >= 2 {
			statusCode := line[:2]
			fileName := strings.TrimSpace(line[2:])

			switch {
			case strings.Contains(statusCode, "A"), strings.Contains(statusCode, "M"):
				status.Staged = append(status.Staged, fileName)
			case strings.Contains(statusCode, "M"), strings.Contains(statusCode, "D"):
				status.Modified = append(status.Modified, fileName)
			case strings.Contains(statusCode, "??"):
				status.Untracked = append(status.Untracked, fileName)
			}
		}
	}

	// 设置总体状态
	if len(status.Staged) > 0 || len(status.Modified) > 0 || len(status.Untracked) > 0 {
		status.Status = "dirty"
	} else {
		status.Status = "clean"
	}

	return nil
}

// getStashCount 获取暂存数量
func getStashCount(basePath string) (int, error) {
	cmd := exec.Command("git", "stash", "list")
	cmd.Dir = basePath
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return 0, nil
	}
	return len(lines), nil
}

// getBranchStatus 获取分支领先/落后信息
func getBranchStatus(basePath string) (ahead int, behind int, err error) {
	// 先获取当前分支
	branch, err := getCurrentBranch(basePath)
	if err != nil {
		return 0, 0, err
	}

	// 获取远程分支信息
	cmd := exec.Command("git", "fetch", "origin")
	cmd.Dir = basePath
	if err := cmd.Run(); err != nil {
		return 0, 0, fmt.Errorf("获取远程信息失败: %v", err)
	}

	// 检查领先的提交数
	cmd = exec.Command("git", "rev-list", "--count", fmt.Sprintf("origin/%s..HEAD", branch))
	cmd.Dir = basePath
	output, err := cmd.Output()
	if err != nil {
		return 0, 0, fmt.Errorf("检查领先提交数失败: %v", err)
	}
	aheadStr := strings.TrimSpace(string(output))
	if aheadStr != "" {
		if ahead, err = strconv.Atoi(aheadStr); err != nil {
			return 0, 0, fmt.Errorf("解析领先提交数失败: %v", err)
		}
	}

	// 检查落后的提交数
	cmd = exec.Command("git", "rev-list", "--count", fmt.Sprintf("HEAD..origin/%s", branch))
	cmd.Dir = basePath
	output, err = cmd.Output()
	if err != nil {
		return 0, 0, fmt.Errorf("检查落后提交数失败: %v", err)
	}
	behindStr := strings.TrimSpace(string(output))
	if behindStr != "" {
		if behind, err = strconv.Atoi(behindStr); err != nil {
			return 0, 0, fmt.Errorf("解析落后提交数失败: %v", err)
		}
	}

	return ahead, behind, nil
}

// CheckLocalBranchExists 检查本地分支是否存在
func CheckLocalBranchExists(localPath, branchName string) bool {
	cmd := exec.Command("git", "branch", "--list", branchName)
	cmd.Dir = localPath
	output, err := cmd.Output()
	if err != nil {
		log.Printf("检查本地分支失败: %v", err)
		return false
	}
	return len(strings.TrimSpace(string(output))) > 0
}

// CheckRemoteBranchExists 检查远程分支是否存在
func CheckRemoteBranchExists(localPath, branchName string) bool {
	// 先fetch获取最新的远程分支信息
	fetchCmd := exec.Command("git", "fetch", "origin")
	fetchCmd.Dir = localPath
	fetchCmd.Run() // 忽略fetch错误，继续检查

	cmd := exec.Command("git", "branch", "-r", "--list", "origin/"+branchName)
	cmd.Dir = localPath
	output, err := cmd.Output()
	if err != nil {
		log.Printf("检查远程分支失败: %v", err)
		return false
	}
	return len(strings.TrimSpace(string(output))) > 0
}
