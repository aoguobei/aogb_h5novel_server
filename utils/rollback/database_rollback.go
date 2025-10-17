package rollback

import (
	"log"

	"gorm.io/gorm"
)

// DatabaseRollbacker 数据库回滚接口
type DatabaseRollbacker interface {
	Begin() *gorm.DB
	Commit(tx *gorm.DB) error
	Rollback(tx *gorm.DB) error
}

// DatabaseRollback 数据库回滚实现
type DatabaseRollback struct {
	db *gorm.DB
}

// NewDatabaseRollback 创建数据库回滚实例
func NewDatabaseRollback(db *gorm.DB) *DatabaseRollback {
	return &DatabaseRollback{
		db: db,
	}
}

// Begin 开始数据库事务
func (dr *DatabaseRollback) Begin() *gorm.DB {
	log.Printf("🔄 数据库回滚器：开始事务")
	return dr.db.Begin()
}

// Commit 提交数据库事务
func (dr *DatabaseRollback) Commit(tx *gorm.DB) error {
	log.Printf("🔄 数据库回滚器：提交事务")
	return tx.Commit().Error
}

// Rollback 回滚数据库事务
func (dr *DatabaseRollback) Rollback(tx *gorm.DB) error {
	log.Printf("🔄 数据库回滚器：回滚事务")
	return tx.Rollback().Error
}

// IsInTransaction 检查是否在事务中
func (dr *DatabaseRollback) IsInTransaction(tx *gorm.DB) bool {
	return tx != nil && tx.Error == nil
}

// GetTransaction 获取当前事务
func (dr *DatabaseRollback) GetTransaction() *gorm.DB {
	return dr.db
}
