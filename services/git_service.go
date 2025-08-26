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
	if err := utils.ValidateGitEnvironment(s.config.GitReposDir); err != nil {
		result.Error = fmt.Sprintf("Git环境验证失败: %v", err)
		return result
	}

	// 执行Git操作流程
	operations := []struct {
		name    string
		command string
		args    []string
		desc    string
		skipIf  func() bool // 可选的跳过条件
	}{
		{"git_stash", "git", []string{"stash"}, "暂存当前工作区修改", nil},
		{"git_reset", "git", []string{"reset", "--hard", "HEAD"}, "重置工作区到最新提交", nil},
		{"git_pull", "git", []string{"pull", req.RemoteName, "HEAD"}, "拉取最新代码", nil},
		{"git_stash_pop", "git", []string{"stash", "pop"}, "恢复暂存的修改", nil},
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

// SafeGitReset 安全的Git重置操作
func (s *GitService) SafeGitReset(basePath string) *types.GitOperationResult {
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
	if err := utils.ValidateGitEnvironment(s.config.GitReposDir); err != nil {
		result.Error = fmt.Sprintf("Git环境验证失败: %v", err)
		return result
	}

	// 执行安全的重置操作
	operations := []struct {
		name    string
		command string
		args    []string
		desc    string
	}{
		{"git_stash", "git", []string{"stash"}, "暂存当前工作区修改"},
		{"git_reset", "git", []string{"reset", "--hard", "HEAD"}, "重置工作区到最新提交"},
		{"git_clean", "git", []string{"clean", "-fd"}, "清理未跟踪的文件和目录"},
	}

	for _, op := range operations {
		detail := utils.ExecuteGitCommand(basePath, op.name, op.command, op.args, op.desc)
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
			result.Error = fmt.Sprintf("操作 %s 失败: %s", op.desc, detail.Message)
			return result
		}
	}

	result.Success = true
	result.Message = "Git重置操作执行成功"
	return result
}

// ResetBranchToRemote 重置分支到远程分支状态（用于处理abandon后的情况）
func (s *GitService) ResetBranchToRemote(basePath, remoteName, branchName string) *types.GitOperationResult {
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
	if err := utils.ValidateGitEnvironment(s.config.GitReposDir); err != nil {
		result.Error = fmt.Sprintf("Git环境验证失败: %v", err)
		return result
	}

	// 执行分支重置操作
	operations := []struct {
		name    string
		command string
		args    []string
		desc    string
	}{
		{"git_fetch", "git", []string{"fetch", remoteName}, "获取远程分支信息"},
		{"git_stash", "git", []string{"stash"}, "暂存当前工作区修改"},
		{"git_reset_hard", "git", []string{"reset", "--hard", fmt.Sprintf("%s/%s", remoteName, branchName)}, "重置分支到远程分支状态"},
		{"git_clean", "git", []string{"clean", "-fd"}, "清理未跟踪的文件和目录"},
		{"git_stash_pop", "git", []string{"stash", "pop"}, "恢复暂存的修改"},
	}

	for _, op := range operations {
		detail := utils.ExecuteGitCommand(basePath, op.name, op.command, op.args, op.desc)
		// 转换为types.GitOperationDetail
		typesDetail := types.GitOperationDetail{
			Operation: detail.Operation,
			Status:    detail.Status,
			Message:   detail.Message,
			Output:    detail.Output,
			Duration:  detail.Duration,
		}
		result.Details = append(result.Details, typesDetail)

		// 对于stash pop操作，如果失败（比如没有stash内容），不认为是错误
		if detail.Status == "error" && op.name != "git_stash_pop" {
			result.Error = fmt.Sprintf("操作 %s 失败: %s", op.desc, detail.Message)
			return result
		}
	}

	result.Success = true
	result.Message = "分支重置操作执行成功"
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
	if err := utils.ValidateGitEnvironment(s.config.GitReposDir); err != nil {
		result.Error = fmt.Sprintf("Git环境验证失败: %v", err)
		return result
	}

	// 执行Git拉取操作流程
	operations := []struct {
		name    string
		command string
		args    []string
		desc    string
		skipIf  func() bool // 可选的跳过条件
	}{
		{"git_stash", "git", []string{"stash"}, "暂存当前工作区修改", func() bool {
			// 检查是否有需要暂存的修改
			cmd := exec.Command("git", "status", "--porcelain")
			cmd.Dir = basePath
			output, err := cmd.Output()
			if err != nil {
				return false // 如果检查失败，不跳过
			}
			return len(strings.TrimSpace(string(output))) == 0 // 如果没有修改，跳过stash
		}},
		{"git_pull", "git", []string{"pull", remoteName, branchName}, "从远程拉取最新代码", nil},
		{"git_stash_pop", "git", []string{"stash", "pop"}, "恢复暂存的修改", func() bool {
			// 检查是否有暂存的内容
			cmd := exec.Command("git", "stash", "list")
			cmd.Dir = basePath
			output, err := cmd.Output()
			if err != nil {
				return false // 如果检查失败，不跳过
			}
			return len(strings.TrimSpace(string(output))) == 0 // 如果没有暂存内容，跳过pop
		}},
	}

	// 执行每个操作
	for _, op := range operations {
		// 检查是否需要跳过此操作
		if op.skipIf != nil && op.skipIf() {
			detail := types.GitOperationDetail{
				Operation: op.name,
				Status:    "skipped",
				Message:   fmt.Sprintf("%s (跳过：无需执行)", op.desc),
				Duration:  0,
			}
			result.Details = append(result.Details, detail)
			continue
		}

		detail := utils.ExecuteGitCommand(basePath, op.name, op.command, op.args, op.desc)
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
	result.Message = "Git拉取操作执行成功"
	return result
}

// PullBranch 在远程仓库创建新分支
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
		// 准备仓库失败，不需要回滚，因为还没有创建分支
		return result, err
	}

	// 检查分支是否已存在
	branchExistsBeforeOperation := utils.CheckLocalBranchExists(localPath, branchName)
	remoteBranchExistsBeforeOperation := utils.CheckRemoteBranchExists(localPath, branchName)

	log.Printf("分支状态检查: 本地分支=%t, 远程分支=%t", branchExistsBeforeOperation, remoteBranchExistsBeforeOperation)

	// 记录操作开始时间，用于回滚判断
	operationStartTime := time.Now()

	// 如果分支已存在，直接返回成功
	if branchExistsBeforeOperation && remoteBranchExistsBeforeOperation {
		result.Success = true
		result.Message = fmt.Sprintf("分支 %s 已存在，无需重复创建", branchName)
		result.Details = append(result.Details, types.GitOperationDetail{
			Operation: "branch_exists_check",
			Status:    "success",
			Message:   fmt.Sprintf("分支 %s 已存在于本地和远程", branchName),
			Output:    "",
			Duration:  0,
		})
		log.Printf("分支 %s 已存在，跳过创建", branchName)
		return result, nil
	}

	// 创建并推送分支
	if err := s.createAndPushBranch(localPath, branchName, defaultBranch, result); err != nil {
		log.Printf("创建分支失败: %v", err)
		// 创建分支失败，需要回滚（只回滚本次操作创建的内容）
		s.rollbackBranchCreation(localPath, branchName, result, operationStartTime, branchExistsBeforeOperation, remoteBranchExistsBeforeOperation)
		return result, err
	}

	result.Success = true
	result.Message = "成功创建远程分支 " + branchName
	log.Printf("成功创建分支: %s", branchName)
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

