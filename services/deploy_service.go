package services

import (
	"brand-config-api/config"
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

// DeployService 部署服务
type DeployService struct {
	config *config.Config
}

// NewDeployService 创建部署服务
func NewDeployService() *DeployService {
	return &DeployService{
		config: config.Load(),
	}
}

// ServerInfo 服务器连接信息
type ServerInfo struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password,omitempty"`
	KeyPath  string `json:"keyPath,omitempty"`
}

// NginxDeployConfig nginx部署配置
type NginxDeployConfig struct {
	Domain       string     `json:"domain"`
	Port         int        `json:"port"`
	RootPath     string     `json:"rootPath"`
	LocationPath string     `json:"locationPath"`
	SSLCertPath  string     `json:"sslCertPath,omitempty"`
	SSLKeyPath   string     `json:"sslKeyPath,omitempty"`
	Server       ServerInfo `json:"server"`
}

// DeployResult 部署结果 (已废弃，仅保留用于兼容性)
type DeployResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Output  string `json:"output,omitempty"`
}

// OutputMessage 输出消息
type OutputMessage struct {
	Type    string `json:"type"` // output, error, success, failed
	Message string `json:"message"`
}

// checkAndUploadScript 检查并上传脚本文件
func (s *DeployService) checkAndUploadScript(client *ssh.Client, scriptName string) error {
	log.Printf("🔍 检查服务器上的脚本文件: %s", scriptName)

	// 检查脚本文件是否存在
	checkCmd := fmt.Sprintf("test -f %s/%s && echo 'exists' || echo 'not_exists'",
		"/opt/scripts", scriptName)

	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("创建SSH会话失败: %v", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput(checkCmd)
	if err != nil {
		return fmt.Errorf("检查脚本文件失败: %v", err)
	}

	if string(output) == "exists\n" {
		log.Printf("✅ 脚本文件已存在: %s", scriptName)
		return nil
	}

	log.Printf("📤 脚本文件不存在，开始上传: %s", scriptName)

	// 创建脚本目录
	mkdirCmd := fmt.Sprintf("mkdir -p %s", "/opt/scripts")
	session, err = client.NewSession()
	if err != nil {
		return fmt.Errorf("创建SSH会话失败: %v", err)
	}
	defer session.Close()

	if err := session.Run(mkdirCmd); err != nil {
		return fmt.Errorf("创建脚本目录失败: %v", err)
	}

	// 上传脚本文件
	if err := s.uploadScriptFile(client, scriptName); err != nil {
		return fmt.Errorf("上传脚本文件失败: %v", err)
	}

	// 设置脚本文件权限
	chmodCmd := fmt.Sprintf("chmod +x %s/%s", "/opt/scripts", scriptName)
	session, err = client.NewSession()
	if err != nil {
		return fmt.Errorf("创建SSH会话失败: %v", err)
	}
	defer session.Close()

	if err := session.Run(chmodCmd); err != nil {
		return fmt.Errorf("设置脚本权限失败: %v", err)
	}

	log.Printf("✅ 脚本文件上传并设置权限成功: %s", scriptName)
	return nil
}

// uploadScriptFile 上传脚本文件到服务器
func (s *DeployService) uploadScriptFile(client *ssh.Client, scriptName string) error {
	// 本地脚本文件路径 - 使用GetLocalScriptPath方法
	localScriptPath := s.config.GetLocalScriptPath(scriptName)

	// 检查本地脚本文件是否存在
	if _, err := os.Stat(localScriptPath); os.IsNotExist(err) {
		return fmt.Errorf("本地脚本文件不存在: %s", localScriptPath)
	}

	// 使用scp上传文件
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("创建SSH会话失败: %v", err)
	}
	defer session.Close()

	// 创建远程文件
	remoteFile, err := session.StdinPipe()
	if err != nil {
		return fmt.Errorf("创建远程文件失败: %v", err)
	}

	// 执行scp命令
	go func() {
		defer remoteFile.Close()

		// 读取本地文件内容
		localFile, err := os.Open(localScriptPath)
		if err != nil {
			log.Printf("❌ 打开本地文件失败: %v", err)
			return
		}
		defer localFile.Close()

		// 写入文件内容
		if _, err := io.Copy(remoteFile, localFile); err != nil {
			log.Printf("❌ 写入远程文件失败: %v", err)
		}
	}()

	// 执行scp命令
	scpCmd := fmt.Sprintf("scp -t %s/%s", "/opt/scripts", scriptName)
	if err := session.Run(scpCmd); err != nil {
		return fmt.Errorf("SCP上传失败: %v", err)
	}

	return nil
}

