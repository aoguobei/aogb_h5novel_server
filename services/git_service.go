package services

import (
	"brand-config-api/config"
	"brand-config-api/types"
	"brand-config-api/utils"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// GitService Git操作服务
type GitService struct {
	config *config.Config
}

// NewGitService 创建Git服务实例
func NewGitService() *GitService {
	return &GitService{
		config: config.Load(),
	}
}

// ExecuteGitCommit 执行完整的Git提交流程
func (s *GitService) ExecuteGitCommit(req *types.GitCommitRequest) *types.GitOperationResult {
	result := &types.GitOperationResult{
		Success: false,
		Details: []types.GitOperationDetail{},
	}

	// 设置默认值
	gitConfig := config.GetGitConfig()
	if req.BasePath == "" {
		// 如果没有指定路径，使用config.go中的basePath/funNovel
		appConfig := config.Load()
		req.BasePath = appConfig.File.ProjectRoot // 使用 basePath/funNovel
	}
	if req.RemoteName == "" {
		req.RemoteName = gitConfig.DefaultRemote
	}
	if req.TargetRef == "" {
		req.TargetRef = gitConfig.DefaultTargetRef
	}
	if req.CommitMsg == "" {
		req.CommitMsg = "feat: 网站配置更新 - " + time.Now().Format("2006-01-02 15:04:05")
	}

	// 验证Git环境
	if err := utils.ValidateGitEnvironment(s.config.File.ProjectRoot); err != nil {
		result.Error = fmt.Sprintf("Git环境验证失败: %v", err)
		return result
	}

	// 检查工作区状态，决定是否需要stash
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = req.BasePath
	output, err := cmd.Output()
	hasLocalChanges := err == nil && len(strings.TrimSpace(string(output))) > 0

	// 执行Git操作流程
	operations := []struct {
		name    string
		command string
		args    []string
		desc    string
		skipIf  func() bool // 可选的跳过条件
	}{
		{"git_config_email", "git", []string{"config", "user.email", "webauto@example.com"}, "配置Git用户邮箱", nil},
		{"git_config_name", "git", []string{"config", "user.name", "webauto"}, "配置Git用户名", nil},
		{"git_stash", "git", []string{"stash"}, "暂存当前工作区修改", func() bool {
			return !hasLocalChanges // 如果没有本地修改，跳过stash
		}},
		{"git_reset", "git", []string{"reset", "--hard", "HEAD"}, "重置工作区到最新提交", nil},
		{"git_pull", "git", []string{"pull", req.RemoteName, "HEAD"}, "拉取最新代码", nil},
		{"git_stash_pop", "git", []string{"stash", "pop"}, "恢复暂存的修改", func() bool {
			return !hasLocalChanges // 如果没有本地修改（没有执行stash），跳过stash pop
		}},
		{"git_add", "git", []string{"add", "."}, "添加所有修改到暂存区", nil},
		{"git_commit", "git", []string{"commit", "-m", req.CommitMsg}, "提交代码", func() bool {
			// 检查是否有需要提交的修改
			cmd := exec.Command("git", "status", "--porcelain")
			cmd.Dir = req.BasePath
			output, err := cmd.Output()
			if err != nil {
				return false // 如果检查失败，不跳过
			}
			return len(strings.TrimSpace(string(output))) == 0 // 如果没有修改，跳过commit
		}},
		{"git_push", "git", []string{"push", req.RemoteName, req.TargetRef}, "推送到远程仓库", nil},
	}

	// 执行每个操作
	for _, op := range operations {
		// 检查是否需要跳过此操作
		if op.skipIf != nil && op.skipIf() {
			detail := types.GitOperationDetail{
				Operation: op.name,
				Status:    "skipped",
				Message:   fmt.Sprintf("%s (跳过：没有需要提交的修改)", op.desc),
				Duration:  0,
			}
			result.Details = append(result.Details, detail)
			continue
		}

		detail := utils.ExecuteGitCommand(req.BasePath, op.name, op.command, op.args, op.desc)
		// 转换为types.GitOperationDetail
		typesDetail := types.GitOperationDetail{
			Operation: detail.Operation,
			Status:    detail.Status,
			Message:   detail.Message,
			Output:    detail.Output,
			Duration:  detail.Duration,
		}
		result.Details = append(result.Details, typesDetail)

		// 如果某个操作失败，停止执行
		if detail.Status == "error" {
			result.Error = fmt.Sprintf("操作 %s 失败: %s", op.desc, detail.Message)
			return result
		}
	}

	result.Success = true
	result.Message = "Git操作执行成功"
	return result
}

// GetGitStatus 获取Git状态信息
func (s *GitService) GetGitStatus(basePath string) (*types.GitStatus, error) {
	// 如果没有指定路径，使用config.go中的basePath/funNovel
	if basePath == "" {
		appConfig := config.Load()
		basePath = appConfig.File.ProjectRoot // 使用 basePath/funNovel
	}

	// 使用工具类获取Git状态
	gitStatus, err := utils.GetGitStatus(basePath)
	if err != nil {
		return nil, err
	}

	// 转换为types.GitStatus
	status := &types.GitStatus{
		Branch:     gitStatus.Branch,
		Status:     gitStatus.Status,
		Staged:     gitStatus.Staged,
		Modified:   gitStatus.Modified,
		Untracked:  gitStatus.Untracked,
		StashCount: gitStatus.StashCount,
		Ahead:      gitStatus.Ahead,
		Behind:     gitStatus.Behind,
	}

	return status, nil
}

// ResetToRemote 重置到远程分支
func (s *GitService) ResetToRemote(basePath, remoteName, branchName string) *types.GitOperationResult {
	result := &types.GitOperationResult{
		Success: false,
		Details: []types.GitOperationDetail{},
	}

	// 如果没有指定路径，使用config.go中的basePath/funNovel
	if basePath == "" {
		appConfig := config.Load()
		basePath = appConfig.File.ProjectRoot // 使用 basePath/funNovel
	}

	// 设置默认值
	if remoteName == "" {
		remoteName = "origin"
	}
	if branchName == "" {
		branchName = "uni/funNovel/devNew"
	}

	// 验证Git环境
	if err := utils.ValidateGitEnvironment(s.config.File.ProjectRoot); err != nil {
		result.Error = fmt.Sprintf("Git环境验证失败: %v", err)
		return result
	}

	// 先执行git fetch更新远程分支引用
	fetchDetail := utils.ExecuteGitCommand(basePath, "git_fetch", "git", []string{"fetch", remoteName}, "更新远程分支引用")

	// 转换为types.GitOperationDetail
	fetchTypesDetail := types.GitOperationDetail{
		Operation: fetchDetail.Operation,
		Status:    fetchDetail.Status,
		Message:   fetchDetail.Message,
		Output:    fetchDetail.Output,
		Duration:  fetchDetail.Duration,
	}
	result.Details = append(result.Details, fetchTypesDetail)

	// 如果fetch失败，返回错误
	if fetchDetail.Status == "error" {
		result.Error = fmt.Sprintf("更新远程分支引用失败: %s", fetchDetail.Message)
		return result
	}

	// 执行重置操作
	detail := utils.ExecuteGitCommand(basePath, "git_reset_hard", "git", []string{"reset", "--hard", fmt.Sprintf("origin/%s", branchName)}, "重置到远程分支")

	// 转换为types.GitOperationDetail
	typesDetail := types.GitOperationDetail{
		Operation: detail.Operation,
		Status:    detail.Status,
		Message:   detail.Message,
		Output:    detail.Output,
		Duration:  detail.Duration,
	}
	result.Details = append(result.Details, typesDetail)

	// 如果重置失败，返回错误
	if detail.Status == "error" {
		result.Error = fmt.Sprintf("重置失败: %s", detail.Message)
		return result
	}

	result.Success = true
	result.Message = "重置操作执行成功"
	return result
}

// PullCode 执行Git拉取操作
func (s *GitService) PullCode(basePath, remoteName, branchName string) *types.GitOperationResult {
	result := &types.GitOperationResult{
		Success: false,
		Details: []types.GitOperationDetail{},
	}

	// 如果没有指定路径，使用config.go中的basePath/funNovel
	if basePath == "" {
		appConfig := config.Load()
		basePath = appConfig.File.ProjectRoot // 使用 basePath/funNovel
	}

	// 验证Git环境
	if err := utils.ValidateGitEnvironment(s.config.File.ProjectRoot); err != nil {
		result.Error = fmt.Sprintf("Git环境验证失败: %v", err)
		return result
	}

	// 检查工作区状态，如果是dirty则不允许拉取
	gitStatus, err := utils.GetGitStatus(basePath)
	if err != nil {
		result.Error = fmt.Sprintf("获取Git状态失败: %v", err)
		return result
	}

	if gitStatus.Status == "dirty" {
		result.Error = "工作区有未提交的修改，请先提交或恢复修改后再拉取代码"
		result.Details = append(result.Details, types.GitOperationDetail{
			Operation: "status_check",
			Status:    "error",
			Message:   "工作区状态检查失败：存在未提交的修改",
			Output:    fmt.Sprintf("暂存文件: %v, 修改文件: %v, 未跟踪文件: %v", gitStatus.Staged, gitStatus.Modified, gitStatus.Untracked),
			Duration:  0,
		})
		return result
	}

	// 记录状态检查成功
	result.Details = append(result.Details, types.GitOperationDetail{
		Operation: "status_check",
		Status:    "success",
		Message:   "工作区状态检查通过：工作区干净",
		Output:    "",
		Duration:  0,
	})

	// 简化操作：只执行git pull
	detail := utils.ExecuteGitCommand(basePath, "git_pull", "git", []string{"pull", remoteName, branchName}, "从远程拉取最新代码")

	// 转换为types.GitOperationDetail
	typesDetail := types.GitOperationDetail{
		Operation: detail.Operation,
		Status:    detail.Status,
		Message:   detail.Message,
		Output:    detail.Output,
		Duration:  detail.Duration,
	}
	result.Details = append(result.Details, typesDetail)

	// 如果拉取失败，返回错误
	if detail.Status == "error" {
		result.Error = fmt.Sprintf("拉取代码失败: %s", detail.Message)
		return result
	}

	result.Success = true
	result.Message = "Git拉取操作执行成功"
	return result
}

// PullBranch 在远程仓库创建新分支（直接推送创建远程分支）
func (s *GitService) PullBranch(repositoryURL, branchName, defaultBranch string) (*types.GitOperationResult, error) {
	result := &types.GitOperationResult{
		Success: false,
		Message: "开始创建远程分支",
		Details: []types.GitOperationDetail{},
	}

	log.Printf("开始创建分支: 仓库=%s, 分支=%s", repositoryURL, branchName)

	// 验证配置
	configStartTime := time.Now()
	if err := utils.ValidateGitEnvironment(s.config.GitReposDir); err != nil {
		result.Details = append(result.Details, types.GitOperationDetail{
			Operation: "config_validation",
			Status:    "error",
			Message:   "配置验证失败: " + err.Error(),
			Output:    "",
			Duration:  int64(time.Since(configStartTime).Milliseconds()),
		})
		result.Message = "配置验证失败"
		log.Printf("配置验证失败: %v", err)
		return result, err
	}

	// 记录配置验证成功
	result.Details = append(result.Details, types.GitOperationDetail{
		Operation: "config_validation",
		Status:    "success",
		Message:   "配置验证成功",
		Output:    "",
		Duration:  int64(time.Since(configStartTime).Milliseconds()),
	})

	// 生成本地路径
	localPath := filepath.Join(s.config.GitReposDir, utils.GenerateRepoName(repositoryURL))
	log.Printf("本地路径: %s", localPath)

	// 准备仓库（克隆或更新）
	if err := s.prepareRepository(repositoryURL, localPath, result); err != nil {
		log.Printf("准备仓库失败: %v", err)
		return result, err
	}

	// 只检查远程分支是否已存在
	remoteBranchExistsBeforeOperation := utils.CheckRemoteBranchExists(localPath, branchName)

	log.Printf("分支状态检查: 远程分支=%t", remoteBranchExistsBeforeOperation)

	// 如果远程分支已存在，直接返回成功
	if remoteBranchExistsBeforeOperation {
		result.Success = true
		result.Message = fmt.Sprintf("分支 %s 已存在，无需重复创建", branchName)
		result.Details = append(result.Details, types.GitOperationDetail{
			Operation: "branch_exists_check",
			Status:    "success",
			Message:   fmt.Sprintf("远程分支 %s 已存在", branchName),
			Output:    "",
			Duration:  0,
		})
		log.Printf("远程分支 %s 已存在，跳过创建", branchName)
		return result, nil
	}

	// 直接推送创建远程分支
	if err := s.createRemoteBranchDirectly(localPath, branchName, defaultBranch, result); err != nil {
		log.Printf("创建远程分支失败: %v", err)
		// 创建失败，尝试回滚
		s.rollbackRemoteBranchCreation(localPath, branchName, result, remoteBranchExistsBeforeOperation)
		return result, err
	}

	result.Success = true
	result.Message = "成功创建远程分支 " + branchName
	log.Printf("成功创建远程分支: %s", branchName)
	return result, nil
}

// prepareRepository 准备仓库（使用ExecuteGitCommand）
func (s *GitService) prepareRepository(repositoryURL, localPath string, result *types.GitOperationResult) error {
	log.Printf("准备仓库: URL=%s, 本地路径=%s", repositoryURL, localPath)

	// 检查本地路径是否存在
	if _, err := os.Stat(localPath); os.IsNotExist(err) {
		log.Printf("本地路径不存在，进行克隆")
		return s.cloneRepository(repositoryURL, localPath, result)
	}

	log.Printf("本地路径已存在，检查是否为Git仓库")

	// 检查是否为Git仓库
	if !utils.IsGitRepository(localPath) {
		log.Printf("不是Git仓库，删除后重新克隆")
		removeStartTime := time.Now()
		if err := os.RemoveAll(localPath); err != nil {
			log.Printf("删除非Git目录失败: %v", err)
			result.Details = append(result.Details, types.GitOperationDetail{
				Operation: "clean_local_path",
				Status:    "error",
				Message:   "删除本地路径失败: " + err.Error(),
				Output:    "",
				Duration:  int64(time.Since(removeStartTime).Milliseconds()),
			})
			return err
		}

		// 记录清理成功
		result.Details = append(result.Details, types.GitOperationDetail{
			Operation: "clean_local_path",
			Status:    "success",
			Message:   "成功清理本地路径",
			Output:    "",
			Duration:  int64(time.Since(removeStartTime).Milliseconds()),
		})

		return s.cloneRepository(repositoryURL, localPath, result)
	}

	log.Printf("Git仓库已存在，进行更新")
	return s.updateRepository(localPath, result)
}

// cloneRepository 克隆仓库（使用ExecuteGitCommand）
func (s *GitService) cloneRepository(repositoryURL, localPath string, result *types.GitOperationResult) error {
	log.Printf("开始克隆仓库: %s", repositoryURL)

	// 创建本地目录
	createStartTime := time.Now()
	if err := os.MkdirAll(localPath, 0755); err != nil {
		log.Printf("创建本地目录失败: %v", err)
		result.Details = append(result.Details, types.GitOperationDetail{
			Operation: "create_directory",
			Status:    "error",
			Message:   "创建本地目录失败: " + err.Error(),
			Output:    "",
			Duration:  int64(time.Since(createStartTime).Milliseconds()),
		})
		return err
	}

	// 记录创建目录成功
	result.Details = append(result.Details, types.GitOperationDetail{
		Operation: "create_directory",
		Status:    "success",
		Message:   "成功创建本地目录",
		Output:    "",
		Duration:  int64(time.Since(createStartTime).Milliseconds()),
	})

	// 使用ExecuteGitCommand克隆仓库
	detail := utils.ExecuteGitCommand("", "git_clone", "git", []string{"clone", repositoryURL, localPath}, "克隆仓库")

	// 转换为types.GitOperationDetail
	typesDetail := types.GitOperationDetail{
		Operation: detail.Operation,
		Status:    detail.Status,
		Message:   detail.Message,
		Output:    detail.Output,
		Duration:  detail.Duration,
	}
	result.Details = append(result.Details, typesDetail)

	if detail.Status == "error" {
		return fmt.Errorf("克隆仓库失败: %s", detail.Message)
	}

	return nil
}

// updateRepository 更新仓库（使用ExecuteGitCommand）
func (s *GitService) updateRepository(localPath string, result *types.GitOperationResult) error {
	log.Printf("更新仓库: %s", localPath)

	// 使用ExecuteGitCommand拉取最新代码
	detail := utils.ExecuteGitCommand(localPath, "git_pull", "git", []string{"pull", "origin"}, "拉取最新代码")

	// 转换为types.GitOperationDetail
	typesDetail := types.GitOperationDetail{
		Operation: detail.Operation,
		Status:    detail.Status,
		Message:   detail.Message,
		Output:    detail.Output,
		Duration:  detail.Duration,
	}
	result.Details = append(result.Details, typesDetail)

	if detail.Status == "error" {
		return fmt.Errorf("拉取最新代码失败: %s", detail.Message)
	}

	return nil
}

// createRemoteBranchDirectly 直接推送创建远程分支（不创建本地分支）
func (s *GitService) createRemoteBranchDirectly(localPath, branchName, defaultBranch string, result *types.GitOperationResult) error {
	log.Printf("直接推送创建远程分支: %s", branchName)

	// 使用传入的默认分支，如果为空则使用master作为兜底
	targetBranch := defaultBranch
	if targetBranch == "" {
		targetBranch = "master"
	}
	log.Printf("使用默认分支: %s", targetBranch)

	// 切换到默认分支
	checkoutDetail := utils.ExecuteGitCommand(localPath, "git_checkout_default", "git", []string{"checkout", targetBranch}, fmt.Sprintf("切换到%s分支", targetBranch))

	checkoutTypesDetail := types.GitOperationDetail{
		Operation: checkoutDetail.Operation,
		Status:    checkoutDetail.Status,
		Message:   checkoutDetail.Message,
		Output:    checkoutDetail.Output,
		Duration:  checkoutDetail.Duration,
	}
	result.Details = append(result.Details, checkoutTypesDetail)

	if checkoutDetail.Status == "error" {
		return fmt.Errorf("切换到%s分支失败: %s", targetBranch, checkoutDetail.Message)
	}

	// 核心：直接推送创建远程分支
	// 命令：git push origin master:refs/heads/new-branch
	pushArgs := []string{"push", "origin", fmt.Sprintf("%s:refs/heads/%s", targetBranch, branchName)}
	pushDetail := utils.ExecuteGitCommand(localPath, "git_push_create_remote", "git", pushArgs, "直接推送创建远程分支")

	pushTypesDetail := types.GitOperationDetail{
		Operation: pushDetail.Operation,
		Status:    pushDetail.Status,
		Message:   pushDetail.Message,
		Output:    pushDetail.Output,
		Duration:  pushDetail.Duration,
	}
	result.Details = append(result.Details, pushTypesDetail)

	if pushDetail.Status == "error" {
		return fmt.Errorf("直接推送创建远程分支失败: %s", pushDetail.Message)
	}

	log.Printf("成功：远程分支 %s 创建成功，本地仓库保持干净", branchName)
	return nil
}

// rollbackRemoteBranchCreation 回滚远程分支创建操作
func (s *GitService) rollbackRemoteBranchCreation(localPath, branchName string, result *types.GitOperationResult, remoteBranchExistedBefore bool) {
	log.Printf("开始回滚远程分支创建: 本地路径=%s, 分支=%s", localPath, branchName)

	rollbackStartTime := time.Now()

	// 记录回滚开始
	result.Details = append(result.Details, types.GitOperationDetail{
		Operation: "rollback_start",
		Status:    "info",
		Message:   "开始回滚远程分支创建操作",
		Output:    "",
		Duration:  0,
	})

	// 检查远程分支是否存在
	remoteBranchExists := utils.CheckRemoteBranchExists(localPath, branchName)

	// 如果远程分支存在且是本次操作创建的，尝试删除
	if remoteBranchExists && !remoteBranchExistedBefore {
		// 删除远程分支
		deleteRemoteDetail := utils.ExecuteGitCommand(localPath, "rollback_delete_remote_branch", "git", []string{"push", "origin", "--delete", branchName}, "回滚：删除远程分支")
		result.Details = append(result.Details, types.GitOperationDetail{
			Operation: deleteRemoteDetail.Operation,
			Status:    deleteRemoteDetail.Status,
			Message:   deleteRemoteDetail.Message,
			Output:    deleteRemoteDetail.Output,
			Duration:  deleteRemoteDetail.Duration,
		})

		if deleteRemoteDetail.Status == "success" {
			log.Printf("成功删除远程分支: %s", branchName)
		} else {
			log.Printf("删除远程分支失败: %s", deleteRemoteDetail.Message)
		}
	}

	// 记录回滚完成
	rollbackDuration := int64(time.Since(rollbackStartTime).Milliseconds())
	result.Details = append(result.Details, types.GitOperationDetail{
		Operation: "rollback_complete",
		Status:    "success",
		Message:   "远程分支创建回滚操作完成",
		Output:    "",
		Duration:  rollbackDuration,
	})

	log.Printf("远程分支创建回滚完成，耗时: %dms", rollbackDuration)
}

// GetGitLog 获取Git提交日志
func (s *GitService) GetGitLog(req *types.GitLogRequest) (*types.GitLogResponse, error) {
	// 设置默认值
	if req.Limit <= 0 {
		req.Limit = 15 // 固定返回15条记录
	}
	if req.Limit > 50 {
		req.Limit = 50 // 限制最大数量
	}

	// 使用工具类获取Git日志
	gitLog, err := utils.GetGitLog(s.config.File.ProjectRoot, req.Limit, req.FilePath)
	if err != nil {
		return nil, err
	}

	// 转换为types.GitLogResponse
	response := &types.GitLogResponse{
		Commits: make([]types.GitCommit, len(gitLog.Commits)),
	}

	for i, commit := range gitLog.Commits {
		response.Commits[i] = types.GitCommit{
			CommitId: commit.CommitId,
			Author:   commit.Author,
			Email:    commit.Email,
			Date:     commit.Date,
			Message:  commit.Message,
		}
	}

	return response, nil
}