// createAndPushBranch 创建并推送分支（使用ExecuteGitCommand）
func (s *GitService) createAndPushBranch(localPath, branchName, defaultBranch string, result *types.GitOperationResult) error {
	log.Printf("创建分支: %s", branchName)

	// 使用传入的默认分支，如果为空则使用master作为兜底
	targetBranch := defaultBranch
	if targetBranch == "" {
		targetBranch = "master"
	}
	log.Printf("使用默认分支: %s", targetBranch)

	// 使用ExecuteGitCommand切换到默认分支
	checkoutDetail := utils.ExecuteGitCommand(localPath, "git_checkout_main", "git", []string{"checkout", targetBranch}, fmt.Sprintf("切换到%s分支", targetBranch))

	// 转换为types.GitOperationDetail
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

	// 检查本地分支是否已存在
	if utils.CheckLocalBranchExists(localPath, branchName) {
		// 如果本地分支已存在，直接切换到该分支
		checkoutExistingDetail := utils.ExecuteGitCommand(localPath, "git_checkout_existing", "git", []string{"checkout", branchName}, "切换到已存在的分支")

		// 转换为types.GitOperationDetail
		checkoutExistingTypesDetail := types.GitOperationDetail{
			Operation: checkoutExistingDetail.Operation,
			Status:    checkoutExistingDetail.Status,
			Message:   checkoutExistingDetail.Message,
			Output:    checkoutExistingDetail.Output,
			Duration:  checkoutExistingDetail.Duration,
		}
		result.Details = append(result.Details, checkoutExistingTypesDetail)

		if checkoutExistingDetail.Status == "error" {
			return fmt.Errorf("切换到已存在的分支失败: %s", checkoutExistingDetail.Message)
		}
	} else {
		// 使用ExecuteGitCommand创建新分支
		createDetail := utils.ExecuteGitCommand(localPath, "git_checkout_b", "git", []string{"checkout", "-b", branchName}, "创建新分支")

		// 转换为types.GitOperationDetail
		createTypesDetail := types.GitOperationDetail{
			Operation: createDetail.Operation,
			Status:    createDetail.Status,
			Message:   createDetail.Message,
			Output:    createDetail.Output,
			Duration:  createDetail.Duration,
		}
		result.Details = append(result.Details, createTypesDetail)

		if createDetail.Status == "error" {
			return fmt.Errorf("创建新分支失败: %s", createDetail.Message)
		}
	}

	// 检查远程分支是否已存在
	if utils.CheckRemoteBranchExists(localPath, branchName) {
		// 如果远程分支已存在，设置上游分支
		setUpstreamDetail := utils.ExecuteGitCommand(localPath, "git_set_upstream", "git", []string{"branch", "--set-upstream-to=origin/" + branchName, branchName}, "设置上游分支")

		// 转换为types.GitOperationDetail
		setUpstreamTypesDetail := types.GitOperationDetail{
			Operation: setUpstreamDetail.Operation,
			Status:    setUpstreamDetail.Status,
			Message:   setUpstreamDetail.Message,
			Output:    setUpstreamDetail.Output,
			Duration:  setUpstreamDetail.Duration,
		}
		result.Details = append(result.Details, setUpstreamTypesDetail)

		if setUpstreamDetail.Status == "error" {
			return fmt.Errorf("设置上游分支失败: %s", setUpstreamDetail.Message)
		}
	} else {
		// 使用ExecuteGitCommand推送到远程
		pushDetail := utils.ExecuteGitCommand(localPath, "git_push", "git", []string{"push", "-u", "origin", branchName}, "推送分支到远程")

		// 转换为types.GitOperationDetail
		pushTypesDetail := types.GitOperationDetail{
			Operation: pushDetail.Operation,
			Status:    pushDetail.Status,
			Message:   pushDetail.Message,
			Output:    pushDetail.Output,
			Duration:  pushDetail.Duration,
		}
		result.Details = append(result.Details, pushTypesDetail)

		if pushDetail.Status == "error" {
			return fmt.Errorf("推送分支失败: %s", pushDetail.Message)
		}
	}

	return nil
}

