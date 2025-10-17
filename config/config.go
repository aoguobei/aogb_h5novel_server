package config

import (
	"path/filepath"
	"strings"
)

// Config 应用配置结构
type Config struct {
	Database    DatabaseConfig
	Server      ServerConfig
	File        FileConfig
	GitReposDir string       // Git仓库目录
	Deploy      DeployConfig // 部署配置
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port string
	Mode string
}

// FileConfig 文件操作配置
type FileConfig struct {
	BasePath       string // 基础路径，如 C:/F_explorer/h5projects/jianruiH5/novel_h5config
	ProjectRoot    string
	ConfigDir      string
	PrebuildDir    string
	StaticDir      string
	ViteConfigFile string
	PackageFile    string
	// 新增配置目录
	BaseConfigsDir   string
	CommonConfigsDir string
	PayConfigsDir    string
	UIConfigsDir     string
	LocalConfigsDir  string
}

// DeployConfig 部署配置
type DeployConfig struct {
	ScriptsDir     string // 部署脚本目录，如 /opt/scripts
	DefaultSSHPort int    // 默认SSH端口，通常为22
	SSHTimeout     int    // SSH连接超时时间(秒)，建议10-30秒
	DeployTimeout  int    // 部署超时时间(秒)，建议30-120秒
}

// GetLocalScriptPath 获取本地脚本路径
func (c *Config) GetLocalScriptPath(scriptName string) string {
	// 如果是构建脚本，放在build子目录下
	if strings.HasPrefix(scriptName, "h5_novel_build") {
		return filepath.Join(c.Deploy.ScriptsDir, "build", scriptName)
	}

	// 如果是部署脚本，放在deploy子目录下
	if strings.Contains(scriptName, "deploy_") || strings.Contains(scriptName, "configure_") {
		return filepath.Join(c.Deploy.ScriptsDir, "deploy", scriptName)
	}

	// 默认放在scripts根目录下
	return filepath.Join(c.Deploy.ScriptsDir, scriptName)
}

// Load 加载配置
func Load() *Config {
	// 获取项目根目录
	// projectRoot := getEnv("PROJECT_ROOT", "C:/F_explorer/h5projects/fun_h5_webconfig")
	projectRoot := getEnv("PROJECT_ROOT", "/opt/websites/novel_h5_webconfig")

	// 小说funNovel的父目录
	// basePath := getEnv("BASE_PATH", "C:/F_explorer/h5projects/fun_h5_webconfig")
	basePath := getEnv("BASE_PATH", "/opt/websites/novel_h5_webconfig/funNovel_edit")

	return &Config{
		Database: DatabaseConfig{
			Host: getEnv("DB_HOST", "localhost"),
			Port: getEnv("DB_PORT", "3306"),
			User: getEnv("DB_USER", "root"),
			// Password: getEnv("DB_PASSWORD", ""),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", "h5novel_config"),
		},
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
			Mode: getEnv("GIN_MODE", "debug"),
		},
		File: FileConfig{
			BasePath:       basePath,
			ProjectRoot:    filepath.Join(basePath, "funNovel"),
			ConfigDir:      filepath.Join(basePath, "funNovel/src/appConfig"),
			PrebuildDir:    filepath.Join(basePath, "funNovel/prebuild/build"),
			StaticDir:      filepath.Join(basePath, "funNovel/src/static"),
			ViteConfigFile: filepath.Join(basePath, "funNovel/vite.config.js"),
			PackageFile:    filepath.Join(basePath, "funNovel/package.json"),
			// 新增配置目录
			BaseConfigsDir:   filepath.Join(basePath, "funNovel/src/appConfig/baseConfigs"),
			CommonConfigsDir: filepath.Join(basePath, "funNovel/src/appConfig/commonConfigs"),
			PayConfigsDir:    filepath.Join(basePath, "funNovel/src/appConfig/payConfigs"),
			UIConfigsDir:     filepath.Join(basePath, "funNovel/src/appConfig/uiConfigs"),
			LocalConfigsDir:  filepath.Join(basePath, "funNovel/src/appConfig/localConfigs"),
		},
		// 进行分支管理的仓库
		GitReposDir: filepath.Join(projectRoot, "repo-branches"),
		Deploy: DeployConfig{
			ScriptsDir:     filepath.Join(projectRoot, "h5ManagerServer", "scripts"), // 部署脚本目录，绝对路径
			DefaultSSHPort: 22,                                                       // 默认SSH端口
			SSHTimeout:     10,                                                       // SSH连接超时时间(秒)
			DeployTimeout:  30,                                                       // 部署超时时间(秒)
		},
	}
}

// GetConfigPath 获取配置文件路径
func (c *Config) GetConfigPath(configType, brandCode string) string {
	return filepath.Join(c.File.ConfigDir, configType+"Configs", brandCode+".js")
}

// GetPrebuildPath 获取prebuild路径
func (c *Config) GetPrebuildPath(brandCode string) string {
	return filepath.Join(c.File.PrebuildDir, brandCode)
}

// GetStaticPath 获取static路径
func (c *Config) GetStaticPath(brandCode string) string {
	return filepath.Join(c.File.StaticDir, "img-"+brandCode)
}
