package rollback

import (
	"fmt"
	"log"

	"brand-config-api/config"

	"gorm.io/gorm"
)

// RollbackManager 回滚管理器
type RollbackManager struct {
	dbManager   *DatabaseRollback
	fileManager *FileRollback
}

// NewRollbackManager 创建回滚管理器
func NewRollbackManager(db *gorm.DB, cfg *config.Config) *RollbackManager {
	return &RollbackManager{
		dbManager:   NewDatabaseRollback(db),
		fileManager: NewFileRollback(cfg),
	}
}

// GetFileManager 获取文件回滚管理器
func (rm *RollbackManager) GetFileManager() *FileRollback {
	return rm.fileManager
}

// TransactionContext 事务上下文
type TransactionContext struct {
	DB    *gorm.DB
	Files *FileRollback
}

// ExecuteWithTransaction 执行事务操作
func (rm *RollbackManager) ExecuteWithTransaction(operation func(*TransactionContext) error) error {
	tx := rm.dbManager.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin database transaction: %v", tx.Error)
	}
	log.Printf("🔄 数据库事务开始")

	ctx := &TransactionContext{
		DB:    tx,
		Files: rm.fileManager,
	}

	defer func() {
		if r := recover(); r != nil {
			log.Printf("❌ 发生panic: %v", r)
			log.Printf("🔄 开始回滚数据库事务...")
			if rollbackErr := tx.Rollback().Error; rollbackErr != nil {
				log.Printf("❌ 数据库回滚失败: %v", rollbackErr)
			} else {
				log.Printf("✅ 数据库事务回滚成功")
			}
		}
	}()

	if err := operation(ctx); err != nil {
		log.Printf("❌ 操作失败，开始回滚: %v", err)
		log.Printf("🔄 开始回滚数据库事务...")
		if rollbackErr := tx.Rollback().Error; rollbackErr != nil {
			log.Printf("❌ 数据库回滚失败: %v", rollbackErr)
		} else {
			log.Printf("✅ 数据库事务回滚成功")
		}
		log.Printf("🔄 开始回滚文件操作...")
		fileRollbackErr := ctx.Files.Rollback()
		if fileRollbackErr != nil {
			log.Printf("❌ 文件回滚失败: %v", fileRollbackErr)
		} else {
			log.Printf("✅ 文件操作回滚成功")
		}
		return fmt.Errorf("操作失败，已回滚: %v", err)
	}

	log.Printf("🔄 开始提交数据库事务...")
	if err := tx.Commit().Error; err != nil {
		log.Printf("❌ 数据库事务提交失败: %v", err)
		return fmt.Errorf("failed to commit transaction: %v", err)
	}
	log.Printf("✅ 数据库事务提交成功")
	return nil
}
