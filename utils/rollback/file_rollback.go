package rollback

import (
	"fmt"
	"log"
	"os"

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

	// 检查文件是否存在
	if _, err := os.Stat(path); err == nil {
		// 文件存在，读取原始内容作为备份
		originalContent, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read file for backup: %v", err)
		}
		fr.backupFiles[path] = string(originalContent)
		log.Printf("✅ 文件备份成功: %s", path)
	} else {
		// 文件不存在，标记为新创建的文件
		fr.createdFiles = append(fr.createdFiles, path)
		log.Printf("📝 标记为新创建文件: %s", path)
	}

	return nil
}

// Restore 恢复文件内容
func (fr *FileRollback) Restore(path string) error {
	log.Printf("🔄 文件回滚器：恢复文件 %s", path)

	// 检查是否有备份
	originalContent, hasBackup := fr.backupFiles[path]
	if hasBackup {
		// 恢复原始内容
		if err := os.WriteFile(path, []byte(originalContent), 0644); err != nil {
			return fmt.Errorf("failed to restore file: %v", err)
		}
		// log.Printf("✅ 文件恢复成功: %s", path)
		delete(fr.backupFiles, path)
	} else {
		// 检查是否是新创建的文件
		for i, createdFile := range fr.createdFiles {
			if createdFile == path {
				// 删除新创建的文件
				if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
					return fmt.Errorf("failed to remove created file: %v", err)
				}
				log.Printf("✅ 删除新创建文件: %s", path)
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

	var errors []error
	var successCount int

	// 恢复所有备份的文件
	for path := range fr.backupFiles {
		if err := fr.Restore(path); err != nil {
			errors = append(errors, fmt.Errorf("failed to restore %s: %v", path, err))
			log.Printf("❌ 文件恢复失败: %s - %v", path, err)
		} else {
			successCount++
			log.Printf("✅ 文件恢复成功: %s", path)
		}
	}

	// 删除所有新创建的文件
	for _, path := range fr.createdFiles {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			errors = append(errors, fmt.Errorf("failed to remove %s: %v", path, err))
			log.Printf("❌ 删除新创建文件失败: %s - %v", path, err)
		} else {
			successCount++
			log.Printf("✅ 删除新创建文件成功: %s", path)
		}
	}

	// 清空列表
	fr.createdFiles = fr.createdFiles[:0]

	// 详细的结果报告
	if len(errors) > 0 {
		log.Printf("⚠️ 文件回滚完成，但有 %d 个错误", len(errors))
		for i, err := range errors {
			log.Printf("  %d. %v", i+1, err)
		}
		return fmt.Errorf("file rollback completed with %d errors: %v", len(errors), errors)
	}

	log.Printf("✅ 文件回滚完成，成功处理 %d 个文件", successCount)
	return nil
}

// Clear 清空所有备份信息
func (fr *FileRollback) Clear() error {
	log.Printf("🧹 文件回滚器：清空所有备份信息")

	fr.backupFiles = make(map[string]string)
	fr.createdFiles = fr.createdFiles[:0]

	log.Printf("✅ 备份信息已清空")
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
