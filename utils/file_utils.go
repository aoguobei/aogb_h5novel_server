package utils

import (
	"io"
	"os"
	"path/filepath"
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

// 这些方法已废弃，使用简化的按行删除逻辑
