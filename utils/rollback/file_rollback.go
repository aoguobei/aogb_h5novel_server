package rollback

import (
	"fmt"
	"log"
	"os"
	"strings"

	"brand-config-api/config"
)

// FileRollbacker 文件回滚接口
type FileRollbacker interface {
	Backup(path, content string) error
	Restore(path string) error
	Rollback() error
	Clear() error
	GetBackupCount() int
	GetCreatedFileCount() int
}

// FileRollback 文件回滚实现
type FileRollback struct {
	config       *config.Config
	backupFiles  map[string]string // 文件路径 -> 原始内容
	createdFiles []string          // 新创建的文件列表
}

// NewFileRollback 创建文件回滚实例
func NewFileRollback(cfg *config.Config) *FileRollback {
	return &FileRollback{
		config:       cfg,
		backupFiles:  make(map[string]string),
		createdFiles: make([]string, 0),
	}
}

// Backup 备份文件内容
func (fr *FileRollback) Backup(path, content string) error {
	log.Printf("🔄 文件回滚器：备份文件 %s", path)

	// 检查路径是否存在
	stat, err := os.Stat(path)
	if err == nil {
		if stat.IsDir() {
			// 是目录，标记为需要删除的目录
			fr.createdFiles = append(fr.createdFiles, path)
			log.Printf("📝 标记为需要删除的目录: %s", path)
		} else {
			// 是文件，读取原始内容作为备份
			originalContent, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("failed to read file for backup: %v", err)
			}
			fr.backupFiles[path] = string(originalContent)
			log.Printf("✅ 文件备份成功: %s (内容长度: %d)", path, len(originalContent))
		}
	} else {
		// 路径不存在，标记为新创建的文件
		fr.createdFiles = append(fr.createdFiles, path)
		log.Printf("📝 标记为新创建文件: %s", path)
	}

	return nil
}

// Restore 恢复文件内容
func (fr *FileRollback) Restore(path string) error {
	// 检查是否有备份
	originalContent, hasBackup := fr.backupFiles[path]
	if hasBackup {
		// 恢复原始内容
		log.Printf("📄 恢复文件: %s (长度: %d)", path, len(originalContent))
		if err := os.WriteFile(path, []byte(originalContent), 0644); err != nil {
			return fmt.Errorf("failed to restore file: %v", err)
		}
		delete(fr.backupFiles, path)
	} else {
		// 检查是否是新创建的文件或目录
		for i, createdPath := range fr.createdFiles {
			if createdPath == path {
				// 检查是文件还是目录
				if stat, err := os.Stat(path); err == nil && stat.IsDir() {
					// 是目录，删除目录及其内容
					if err := os.RemoveAll(path); err != nil && !os.IsNotExist(err) {
						return fmt.Errorf("failed to remove created directory: %v", err)
					}
					log.Printf("✅ 删除新创建目录: %s", path)
				} else {
					// 是文件，删除文件
					if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
						return fmt.Errorf("failed to remove created file: %v", err)
					}
					log.Printf("✅ 删除新创建文件: %s", path)
				}
				// 从列表中移除
				fr.createdFiles = append(fr.createdFiles[:i], fr.createdFiles[i+1:]...)
				break
			}
		}
	}

	return nil
}

