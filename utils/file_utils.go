package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
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

// RemoveScriptsEntry 删除scripts中的指定条目
func (ju *JSONUtils) RemoveScriptsEntry(content, platformKey string) string {
	// 查找scripts块
	scriptsStart := findStringIndex(content, `"scripts": {`)
	if scriptsStart == -1 {
		return content
	}

	scriptsEnd := findMatchingBrace(content, scriptsStart)
	if scriptsEnd == -1 {
		return content
	}

	// 在scripts块中查找要删除的条目
	devKey := fmt.Sprintf(`"dev:%s"`, platformKey)
	buildKey := fmt.Sprintf(`"build:%s"`, platformKey)

	// 找到要删除的条目的开始和结束位置
	devStart := findStringIndexFrom(content, devKey, scriptsStart)
	if devStart == -1 {
		return content
	}

	// 找到dev条目的结束位置（下一个逗号或大括号）
	devEnd := findScriptEntryEnd(content, devStart, scriptsEnd)

	// 找到build条目的结束位置
	buildStart := findStringIndexFrom(content, buildKey, scriptsStart)
	var buildEnd int
	if buildStart != -1 {
		buildEnd = findScriptEntryEnd(content, buildStart, scriptsEnd)
	} else {
		buildEnd = -1
	}

	// 确定删除范围
	deleteStart := devStart
	deleteEnd := buildEnd
	if buildEnd == -1 {
		deleteEnd = devEnd
	}

	// 使用公共方法处理删除逻辑
	return ju.removeEntryWithCommaHandling(content, deleteStart, deleteEnd, scriptsStart, scriptsEnd)
}

// RemoveUniAppScriptsEntry 删除uni-app.scripts中的指定条目
func (ju *JSONUtils) RemoveUniAppScriptsEntry(content, platformKey string) string {
	// 查找uni-app.scripts块
	uniAppStart := findStringIndex(content, `"uni-app": {`)
	if uniAppStart == -1 {
		return content
	}

	scriptsStart := findStringIndexFrom(content, `"scripts": {`, uniAppStart)
	if scriptsStart == -1 {
		return content
	}

	scriptsEnd := findMatchingBrace(content, scriptsStart)
	if scriptsEnd == -1 {
		return content
	}

	// 在uni-app.scripts块中查找要删除的条目
	platformKeyQuoted := fmt.Sprintf(`"%s"`, platformKey)

	// 找到要删除的条目的开始位置
	entryStart := findStringIndexFrom(content, platformKeyQuoted, scriptsStart)
	if entryStart == -1 {
		return content
	}

	// 找到条目的结束位置（整个配置对象）
	entryEnd := findUniAppScriptEntryEnd(content, entryStart, scriptsEnd)

	// 使用公共方法处理删除逻辑
	return ju.removeEntryWithCommaHandling(content, entryStart, entryEnd, scriptsStart, scriptsEnd)
}

// findScriptEntryEnd 找到scripts条目的结束位置
func findScriptEntryEnd(content string, startIndex, scriptsEnd int) int {
	// 从开始位置向后查找，直到找到下一个逗号或大括号
	// 需要跳过JSON字符串内部的逗号
	inString := false
	escapeNext := false

	for i := startIndex; i < scriptsEnd; i++ {
		char := content[i]

		if escapeNext {
			escapeNext = false
			continue
		}

		if char == '\\' {
			escapeNext = true
			continue
		}

		if char == '"' {
			inString = !inString
			continue
		}

		// 只有在不在字符串内部时才处理逗号和大括号
		if !inString {
			if char == ',' {
				return i + 1
			} else if char == '}' {
				return i
			}
		}
	}
	return scriptsEnd
}

// findUniAppScriptEntryEnd 找到uni-app.scripts条目的结束位置
func findUniAppScriptEntryEnd(content string, startIndex, scriptsEnd int) int {
	// 找到配置对象的开始大括号
	braceStart := findStringIndexFrom(content, "{", startIndex)
	if braceStart == -1 {
		return startIndex
	}

	// 找到配置对象的结束大括号
	braceEnd := findMatchingBrace(content, braceStart)
	if braceEnd == -1 {
		return startIndex
	}

	// 返回配置对象结束后的位置
	return braceEnd + 1
}

// removeEntryWithCommaHandling 公共方法：处理删除条目时的逗号逻辑
func (ju *JSONUtils) removeEntryWithCommaHandling(content string, deleteStart, deleteEnd, scriptsStart, scriptsEnd int) string {
	// 检查删除范围后面是否有逗号，如果有则包含在删除范围内
	if deleteEnd < len(content) && content[deleteEnd] == ',' {
		deleteEnd++
	}

	// 检查删除范围前面是否有逗号，如果有则包含在删除范围内
	// 从删除开始位置向前查找最近的逗号
	if deleteStart > 0 {
		// 向前查找最近的逗号，但要确保不在字符串内部
		inString := false
		escapeNext := false
		for i := deleteStart - 1; i >= scriptsStart; i-- {
			char := content[i]

			if escapeNext {
				escapeNext = false
				continue
			}

			if char == '\\' {
				escapeNext = true
				continue
			}

			if char == '"' {
				inString = !inString
				continue
			}

			// 只有在不在字符串内部时才处理逗号
			if !inString {
				if char == ',' {
					// 找到逗号，检查逗号后面是否只有空白字符
					// 如果逗号后面直接是要删除的内容，则包含这个逗号
					afterComma := strings.TrimSpace(content[i+1 : deleteStart])
					if afterComma == "" {
						// 检查删除范围后面是否还有内容（不是最后一个条目）
						// 从删除结束位置向后查找，看是否还有其他内容
						hasMoreContent := false
						inStringAfter := false
						escapeNextAfter := false

						for j := deleteEnd; j < scriptsEnd; j++ {
							charAfter := content[j]

							if escapeNextAfter {
								escapeNextAfter = false
								continue
							}

							if charAfter == '\\' {
								escapeNextAfter = true
								continue
							}

							if charAfter == '"' {
								inStringAfter = !inStringAfter
								continue
							}

							// 只有在不在字符串内部时才处理
							if !inStringAfter {
								if charAfter == '"' {
									// 找到了下一个内容的开始引号，说明还有更多内容
									hasMoreContent = true
									break
								} else if charAfter == '}' {
									// 遇到了结束大括号，说明没有更多内容了
									break
								}
							}
						}

						// 只有在没有更多内容时才删除前面的逗号
						if !hasMoreContent {
							deleteStart = i
						}
					}
					break
				} else if char == '{' {
					// 遇到大括号，停止查找
					break
				}
			}
		}
	}

	// 删除条目
	newContent := content[:deleteStart] + content[deleteEnd:]

	return newContent
}