// TestServerConnection 测试服务器连接
func (s *DeployService) TestServerConnection(server ServerInfo) error {
	log.Printf("🔍 测试服务器连接: %s@%s:%d", server.Username, server.Host, server.Port)

	// 使用抽取的公共函数创建SSH连接
	client, err := s.createSSHClient(server, time.Duration(s.config.Deploy.SSHTimeout)*time.Second)
	if err != nil {
		return err
	}
	defer client.Close()

	log.Printf("✅ 服务器连接成功")
	return nil
}

// createSSHClient 创建SSH客户端连接（抽取公共代码）
func (s *DeployService) createSSHClient(server ServerInfo, timeout time.Duration) (*ssh.Client, error) {
	config := &ssh.ClientConfig{
		User:            server.Username,
		Auth:            []ssh.AuthMethod{},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         timeout,
	}

	// 设置认证方式
	if server.Password != "" {
		config.Auth = append(config.Auth, ssh.Password(server.Password))
	} else if server.KeyPath != "" {
		return nil, fmt.Errorf("SSH密钥认证暂未实现")
	} else {
		return nil, fmt.Errorf("必须提供密码或SSH密钥")
	}

	// 建立连接
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", server.Host, server.Port), config)
	if err != nil {
		return nil, fmt.Errorf("SSH连接失败: %v", err)
	}

	return client, nil
}

// buildScriptCommand 构建脚本命令（抽取公共代码）
func (s *DeployService) buildScriptCommand(config NginxDeployConfig, scriptDir string) string {
	scriptCmd := fmt.Sprintf("bash %s/%s %s %d %s %s",
		scriptDir,
		"deploy_nginx_config_linux_server.sh",
		config.Domain,
		config.Port,
		config.RootPath,
		config.LocationPath,
	)

	// 添加SSL参数
	if config.SSLCertPath != "" && config.SSLKeyPath != "" {
		scriptCmd += fmt.Sprintf(" %s %s", config.SSLCertPath, config.SSLKeyPath)
	} else {
		scriptCmd += " \"\" \"\""
	}

	return scriptCmd
}