// Rollback 执行文件回滚
func (fr *FileRollback) Rollback() error {
	log.Printf("🔄 文件回滚器：开始回滚所有文件操作")
	log.Printf("📊 需要回滚的文件数量: 备份文件=%d, 新创建文件=%d", len(fr.backupFiles), len(fr.createdFiles))

	// 打印备份文件列表
	if len(fr.backupFiles) > 0 {
		log.Printf("📋 备份文件列表:")
		for path := range fr.backupFiles {
			log.Printf("  - %s", path)
		}
	}

	// 打印新创建文件列表
	if len(fr.createdFiles) > 0 {
		log.Printf("📋 新创建文件列表:")
		for _, path := range fr.createdFiles {
			log.Printf("  - %s", path)
		}
	}

	var errors []error
	var successCount int

	// 恢复所有备份的文件
	for path := range fr.backupFiles {
		if err := fr.Restore(path); err != nil {
			errors = append(errors, fmt.Errorf("failed to restore %s: %v", path, err))
			log.Printf("❌ 恢复失败: %s - %v", path, err)
		} else {
			successCount++
		}
	}

	// 删除所有新创建的文件和目录（按深度排序，先删除深层文件）
	sortedPaths := fr.sortPathsByDepth(fr.createdFiles)

	for _, path := range sortedPaths {
		if err := fr.deletePath(path); err != nil {
			errors = append(errors, fmt.Errorf("failed to delete %s: %v", path, err))
			log.Printf("❌ 删除失败: %s - %v", path, err)
		} else {
			successCount++
		}
	}

	// 清空列表
	fr.createdFiles = fr.createdFiles[:0]

	// 结果报告
	if len(errors) > 0 {
		log.Printf("⚠️ 文件回滚完成，但有 %d 个错误", len(errors))
		return fmt.Errorf("file rollback completed with %d errors: %v", len(errors), errors)
	}

	log.Printf("✅ 文件回滚完成，处理 %d 个文件", successCount)
	return nil
}

// deletePath 删除指定路径（文件或目录）
func (fr *FileRollback) deletePath(path string) error {
	// 检查是文件还是目录
	if stat, err := os.Stat(path); err == nil {
		if stat.IsDir() {
			// 是目录，删除目录及其内容
			if err := os.RemoveAll(path); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("failed to remove directory %s: %v", path, err)
			}
			log.Printf("🗑️ 删除目录成功: %s", path)
		} else {
			// 是文件，删除文件
			if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("failed to remove file %s: %v", path, err)
			}
			log.Printf("🗑️ 删除文件成功: %s", path)
		}
	} else if !os.IsNotExist(err) {
		// 其他错误
		return fmt.Errorf("failed to stat %s: %v", path, err)
	} else {
		// 文件不存在，跳过
		log.Printf("⚠️ 文件不存在，跳过删除: %s", path)
	}

	return nil
}

// sortPathsByDepth 按路径深度排序，深度深的排在前面（先删除子文件/目录）
func (fr *FileRollback) sortPathsByDepth(paths []string) []string {
	// 简单的深度排序：按路径分隔符数量排序
	type pathInfo struct {
		path  string
		depth int
	}

	var pathInfos []pathInfo
	for _, path := range paths {
		depth := strings.Count(path, string(os.PathSeparator))
		pathInfos = append(pathInfos, pathInfo{path: path, depth: depth})
	}

	// 按深度降序排序
	for i := 0; i < len(pathInfos)-1; i++ {
		for j := i + 1; j < len(pathInfos); j++ {
			if pathInfos[i].depth < pathInfos[j].depth {
				pathInfos[i], pathInfos[j] = pathInfos[j], pathInfos[i]
			}
		}
	}

	// 提取排序后的路径
	var sortedPaths []string
	for _, info := range pathInfos {
		sortedPaths = append(sortedPaths, info.path)
	}

	return sortedPaths
}

// Clear 清空所有备份信息
func (fr *FileRollback) Clear() error {
	fr.backupFiles = make(map[string]string)
	fr.createdFiles = fr.createdFiles[:0]
	return nil
}

// GetBackupCount 获取备份文件数量
func (fr *FileRollback) GetBackupCount() int {
	return len(fr.backupFiles)
}

// GetCreatedFileCount 获取新创建文件数量
func (fr *FileRollback) GetCreatedFileCount() int {
	return len(fr.createdFiles)
}

// HasBackup 检查指定文件是否已有备份
func (fr *FileRollback) HasBackup(filePath string) bool {
	_, exists := fr.backupFiles[filePath]
	return exists
}
