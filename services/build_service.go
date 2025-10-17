package services

import (
	"brand-config-api/config"
	"brand-config-api/utils"
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

// BuildService 构建服务
type BuildService struct {
	config *config.Config
}

// NewBuildService 创建构建服务
func NewBuildService() *BuildService {
	return &BuildService{
		config: config.Load(),
	}
}

// ValidateSSHConnection 验证SSH连接
func (s *BuildService) ValidateSSHConnection(host, user, password string) error {
	log.Printf("🔐 开始验证SSH连接: %s@%s", user, host)

	// 配置SSH客户端
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 生产环境需要更安全的验证
		Timeout:         time.Duration(s.config.Deploy.SSHTimeout) * time.Second,
	}

	// 构建连接地址
	address := net.JoinHostPort(host, fmt.Sprintf("%d", s.config.Deploy.DefaultSSHPort))

	// 尝试建立SSH连接
	client, err := ssh.Dial("tcp", address, config)
	if err != nil {
		log.Printf("❌ SSH连接失败: %v", err)
		return fmt.Errorf("SSH连接失败: %v", err)
	}
	defer client.Close()

	// 创建一个简单的会话来验证连接
	session, err := client.NewSession()
	if err != nil {
		log.Printf("❌ 创建SSH会话失败: %v", err)
		return fmt.Errorf("创建SSH会话失败: %v", err)
	}
	defer session.Close()

	// 执行一个简单的命令来验证连接
	output, err := session.CombinedOutput("echo 'SSH连接验证成功'")
	if err != nil {
		log.Printf("❌ SSH命令执行失败: %v", err)
		return fmt.Errorf("SSH命令执行失败: %v", err)
	}

	log.Printf("✅ SSH连接验证成功: %s", strings.TrimSpace(string(output)))
	return nil
}

// BatchBuildRequest 批量构建请求
type BatchBuildRequest struct {
	Projects     []string `json:"projects"`      // 项目列表，如 ["tth5-xingchen", "ksh5-xingchen"]
	Branch       string   `json:"branch"`        // Git分支
	Version      string   `json:"version"`       // 版本号
	Environment  string   `json:"environment"`   // 环境：master,test,local
	ForceForeign bool     `json:"force_foreign"` // 强制外网套餐
	SSHHost      string   `json:"ssh_host"`      // SSH主机
	SSHUser      string   `json:"ssh_user"`      // SSH用户名
	SSHPassword  string   `json:"ssh_password"`  // SSH密码
}

// BuildProgress 构建进度
type BuildProgress struct {
	Percentage int    `json:"percentage"`
	Status     string `json:"status"`  // running, success, failed
	Text       string `json:"text"`    // 当前步骤描述
	Detail     string `json:"detail"`  // 详细信息
	Project    string `json:"project"` // 当前处理的项目
	Output     string `json:"output"`  // 实时输出
}

// BuildResult 构建结果
type BuildResult struct {
	Success    bool                     `json:"success"`
	TotalTime  string                   `json:"total_time"`
	Projects   []string                 `json:"projects"`
	Results    map[string]ProjectResult `json:"results"`     // 每个项目的结果
	OutputPath string                   `json:"output_path"` // 构建产物路径
	LogPath    string                   `json:"log_path"`    // 日志路径
}