// ExecuteDeployScriptWithStream 远程执行nginx部署脚本（带流式输出）
func (s *DeployService) ExecuteDeployScriptWithStream(config NginxDeployConfig, outputChan chan<- OutputMessage) error {
	log.Printf("🚀 开始远程执行nginx部署脚本: %s -> %s (端口: %d)", config.Domain, config.LocationPath, config.Port)

	// 发送开始消息
	outputChan <- OutputMessage{
		Type:    "output",
		Message: fmt.Sprintf("🚀 开始远程部署: %s:%d%s", config.Domain, config.Port, config.LocationPath),
	}

	// 首先测试连接
	outputChan <- OutputMessage{
		Type:    "output",
		Message: fmt.Sprintf("🔍 测试服务器连接: %s@%s:%d", config.Server.Username, config.Server.Host, config.Server.Port),
	}

	if err := s.TestServerConnection(config.Server); err != nil {
		errMsg := fmt.Sprintf("服务器连接失败: %v", err)
		outputChan <- OutputMessage{Type: "error", Message: errMsg}
		outputChan <- OutputMessage{Type: "failed", Message: errMsg}
		return err
	}

	outputChan <- OutputMessage{Type: "output", Message: "✅ 服务器连接成功"}

	// 建立SSH连接
	outputChan <- OutputMessage{Type: "output", Message: "🔗 建立SSH连接..."}

	client, err := s.createSSHClient(config.Server, time.Duration(s.config.Deploy.DeployTimeout)*time.Second)
	if err != nil {
		errMsg := fmt.Sprintf("SSH连接失败: %v", err)
		outputChan <- OutputMessage{Type: "error", Message: errMsg}
		outputChan <- OutputMessage{Type: "failed", Message: errMsg}
		return err
	}
	defer client.Close()

	outputChan <- OutputMessage{Type: "output", Message: "✅ SSH连接建立成功"}

	// 检查并上传脚本文件
	outputChan <- OutputMessage{Type: "output", Message: "📁 检查部署脚本文件..."}

	// 需要上传的Linux脚本文件列表 - 更新到deploy目录
	scriptsToUpload := []string{
		"deploy_nginx_config_linux_server.sh",
		"configure_dns_linux_server.sh",
	}

	// 注释：简化版本不再需要保存密码

	// 确定脚本目录
	scriptDir, err := s.ensureScriptDirectory(client, outputChan)
	if err != nil {
		errMsg := fmt.Sprintf("确定脚本目录失败: %v", err)
		outputChan <- OutputMessage{Type: "error", Message: errMsg}
		outputChan <- OutputMessage{Type: "failed", Message: errMsg}
		return err
	}

	// 逐个检查并上传脚本文件
	for _, scriptName := range scriptsToUpload {
		if err := s.checkAndUploadScriptWithStream(client, scriptName, scriptDir, outputChan); err != nil {
			errMsg := fmt.Sprintf("上传脚本文件 %s 失败: %v", scriptName, err)
			outputChan <- OutputMessage{Type: "error", Message: errMsg}
			outputChan <- OutputMessage{Type: "failed", Message: errMsg}
			return err
		}
	}

	// 构建脚本命令 - 使用抽取的公共函数
	scriptCmd := s.buildScriptCommand(config, scriptDir)

	// 添加SSL日志信息
	if config.SSLCertPath != "" && config.SSLKeyPath != "" {
		outputChan <- OutputMessage{Type: "output", Message: "🔒 检测到SSL证书配置，将启用HTTPS"}
	} else {
		outputChan <- OutputMessage{Type: "output", Message: "🔓 未配置SSL证书，将使用HTTP"}
	}

	outputChan <- OutputMessage{Type: "output", Message: fmt.Sprintf("📝 执行命令: %s", scriptCmd)}

	// 创建SSH会话
	session, err := client.NewSession()
	if err != nil {
		errMsg := fmt.Sprintf("创建SSH会话失败: %v", err)
		outputChan <- OutputMessage{Type: "error", Message: errMsg}
		outputChan <- OutputMessage{Type: "failed", Message: errMsg}
		return err
	}
	defer session.Close()

	// 请求伪终端以支持sudo密码输入
	if err := session.RequestPty("xterm", 80, 24, ssh.TerminalModes{
		ssh.ECHO:          0,     // 不回显密码
		ssh.TTY_OP_ISPEED: 14400, // 输入速度
		ssh.TTY_OP_OSPEED: 14400, // 输出速度
	}); err != nil {
		errMsg := fmt.Sprintf("请求伪终端失败: %v", err)
		outputChan <- OutputMessage{Type: "error", Message: errMsg}
		outputChan <- OutputMessage{Type: "failed", Message: errMsg}
		return err
	}

	// 创建stdin管道用于发送密码
	stdin, err := session.StdinPipe()
	if err != nil {
		errMsg := fmt.Sprintf("创建stdin管道失败: %v", err)
		outputChan <- OutputMessage{Type: "error", Message: errMsg}
		outputChan <- OutputMessage{Type: "failed", Message: errMsg}
		return err
	}

	// 创建stdout和stderr管道
	stdout, err := session.StdoutPipe()
	if err != nil {
		errMsg := fmt.Sprintf("创建stdout管道失败: %v", err)
		outputChan <- OutputMessage{Type: "error", Message: errMsg}
		outputChan <- OutputMessage{Type: "failed", Message: errMsg}
		return err
	}

	stderr, err := session.StderrPipe()
	if err != nil {
		errMsg := fmt.Sprintf("创建stderr管道失败: %v", err)
		outputChan <- OutputMessage{Type: "error", Message: errMsg}
		outputChan <- OutputMessage{Type: "failed", Message: errMsg}
		return err
	}

	// 启动命令
	if err := session.Start(scriptCmd); err != nil {
		errMsg := fmt.Sprintf("启动远程脚本失败: %v", err)
		outputChan <- OutputMessage{Type: "error", Message: errMsg}
		outputChan <- OutputMessage{Type: "failed", Message: errMsg}
		return err
	}

	// 发送密码到stdin以供sudo使用
	go func() {
		defer stdin.Close()
		// 先发送一次密码给脚本开始时的sudo调用
		stdin.Write([]byte(config.Server.Password + "\n"))
		time.Sleep(500 * time.Millisecond)

		// 然后持续发送密码以应对可能的其他sudo提示
		for i := 0; i < 10; i++ {
			stdin.Write([]byte(config.Server.Password + "\n"))
			time.Sleep(200 * time.Millisecond)
		}
	}()

	outputChan <- OutputMessage{Type: "output", Message: "🚀 远程脚本开始执行..."}
	outputChan <- OutputMessage{Type: "output", Message: strings.Repeat("=", 60)}

	// 读取stdout
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			outputChan <- OutputMessage{Type: "output", Message: line}
		}
	}()

	// 读取stderr
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			outputChan <- OutputMessage{Type: "error", Message: line}
		}
	}()

	// 等待命令完成
	if err := session.Wait(); err != nil {
		errMsg := fmt.Sprintf("远程脚本执行失败: %v", err)
		outputChan <- OutputMessage{Type: "error", Message: errMsg}
		outputChan <- OutputMessage{Type: "failed", Message: "远程部署失败"}
		return err
	}

	outputChan <- OutputMessage{Type: "output", Message: strings.Repeat("=", 60)}
	outputChan <- OutputMessage{Type: "success", Message: "远程部署成功完成"}
	log.Printf("✅ 远程nginx部署脚本执行成功")

	return nil
}

