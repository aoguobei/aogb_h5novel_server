package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"brand-config-api/config"
	"brand-config-api/utils/rollback"
)

// ConfigFileManager 配置文件写入工具
type ConfigFileManager struct {
	config *config.Config
}

// NewConfigFileManager 创建配置文件写入工具实例
func NewConfigFileManager() *ConfigFileManager {
	return &ConfigFileManager{
		config: config.Load(),
	}
}

// WriteConfigToFile 写入配置文件
func (w *ConfigFileManager) WriteConfigToFile(configType string, formattedConfig map[string]interface{}, brandCode, host string, rollbackManager *rollback.RollbackManager) error {
	// 根据配置类型确定文件路径
	var configDir string
	var fileName string

	switch configType {
	case "base":
		configDir = w.config.File.BaseConfigsDir
		fileName = brandCode + ".js"
	case "common":
		configDir = w.config.File.CommonConfigsDir
		fileName = brandCode + ".js"
	case "pay":
		configDir = w.config.File.PayConfigsDir
		fileName = brandCode + ".js"
	case "ui":
		configDir = w.config.File.UIConfigsDir
		fileName = brandCode + ".js"
	case "novel":
		configDir = w.config.File.LocalConfigsDir
		fileName = "novelConfig.js"
	default:
		return fmt.Errorf("unknown config type: %s", configType)
	}

	// 确保目录存在
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %v", configDir, err)
	}

	// 构建文件路径
	configFile := filepath.Join(configDir, fileName)

	// 在写入前先备份文件（如果存在）
	if _, err := os.Stat(configFile); err == nil {
		// 文件存在，进行备份
		if err := rollbackManager.GetFileManager().Backup(configFile, ""); err != nil {
			return fmt.Errorf("failed to backup file before writing: %v", err)
		}
		fmt.Printf("📋 备份文件: %s\n", configFile)
	}

	// 读取现有配置文件，如果文件不存在则创建新的配置对象
	var configData map[string]interface{}
	configData, err := w.ReadConfigFile(configFile)
	if err != nil {
		// 如果文件不存在，创建新的配置对象
		if os.IsNotExist(err) {
			configData = make(map[string]interface{})
			fmt.Printf("📄 配置文件不存在，创建新的配置: %s\n", configFile)
		} else {
			return fmt.Errorf("failed to read existing config file: %v", err)
		}
	}

	// 使用host作为key，添加配置
	if configType == "novel" {
		// novel配置使用特殊的结构：brandCode -> host -> config
		if configData[brandCode] == nil {
			configData[brandCode] = make(map[string]interface{})
		}
		brandConfig := configData[brandCode].(map[string]interface{})
		brandConfig[host] = formattedConfig
	} else {
		// 其他配置直接使用host作为key
		configData[host] = formattedConfig
	}

	// 写入文件
	if err := w.WriteConfigDataToFile(configData, configFile); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	fmt.Printf("✅ %s config written to: %s with host key: %s\n", configType, configFile, host)
	return nil
}

// ReadConfigFile 读取配置文件
func (w *ConfigFileManager) ReadConfigFile(configFile string) (map[string]interface{}, error) {
	// 检查文件是否存在
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return nil, err
	}

	// 读取文件内容
	content, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	fmt.Printf("📖 读取文件内容，长度: %d\n", len(content))
	fmt.Printf("📖 文件内容预览: %s\n", string(content[:func() int {
		if 100 < len(content) {
			return 100
		} else {
			return len(content)
		}
	}()]))

	// 解析配置 - 采用base_config_service.go中的方法
	contentStr := string(content)

	// 移除export default前缀
	if strings.HasPrefix(contentStr, "export default ") {
		contentStr = strings.TrimPrefix(contentStr, "export default ")
	}

	// 将单引号替换为双引号，使其符合JSON格式
	contentStr = strings.ReplaceAll(contentStr, "'", `"`)

	// 清理多余的逗号，修复JSON格式
	fmt.Printf("🧹 开始清理JSON格式...\n")

	// 使用正则表达式移除对象末尾的多余逗号
	// 匹配模式：, 后面跟着 } 或 ]
	re := regexp.MustCompile(`,(\s*[}\]])`)
	cleanedContent := re.ReplaceAllString(contentStr, "$1")

	// 移除行末的多余逗号
	lines := strings.Split(cleanedContent, "\n")
	var finalLines []string

	for i, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// 如果这行以逗号结尾，检查下一行
		if strings.HasSuffix(trimmedLine, ",") {
			if i+1 < len(lines) {
				nextLine := strings.TrimSpace(lines[i+1])
				// 如果下一行是 } 或 } 开头，移除逗号
				if strings.HasPrefix(nextLine, "}") || strings.HasPrefix(nextLine, "]") {
					cleanedLine := strings.TrimSuffix(line, ",")
					finalLines = append(finalLines, cleanedLine)
					fmt.Printf("🧹 移除行 %d 的多余逗号\n", i+1)
					continue
				}
			}
		}

		finalLines = append(finalLines, line)
	}

	// 重新组合内容
	finalContent := strings.Join(finalLines, "\n")

	// 最后检查：移除对象末尾的逗号
	finalContent = strings.ReplaceAll(finalContent, ",}", "}")
	finalContent = strings.ReplaceAll(finalContent, ",]", "]")

	fmt.Printf("🧹 清理完成，内容长度: %d -> %d\n", len(contentStr), len(finalContent))
	fmt.Printf("🧹 清理后的内容预览: %s\n", finalContent[:func() int {
		if 200 < len(finalContent) {
			return 200
		} else {
			return len(finalContent)
		}
	}()])

	contentStr = finalContent

	var configData map[string]interface{}
	if err := json.Unmarshal([]byte(contentStr), &configData); err != nil {
		fmt.Printf("❌ JSON解析失败: %v\n", err)
		fmt.Printf("❌ 处理后的内容: %s\n", contentStr)
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	fmt.Printf("📖 JSON解析成功，找到host: %v\n", func() []string {
		var keys []string
		for key := range configData {
			keys = append(keys, key)
		}
		return keys
	}())
	return configData, nil
}

// WriteConfigDataToFile 写入配置文件
func (w *ConfigFileManager) WriteConfigDataToFile(configData map[string]interface{}, configFile string) error {
	fmt.Printf("🔄 开始序列化配置数据...\n")

	// 转换为JSON
	configJSON, err := json.MarshalIndent(configData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config to JSON: %v", err)
	}

	fmt.Printf("📄 JSON序列化成功，长度: %d\n", len(configJSON))

	// 添加export default前缀
	content := fmt.Sprintf("export default %s\n", string(configJSON))
	fmt.Printf("📝 生成文件内容，总长度: %d\n", len(content))

	// 写入文件
	fmt.Printf("💾 写入文件: %s\n", configFile)
	if err := os.WriteFile(configFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	fmt.Printf("✅ 文件写入成功\n")
	return nil
}