// ProjectResult 单个项目构建结果
type ProjectResult struct {
	Success   bool   `json:"success"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	Duration  string `json:"duration"`
	Error     string `json:"error,omitempty"`
}

// ProgressCallback 进度回调函数类型
type ProgressCallback func(progress BuildProgress)

// ExecuteBatchBuild 执行批量构建
func (s *BuildService) ExecuteBatchBuild(req *BatchBuildRequest, progressCallback ProgressCallback) (*BuildResult, error) {
	startTime := time.Now()

	log.Printf("🚀 开始批量构建: 项目=%v, 分支=%s, 环境=%s", req.Projects, req.Branch, req.Environment)

	// 初始化结果
	result := &BuildResult{
		Success:  true,
		Projects: req.Projects,
		Results:  make(map[string]ProjectResult),
	}

	// 发送开始进度
	if progressCallback != nil {
		progressCallback(BuildProgress{
			Percentage: 0,
			Status:     "running",
			Text:       "开始批量构建...",
			Detail:     fmt.Sprintf("准备构建 %d 个项目", len(req.Projects)),
		})
	}

	// 第一步：验证SSH连接
	if progressCallback != nil {
		progressCallback(BuildProgress{
			Percentage: 1,
			Status:     "running",
			Text:       "验证SSH连接...",
			Detail:     fmt.Sprintf("正在连接 %s@%s", req.SSHUser, req.SSHHost),
		})
	}

	if err := s.ValidateSSHConnection(req.SSHHost, req.SSHUser, req.SSHPassword); err != nil {
		result.Success = false
		if progressCallback != nil {
			progressCallback(BuildProgress{
				Percentage: 0,
				Status:     "failed",
				Text:       "SSH连接验证失败",
				Detail:     err.Error(),
			})
		}
		return result, fmt.Errorf("SSH connection validation failed: %v", err)
	}

	// SSH验证成功
	if progressCallback != nil {
		progressCallback(BuildProgress{
			Percentage: 5,
			Status:     "running",
			Text:       "SSH连接验证成功",
			Detail:     "开始验证构建环境...",
		})
	}

	// 第二步：验证构建环境
	if err := s.validateBuildEnvironment(); err != nil {
		result.Success = false
		if progressCallback != nil {
			progressCallback(BuildProgress{
				Percentage: 5,
				Status:     "failed",
				Text:       "环境验证失败",
				Detail:     err.Error(),
			})
		}
		return result, fmt.Errorf("build environment validation failed: %v", err)
	}

	// 设置默认值
	if req.Branch == "" {
		req.Branch = "uni/funNovel/devNew"
	}
	if req.Version == "" {
		req.Version = "1.0.0"
	}
	if req.Environment == "" {
		req.Environment = "master"
	}

	// 构建项目列表字符串
	projectsStr := strings.Join(req.Projects, ",")

	// 调用构建脚本
	buildResult, err := s.executeBuildScript(req, projectsStr, progressCallback)
	if err != nil {
		result.Success = false
		return result, err
	}

	// 解析构建结果
	s.parseBuildResults(buildResult, result)

	// 计算总耗时
	result.TotalTime = time.Since(startTime).String()

	// 发送完成进度
	if progressCallback != nil {
		status := "success"
		text := "批量构建完成"
		if !result.Success {
			status = "failed"
			text = "批量构建失败"
		}

		progressCallback(BuildProgress{
			Percentage: 100,
			Status:     status,
			Text:       text,
			Detail:     fmt.Sprintf("总耗时: %s", result.TotalTime),
		})
	}

	log.Printf("✅ 批量构建完成: 成功=%v, 耗时=%s", result.Success, result.TotalTime)
	return result, nil
}

// validateBuildEnvironment 验证构建环境
func (s *BuildService) validateBuildEnvironment() error {
	// 检查构建脚本是否存在 - 使用GetLocalScriptPath方法
	scriptPath := s.config.GetLocalScriptPath("h5_novel_build_linux.sh")
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return fmt.Errorf("build script not found: %s", scriptPath)
	}

	// 检查脚本是否可执行
	if err := s.ensureScriptExecutable(scriptPath); err != nil {
		return fmt.Errorf("failed to make script executable: %v", err)
	}

	// 检查工作目录 - 使用File.BasePath作为项目根目录
	workspaceDir := filepath.Join(s.config.File.BasePath, "workspace")
	if err := os.MkdirAll(workspaceDir, 0755); err != nil {
		return fmt.Errorf("failed to create workspace directory: %v", err)
	}

	return nil
}

// ensureScriptExecutable 确保脚本可执行
func (s *BuildService) ensureScriptExecutable(scriptPath string) error {
	return os.Chmod(scriptPath, 0755)
}

// executeBuildScript 执行构建脚本
func (s *BuildService) executeBuildScript(req *BatchBuildRequest, projectsStr string, progressCallback ProgressCallback) (string, error) {
	// 使用GetLocalScriptPath方法获取构建脚本路径
	scriptPath := s.config.GetLocalScriptPath("h5_novel_build_linux.sh")

	// 构建脚本参数
	scriptArgs := []string{
		"-b", req.Branch,
		"-v", req.Version,
		"-e", req.Environment,
		"-p", projectsStr,
	}

	// 添加可选参数
	if req.ForceForeign {
		scriptArgs = append(scriptArgs, "-f")
	}
	// 默认自动部署，总是添加 -d 参数
	scriptArgs = append(scriptArgs, "-d")

	// 检测操作系统并构建适当的命令
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		// Windows系统使用bash执行shell脚本
		args := append([]string{scriptPath}, scriptArgs...)
		cmd = exec.Command("bash", args...)

		// 临时注释掉Windows特定字段，避免跨平台编译问题
		// 在Windows环境下运行时，这些字段的缺失不会影响基本功能
		// cmd.SysProcAttr = &syscall.SysProcAttr{
		//     HideWindow:    true,
		//     CreationFlags: 0x08000000, // CREATE_NO_WINDOW
		// }

		log.Printf("📋 执行构建命令 (Windows): bash %s %v", scriptPath, scriptArgs)
	} else {
		// Unix系统直接执行脚本
		cmd = exec.Command(scriptPath, scriptArgs...)
		log.Printf("📋 执行构建命令 (Unix): %s %v", scriptPath, scriptArgs)
	}

	cmd.Dir = s.config.File.BasePath

	// 设置环境变量 - 重点是禁用缓冲和SSH配置
	env := append(os.Environ(),
		"PYTHONUNBUFFERED=1",             // Python无缓冲输出
		"NODE_NO_WARNINGS=1",             // 减少Node.js警告
		"FORCE_COLOR=0",                  // 禁用彩色输出，避免ANSI转义序列干扰
		"CI=true",                        // 设置CI环境，通常会减少缓冲
		"TERM=dumb",                      // 设置终端类型，避免交互式提示
		"DEBIAN_FRONTEND=noninteractive", // 非交互式模式
		"STDBUF=--output=L",              // 强制行缓冲输出
		"UNBUFFER=1",                     // 禁用缓冲
		// SSH配置环境变量 - 从请求中获取
		"SSH_HOST="+req.SSHHost,
		"SSH_USER="+req.SSHUser,
		"SSH_PASSWORD="+req.SSHPassword,
	)

	// 根据操作系统设置不同的环境变量
	if runtime.GOOS == "windows" {
		// Windows下确保bash和相关工具可用
		env = append(env, "MSYSTEM=MINGW64")
		env = append(env, "CHERE_INVOKING=1")
	} else {
		// Unix系统的Node.js路径
		env = append(env,
			"NODE_HOME=/home/fun/.nvm/versions/node/v20.18.1/bin",
			"PATH=/home/fun/.nvm/versions/node/v20.18.1/bin:"+os.Getenv("PATH"),
		)
	}

	cmd.Env = env

	// 创建管道捕获输出
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("failed to create stdout pipe: %v", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", fmt.Errorf("failed to create stderr pipe: %v", err)
	}

	// 启动命令
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start build script: %v", err)
	}

	// 发送开始消息
	if progressCallback != nil {
		progressCallback(BuildProgress{
			Percentage: 1,
			Status:     "running",
			Text:       "构建脚本已启动",
			Detail:     fmt.Sprintf("执行命令: %s %v", scriptPath, scriptArgs),
			Output:     "🚀 构建脚本开始执行...",
		})
		log.Printf("✅ 构建进度回调已设置，开始捕获输出")
	} else {
		log.Printf("⚠️ 构建进度回调为空，无法发送实时输出")
	}

	// 使用 WaitGroup 等待所有输出处理完成
	var wg sync.WaitGroup
	var outputBuilder strings.Builder
	var mu sync.Mutex

	// 处理标准输出
	wg.Add(1)
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stdout)
		// 设置缓冲区，但不要太大以确保实时性
		buf := make([]byte, 0, 4*1024)
		scanner.Buffer(buf, 64*1024)

		for scanner.Scan() {
			line := scanner.Text()

			mu.Lock()
			outputBuilder.WriteString(line + "\n")
			mu.Unlock()

			// 发送有意义的输出到前端，但不包含进度百分比
			if progressCallback != nil {
				progressCallback(BuildProgress{
					Status:     "running",
					Text:       "",
					Detail:     "",
					Output:     line, // 发送原始输出用于前端日志显示
					Percentage: -999, // 使用特殊值表示不显示进度
				})
			}

			log.Printf("[BUILD] %s", line)
		}

		if err := scanner.Err(); err != nil {
			log.Printf("构建输出扫描错误: %v", err)
		}
	}()

	// 处理标准错误输出
	wg.Add(1)
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stderr)
		// 设置缓冲区，但不要太大以确保实时性
		buf := make([]byte, 0, 4*1024)
		scanner.Buffer(buf, 64*1024)

		for scanner.Scan() {
			line := scanner.Text()

			// 智能判断是否为真正的错误
			isError := utils.IsRealError(line)

			mu.Lock()
			if isError {
				outputBuilder.WriteString("ERROR: " + line + "\n")
			} else {
				outputBuilder.WriteString(line + "\n")
			}
			mu.Unlock()

			// 发送有意义的错误输出到前端，但不包含进度百分比
			if progressCallback != nil {
				var outputLine string
				if isError {
					outputLine = "ERROR: " + line
				} else {
					outputLine = line
				}

				progressCallback(BuildProgress{
					Status:     "running",
					Text:       "",
					Detail:     "",
					Output:     outputLine, // 发送处理后的输出用于前端日志显示
					Percentage: -999,       // 使用特殊值表示不显示进度
				})
			}

			log.Printf("[BUILD-STDERR] %s", line)
		}

		if err := scanner.Err(); err != nil {
			log.Printf("构建错误输出扫描错误: %v", err)
		}
	}()

	// 等待命令完成
	err = cmd.Wait()

	// 等待所有输出处理完成
	wg.Wait()

	output := outputBuilder.String()

	if err != nil {
		if progressCallback != nil {
			progressCallback(BuildProgress{
				Percentage: 0,
				Status:     "failed",
				Text:       "构建脚本执行失败",
				Detail:     err.Error(),
				Output:     fmt.Sprintf("构建失败: %v", err),
			})
		}
		return output, fmt.Errorf("build script execution failed: %v", err)
	}

	return output, nil
}

// parseBuildResults 解析构建结果
func (s *BuildService) parseBuildResults(output string, result *BuildResult) {
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		// 解析项目构建结果
		if strings.Contains(line, "build:") && strings.Contains(line, "success") {
			// 示例: "build: tth5-xingchen, master, success"
			parts := strings.Split(line, ",")
			if len(parts) >= 3 {
				// 安全地解析项目名称
				projectPart := strings.TrimSpace(parts[1])
				projectColonParts := strings.Split(projectPart, ":")
				if len(projectColonParts) >= 2 {
					project := strings.TrimSpace(projectColonParts[1])
					result.Results[project] = ProjectResult{
						Success: true,
					}
				}
			}
		} else if strings.Contains(line, "build:") && strings.Contains(line, "fail") {
			// 示例: "build: tth5-xingchen, master, fail"
			parts := strings.Split(line, ",")
			if len(parts) >= 3 {
				// 安全地解析项目名称
				projectPart := strings.TrimSpace(parts[1])
				projectColonParts := strings.Split(projectPart, ":")
				if len(projectColonParts) >= 2 {
					project := strings.TrimSpace(projectColonParts[1])
					result.Results[project] = ProjectResult{
						Success: false,
						Error:   "Build failed",
					}
					result.Success = false
				}
			}
		}
	}

	// 设置输出路径
	workspaceDir := filepath.Join(s.config.File.BasePath, "workspace")
	result.OutputPath = filepath.Join(workspaceDir, "dist_backup")
	result.LogPath = filepath.Join(workspaceDir, "realtime.log")
}

// ExecuteH5Build 执行H5项目构建（保持向后兼容）
func (s *BuildService) ExecuteH5Build(branch, version string, environments, projects []string,
	forceForeignNet, deployAfterBuild bool, outputChan chan<- OutputMessage) error {

	log.Printf("🏗️ 开始H5项目构建（兼容模式）")

	// 转换为新的批量构建请求
	batchReq := &BatchBuildRequest{
		Projects:     projects,
		Branch:       branch,
		Version:      version,
		Environment:  strings.Join(environments, ","),
		ForceForeign: forceForeignNet,
		// 不再使用deployAfterBuild参数，默认自动部署
	}

	// 执行批量构建
	_, err := s.ExecuteBatchBuild(batchReq, func(progress BuildProgress) {
		// 转换为旧的输出消息格式
		outputType := "output"
		if progress.Status == "failed" {
			outputType = "failed"
		} else if progress.Status == "success" {
			outputType = "success"
		}

		outputChan <- OutputMessage{
			Type:    outputType,
			Message: progress.Output,
		}
	})

	return err
}