// rollbackBranchCreation 回滚分支创建操作
func (s *GitService) rollbackBranchCreation(localPath, branchName string, result *types.GitOperationResult, operationStartTime time.Time, branchExistedBefore bool, remoteBranchExistedBefore bool) {
	log.Printf("开始回滚分支创建: 本地路径=%s, 分支=%s, 操作前本地分支=%t, 操作前远程分支=%t", localPath, branchName, branchExistedBefore, remoteBranchExistedBefore)

	rollbackStartTime := time.Now()

	// 记录回滚开始
	result.Details = append(result.Details, types.GitOperationDetail{
		Operation: "rollback_start",
		Status:    "info",
		Message:   "开始回滚分支创建操作",
		Output:    "",
		Duration:  0,
	})

	// 检查本地分支是否存在
	branchExists := utils.CheckLocalBranchExists(localPath, branchName)
	remoteBranchExists := utils.CheckRemoteBranchExists(localPath, branchName)

	// 如果本地分支存在且是本次操作创建的，尝试删除
	if branchExists && !branchExistedBefore {
		// 直接使用master作为默认分支
		defaultBranch := "master"

		checkoutDetail := utils.ExecuteGitCommand(localPath, "rollback_checkout_main", "git", []string{"checkout", defaultBranch}, fmt.Sprintf("回滚：切换到%s分支", defaultBranch))
		result.Details = append(result.Details, types.GitOperationDetail{
			Operation: checkoutDetail.Operation,
			Status:    checkoutDetail.Status,
			Message:   checkoutDetail.Message,
			Output:    checkoutDetail.Output,
			Duration:  checkoutDetail.Duration,
		})

		if checkoutDetail.Status == "success" {
			// 删除本地分支
			deleteLocalDetail := utils.ExecuteGitCommand(localPath, "rollback_delete_local_branch", "git", []string{"branch", "-D", branchName}, "回滚：删除本地分支")
			result.Details = append(result.Details, types.GitOperationDetail{
				Operation: deleteLocalDetail.Operation,
				Status:    deleteLocalDetail.Status,
				Message:   deleteLocalDetail.Message,
				Output:    deleteLocalDetail.Output,
				Duration:  deleteLocalDetail.Duration,
			})

			if deleteLocalDetail.Status == "success" {
				log.Printf("成功删除本地分支: %s", branchName)
			} else {
				log.Printf("删除本地分支失败: %s", deleteLocalDetail.Message)
			}
		} else {
			log.Printf("切换到main分支失败，无法删除本地分支: %s", checkoutDetail.Message)
		}
	}

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
		Message:   "分支创建回滚操作完成",
		Output:    "",
		Duration:  rollbackDuration,
	})

	log.Printf("分支创建回滚完成，耗时: %dms", rollbackDuration)
}