// ensureScriptDirectory 确保脚本目录存在并返回可用的目录路径（简化版本）
func (s *DeployService) ensureScriptDirectory(client *ssh.Client, outputChan chan<- OutputMessage) (string, error) {
	outputChan <- OutputMessage{Type: "output", Message: "📁 检查脚本目录..."}

	// 简单策略：直接使用用户主目录，避免权限问题
	scriptDir := "~/scripts"

	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("创建SSH会话失败: %v", err)
	}
	defer session.Close()

	// 创建用户主目录下的scripts文件夹
	mkdirCmd := fmt.Sprintf("mkdir -p %s && chmod 755 %s", scriptDir, scriptDir)
	output, err := session.CombinedOutput(mkdirCmd)

	if err != nil {
		return "", fmt.Errorf("创建脚本目录失败: %v, 输出: %s", err, string(output))
	}

	outputChan <- OutputMessage{Type: "output", Message: "✅ 脚本目录创建成功: ~/scripts"}
	return scriptDir, nil
}

// checkAndUploadScriptWithStream 检查并上传脚本文件（带流式输出）
func (s *DeployService) checkAndUploadScriptWithStream(client *ssh.Client, scriptName string, scriptDir string, outputChan chan<- OutputMessage) error {
	outputChan <- OutputMessage{Type: "output", Message: fmt.Sprintf("🔍 检查服务器上的脚本文件: %s", scriptName)}

	// 检查脚本文件是否存在
	checkCmd := fmt.Sprintf("test -f %s/%s && echo 'exists' || echo 'not_exists'",
		scriptDir, scriptName)

	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("创建SSH会话失败: %v", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput(checkCmd)
	if err != nil {
		return fmt.Errorf("检查脚本文件失败: %v", err)
	}

	if string(output) == "exists\n" {
		outputChan <- OutputMessage{Type: "output", Message: "✅ 脚本文件已存在"}
		return nil
	}

	outputChan <- OutputMessage{Type: "output", Message: "📤 脚本文件不存在，开始上传..."}

	// 上传脚本文件
	if err := s.uploadScriptFileWithStream(client, scriptName, scriptDir, outputChan); err != nil {
		return fmt.Errorf("上传脚本文件失败: %v", err)
	}

	// 设置脚本文件权限
	chmodCmd := fmt.Sprintf("chmod +x %s/%s", scriptDir, scriptName)
	session, err = client.NewSession()
	if err != nil {
		return fmt.Errorf("创建SSH会话失败: %v", err)
	}
	defer session.Close()

	if err := session.Run(chmodCmd); err != nil {
		return fmt.Errorf("设置脚本权限失败: %v", err)
	}

	outputChan <- OutputMessage{Type: "output", Message: "✅ 脚本文件上传并设置权限成功"}
	return nil
}

// uploadScriptFileWithStream 上传脚本文件到服务器（带流式输出）
func (s *DeployService) uploadScriptFileWithStream(client *ssh.Client, scriptName string, scriptDir string, outputChan chan<- OutputMessage) error {
	// 本地脚本文件路径 - 使用GetLocalScriptPath方法
	localScriptPath := s.config.GetLocalScriptPath(scriptName)

	// 检查本地脚本文件是否存在
	if _, err := os.Stat(localScriptPath); os.IsNotExist(err) {
		return fmt.Errorf("本地脚本文件不存在: %s", localScriptPath)
	}

	outputChan <- OutputMessage{Type: "output", Message: fmt.Sprintf("📄 读取本地脚本文件: %s", localScriptPath)}

	// 读取本地文件内容
	localFile, err := os.Open(localScriptPath)
	if err != nil {
		return fmt.Errorf("打开本地文件失败: %v", err)
	}
	defer localFile.Close()

	// 获取文件信息
	fileInfo, err := localFile.Stat()
	if err != nil {
		return fmt.Errorf("获取文件信息失败: %v", err)
	}

	outputChan <- OutputMessage{Type: "output", Message: fmt.Sprintf("📊 文件大小: %d 字节", fileInfo.Size())}

	// 使用简单的方式上传文件内容
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("创建SSH会话失败: %v", err)
	}
	defer session.Close()

	// 读取文件内容
	content, err := io.ReadAll(localFile)
	if err != nil {
		return fmt.Errorf("读取文件内容失败: %v", err)
	}

	// 创建远程文件
	createCmd := fmt.Sprintf("cat > %s/%s", scriptDir, scriptName)
	session.Stdin = bytes.NewReader(content)

	if err := session.Run(createCmd); err != nil {
		return fmt.Errorf("创建远程文件失败: %v", err)
	}

	outputChan <- OutputMessage{Type: "output", Message: "✅ 文件上传成功"}
	return nil
}
