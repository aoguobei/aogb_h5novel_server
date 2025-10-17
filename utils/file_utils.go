package utils

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"brand-config-api/utils/rollback"
)

// FileUtils 文件操作工具类
type FileUtils struct{}

// NewFileUtils 创建文件工具类实例
func NewFileUtils() *FileUtils {
	return &FileUtils{}
}

// CopyDirectory 递归拷贝目录内容
func (fu *FileUtils) CopyDirectory(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			// 创建子目录
			if err := os.MkdirAll(dstPath, 0755); err != nil {
				return err
			}
			// 递归拷贝子目录
			if err := fu.CopyDirectory(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			// 拷贝文件
			if err := fu.CopyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// CopyFile 拷贝单个文件
func (fu *FileUtils) CopyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// GetUniPlatform 根据host获取对应的uni平台标识
func (fu *FileUtils) GetUniPlatform(host string) string {
	switch host {
	case "tt":
		return "mp-toutiao"
	case "ks":
		return "mp-kuaishou"
	case "wx":
		return "mp-weixin"
	case "bd":
		return "mp-baidu"
	default:
		return "h5"
	}
}

// JSONUtils JSON配置文件解析工具类
type JSONUtils struct{}

// NewJSONUtils 创建JSON工具类实例
func NewJSONUtils() *JSONUtils {
	return &JSONUtils{}
}

// FindScriptsEndIndex 找到scripts块的结束位置
func (ju *JSONUtils) FindScriptsEndIndex(content string) int {
	// 查找 "scripts": { 的开始位置
	scriptsStart := findStringIndex(content, `"scripts": {`)
	if scriptsStart == -1 {
		return -1
	}

	// 从scripts开始位置向后查找对应的结束大括号
	return findMatchingBrace(content, scriptsStart)
}

// FindLastScriptEndIndex 找到scripts块中最后一个脚本的结束位置
func (ju *JSONUtils) FindLastScriptEndIndex(content string, scriptsEndIndex int) int {
	// 从scripts块开始位置向后查找
	scriptsStart := findStringIndex(content, `"scripts": {`)
	if scriptsStart == -1 {
		return -1
	}

	// 在scripts块中查找最后一个脚本的结束位置
	// 从scripts结束位置向前查找最后一个脚本的结束引号
	for i := scriptsEndIndex - 1; i > scriptsStart; i-- {
		// 查找最后一个脚本的结束引号
		if content[i] == '"' {
			// 向前查找这个脚本的开始引号
			for j := i - 1; j > scriptsStart; j-- {
				if content[j] == '"' {
					// 找到了脚本值的开始引号，现在需要找到这个值的结束引号
					// 从开始引号向后查找结束引号
					for k := j + 1; k < scriptsEndIndex; k++ {
						if content[k] == '"' {
							// 找到了值的结束引号，返回这个位置
							return k + 1
						}
					}
				}
			}
		}
	}

	return -1
}

// FindUniAppScriptsEndIndex 找到uni-app.scripts块的结束位置
func (ju *JSONUtils) FindUniAppScriptsEndIndex(content string) int {
	// 查找 "uni-app": { 的开始位置
	uniAppStart := findStringIndex(content, `"uni-app": {`)
	if uniAppStart == -1 {
		return -1
	}

	// 从uni-app开始位置向后查找 "scripts": { 的开始位置
	scriptsStart := findStringIndex(content[uniAppStart:], `"scripts": {`)
	if scriptsStart == -1 {
		return -1
	}

	scriptsStart += uniAppStart

	// 从scripts开始位置向后查找对应的结束大括号
	return findMatchingBrace(content, scriptsStart)
}

// FindLastUniAppScriptEndIndex 找到uni-app.scripts块中最后一个配置的结束位置
func (ju *JSONUtils) FindLastUniAppScriptEndIndex(content string, uniAppScriptsEndIndex int) int {
	// 从uni-app.scripts块开始位置向后查找
	uniAppStart := findStringIndex(content, `"uni-app": {`)
	if uniAppStart == -1 {
		return -1
	}

	// 从uni-app开始位置向后查找 "scripts": { 的开始位置
	scriptsStart := findStringIndex(content[uniAppStart:], `"scripts": {`)
	if scriptsStart == -1 {
		return -1
	}

	scriptsStart += uniAppStart

	// 在uni-app.scripts块中查找最后一个配置的结束位置
	// 从scripts结束位置向前查找最后一个配置的结束大括号
	for i := uniAppScriptsEndIndex - 1; i > scriptsStart; i-- {
		// 查找最后一个配置的结束大括号
		if content[i] == '}' {
			// 向前查找这个配置的开始大括号
			braceCount := 1
			for j := i - 1; j > scriptsStart; j-- {
				if content[j] == '}' {
					braceCount++
				} else if content[j] == '{' {
					braceCount--
					if braceCount == 0 {
						// 找到了配置的开始大括号，返回结束大括号后的位置
						return i + 1
					}
				}
			}
		}
	}

	return -1
}

// findStringIndex 查找字符串在内容中的位置（辅助方法）
func findStringIndex(content, target string) int {
	return findStringIndexFrom(content, target, 0)
}

// findStringIndexFrom 从指定位置开始查找字符串
func findStringIndexFrom(content, target string, startIndex int) int {
	if startIndex >= len(content) {
		return -1
	}

	for i := startIndex; i <= len(content)-len(target); i++ {
		if content[i:i+len(target)] == target {
			return i
		}
	}
	return -1
}

// findMatchingBrace 查找匹配的大括号位置
func findMatchingBrace(content string, startIndex int) int {
	braceCount := 0

	for i := startIndex; i < len(content); i++ {
		if content[i] == '{' {
			braceCount++
		} else if content[i] == '}' {
			braceCount--
			if braceCount == 0 {
				return i
			}
		}
	}

	return -1
}

// RemoveScriptsEntry 删除scripts中的指定条目（已废弃，使用按行删除）
func (ju *JSONUtils) RemoveScriptsEntry(content, platformKey string) string {
	return content
}

// RemoveUniAppScriptsEntry 删除uni-app.scripts中的指定条目（已废弃，使用按行删除）
func (ju *JSONUtils) RemoveUniAppScriptsEntry(content, platformKey string) string {
	return content
}

// ConfigFileUtils 配置文件操作工具
type ConfigFileUtils struct{}

// NewConfigFileUtils 创建配置文件操作工具实例
func NewConfigFileUtils() *ConfigFileUtils {
	return &ConfigFileUtils{}
}

// GenerateConfigFile 通用的配置文件生成方法
func (cfu *ConfigFileUtils) GenerateConfigFile(ctx *rollback.TransactionContext, configFile string, hostConfig map[string]interface{}, host string) error {
	// 检查文件是否存在，如果存在则备份
	// 无论文件是否存在，都调用Backup方法（方法内部会自动判断）
	if err := ctx.Files.Backup(configFile, ""); err != nil {
		return fmt.Errorf("failed to backup file: %v", err)
	}

	// 根据文件状态记录日志
	if _, err := os.Stat(configFile); err == nil {
		log.Printf("📝 备份已存在的配置文件: %s", configFile)
	} else if os.IsNotExist(err) {
		log.Printf("📝 标记新创建的配置文件: %s", configFile)
	} else {
		// 其他错误
		return fmt.Errorf("failed to check file existence: %v", err)
	}

	// 读取现有配置文件或创建新的配置对象
	configfileManager := NewConfigFileManager()
	configData, err := configfileManager.ReadConfigFile(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			// 文件不存在，创建新的配置对象
			configData = make(map[string]interface{})
			log.Printf("📄 配置文件不存在，创建新的配置: %s", configFile)
		} else {
			return fmt.Errorf("failed to read existing config file: %v", err)
		}
	}

	// 更新指定host的配置
	configData[host] = hostConfig

	// 写入文件
	if err := configfileManager.WriteConfigDataToFile(configData, configFile); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	return nil
}

// DeleteConfigFileHost 通用的删除配置文件指定host配置的方法
func (cfu *ConfigFileUtils) DeleteConfigFileHost(ctx *rollback.TransactionContext, configFile string, host string) error {
	// 检查文件是否存在
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return nil
	}

	// 备份文件
	if err := ctx.Files.Backup(configFile, ""); err != nil {
		return fmt.Errorf("failed to backup file: %v", err)
	}

	// 读取现有配置文件
	configfileManager := NewConfigFileManager()
	configData, err := configfileManager.ReadConfigFile(configFile)
	if err != nil {
		return fmt.Errorf("failed to read existing config file: %v", err)
	}

	// 删除指定host的配置
	if _, exists := configData[host]; exists {
		delete(configData, host)
	}

	// 写入更新后的配置文件
	if err := configfileManager.WriteConfigDataToFile(configData, configFile); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	return nil
}

// UpdateConfigFileHost 通用的更新配置文件指定host配置的方法
func (cfu *ConfigFileUtils) UpdateConfigFileHost(ctx *rollback.TransactionContext, configFile string, hostConfig map[string]interface{}, host string) error {
	// 备份文件
	if err := ctx.Files.Backup(configFile, ""); err != nil {
		return fmt.Errorf("failed to backup file: %v", err)
	}

	// 读取现有配置文件
	configfileManager := NewConfigFileManager()
	configData, err := configfileManager.ReadConfigFile(configFile)
	if err != nil {
		return fmt.Errorf("failed to read existing config file: %v", err)
	}

	// 更新指定host的配置
	configData[host] = hostConfig

	// 写入文件
	if err := configfileManager.WriteConfigDataToFile(configData, configFile); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	return nil
}
