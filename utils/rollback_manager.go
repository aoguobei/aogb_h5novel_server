package utils

import (
	"fmt"
	"os"
	"path/filepath"

	"brand-config-api/config"
	"brand-config-api/database"
	"brand-config-api/models"
)

// RollbackManager 回滚管理器
type RollbackManager struct {
	createdResources *CreatedResources
	config           *config.Config
}

// CreatedResources 记录创建的资源，用于回滚
type CreatedResources struct {
	ClientID          int
	ExtraClientID     *int
	BaseConfigID      int
	ExtraBaseConfigID *int
	CommonConfigID    int
	PayConfigID       int
	UIConfigID        int
	NovelConfigID     int
	CreatedFiles      []string
	ModifiedFiles     map[string]string // 文件路径 -> 原始内容
}

// NewRollbackManager 创建回滚管理器
func NewRollbackManager(cfg *config.Config) *RollbackManager {
	return &RollbackManager{
		createdResources: &CreatedResources{
			ModifiedFiles: make(map[string]string),
		},
		config: cfg,
	}
}

// AddCreatedClient 添加创建的客户端
func (r *RollbackManager) AddCreatedClient(clientID int) {
	r.createdResources.ClientID = clientID
}

// AddCreatedExtraClient 添加创建的额外客户端
func (r *RollbackManager) AddCreatedExtraClient(clientID int) {
	r.createdResources.ExtraClientID = &clientID
}

// AddCreatedBaseConfig 添加创建的基础配置
func (r *RollbackManager) AddCreatedBaseConfig(configID int) {
	r.createdResources.BaseConfigID = configID
}

// AddCreatedExtraBaseConfig 添加创建的额外基础配置
func (r *RollbackManager) AddCreatedExtraBaseConfig(configID int) {
	r.createdResources.ExtraBaseConfigID = &configID
}

// AddCreatedCommonConfig 添加创建的通用配置
func (r *RollbackManager) AddCreatedCommonConfig(configID int) {
	r.createdResources.CommonConfigID = configID
}

// AddCreatedPayConfig 添加创建的支付配置
func (r *RollbackManager) AddCreatedPayConfig(configID int) {
	r.createdResources.PayConfigID = configID
}

// AddCreatedUIConfig 添加创建的UI配置
func (r *RollbackManager) AddCreatedUIConfig(configID int) {
	r.createdResources.UIConfigID = configID
}

// AddCreatedNovelConfig 添加创建的小说配置
func (r *RollbackManager) AddCreatedNovelConfig(configID int) {
	r.createdResources.NovelConfigID = configID
}

// AddCreatedFile 添加创建的文件
func (r *RollbackManager) AddCreatedFile(filePath string) {
	r.createdResources.CreatedFiles = append(r.createdResources.CreatedFiles, filePath)
}

// AddModifiedFile 添加修改的文件
func (r *RollbackManager) AddModifiedFile(filePath, originalContent string) {
	r.createdResources.ModifiedFiles[filePath] = originalContent
}

// Rollback 执行回滚操作
func (r *RollbackManager) Rollback() error {
	fmt.Printf("🔄 开始执行回滚操作...\n")

	// 1. 回滚文件操作
	if err := r.rollbackFileOperations(); err != nil {
		fmt.Printf("❌ 文件回滚失败: %v\n", err)
	}

	// 2. 回滚数据库操作
	if err := r.rollbackDatabaseOperations(); err != nil {
		fmt.Printf("❌ 数据库回滚失败: %v\n", err)
		return err
	}

	fmt.Printf("✅ 回滚操作完成\n")
	return nil
}

// rollbackFileOperations 回滚文件操作
func (r *RollbackManager) rollbackFileOperations() error {
	fmt.Printf("🔄 开始回滚文件操作...\n")

	// 1. 删除创建的文件
	for _, filePath := range r.createdResources.CreatedFiles {
		if err := os.Remove(filePath); err != nil {
			fmt.Printf("❌ 删除文件失败 %s: %v\n", filePath, err)
		} else {
			fmt.Printf("✅ 删除文件成功: %s\n", filePath)
		}
	}

	// 2. 恢复修改的文件
	for filePath, originalContent := range r.createdResources.ModifiedFiles {
		if err := os.WriteFile(filePath, []byte(originalContent), 0644); err != nil {
			fmt.Printf("❌ 恢复文件失败 %s: %v\n", filePath, err)
		} else {
			fmt.Printf("✅ 恢复文件成功: %s\n", filePath)
		}
	}

	fmt.Printf("✅ 文件操作回滚完成\n")
	return nil
}

// rollbackDatabaseOperations 回滚数据库操作
func (r *RollbackManager) rollbackDatabaseOperations() error {
	fmt.Printf("🔄 开始回滚数据库操作...\n")

	// 删除UIConfig
	if r.createdResources.UIConfigID > 0 {
		if err := database.DB.Delete(&models.UIConfig{}, r.createdResources.UIConfigID).Error; err != nil {
			fmt.Printf("❌ 删除UIConfig失败: %v\n", err)
		} else {
			fmt.Printf("✅ 删除UIConfig成功: ID=%d\n", r.createdResources.UIConfigID)
		}
	}

	// 删除NovelConfig
	if r.createdResources.NovelConfigID > 0 {
		if err := database.DB.Delete(&models.NovelConfig{}, r.createdResources.NovelConfigID).Error; err != nil {
			fmt.Printf("❌ 删除NovelConfig失败: %v\n", err)
		} else {
			fmt.Printf("✅ 删除NovelConfig成功: ID=%d\n", r.createdResources.NovelConfigID)
		}
	}

	// 删除PayConfig
	if r.createdResources.PayConfigID > 0 {
		if err := database.DB.Delete(&models.PayConfig{}, r.createdResources.PayConfigID).Error; err != nil {
			fmt.Printf("❌ 删除PayConfig失败: %v\n", err)
		} else {
			fmt.Printf("✅ 删除PayConfig成功: ID=%d\n", r.createdResources.PayConfigID)
		}
	}

	// 删除CommonConfig
	if r.createdResources.CommonConfigID > 0 {
		if err := database.DB.Delete(&models.CommonConfig{}, r.createdResources.CommonConfigID).Error; err != nil {
			fmt.Printf("❌ 删除CommonConfig失败: %v\n", err)
		} else {
			fmt.Printf("✅ 删除CommonConfig成功: ID=%d\n", r.createdResources.CommonConfigID)
		}
	}

	// 删除额外的BaseConfig
	if r.createdResources.ExtraBaseConfigID != nil && *r.createdResources.ExtraBaseConfigID > 0 {
		if err := database.DB.Delete(&models.BaseConfig{}, *r.createdResources.ExtraBaseConfigID).Error; err != nil {
			fmt.Printf("❌ 删除额外BaseConfig失败: %v\n", err)
		} else {
			fmt.Printf("✅ 删除额外BaseConfig成功: ID=%d\n", *r.createdResources.ExtraBaseConfigID)
		}
	}

	// 删除主BaseConfig
	if r.createdResources.BaseConfigID > 0 {
		if err := database.DB.Delete(&models.BaseConfig{}, r.createdResources.BaseConfigID).Error; err != nil {
			fmt.Printf("❌ 删除BaseConfig失败: %v\n", err)
		} else {
			fmt.Printf("✅ 删除BaseConfig成功: ID=%d\n", r.createdResources.BaseConfigID)
		}
	}

	// 删除额外的Client
	if r.createdResources.ExtraClientID != nil && *r.createdResources.ExtraClientID > 0 {
		if err := database.DB.Delete(&models.Client{}, *r.createdResources.ExtraClientID).Error; err != nil {
			fmt.Printf("❌ 删除额外Client失败: %v\n", err)
		} else {
			fmt.Printf("✅ 删除额外Client成功: ID=%d\n", *r.createdResources.ExtraClientID)
		}
	}

	// 删除主Client
	if r.createdResources.ClientID > 0 {
		if err := database.DB.Delete(&models.Client{}, r.createdResources.ClientID).Error; err != nil {
			fmt.Printf("❌ 删除Client失败: %v\n", err)
		} else {
			fmt.Printf("✅ 删除Client成功: ID=%d\n", r.createdResources.ClientID)
		}
	}

	fmt.Printf("✅ 数据库操作回滚完成\n")
	return nil
}

// BackupFile 备份文件内容
func (r *RollbackManager) BackupFile(filePath string) error {
	if _, err := os.Stat(filePath); err == nil {
		content, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %v", filePath, err)
		}
		r.AddModifiedFile(filePath, string(content))
	} else {
		// 文件不存在，标记为创建的文件
		r.AddCreatedFile(filePath)
	}
	return nil
}

// BackupConfigFile 备份配置文件
func (r *RollbackManager) BackupConfigFile(configType, brandCode string) error {
	var configDir string
	var fileName string

	switch configType {
	case "base":
		configDir = r.config.File.BaseConfigsDir
		fileName = brandCode + ".js"
	case "common":
		configDir = r.config.File.CommonConfigsDir
		fileName = brandCode + ".js"
	case "pay":
		configDir = r.config.File.PayConfigsDir
		fileName = brandCode + ".js"
	case "ui":
		configDir = r.config.File.UIConfigsDir
		fileName = brandCode + ".js"
	case "novel":
		configDir = r.config.File.LocalConfigsDir
		fileName = "novelConfig.js"
	default:
		configDir = r.config.File.ConfigDir
		fileName = brandCode + ".js"
	}

	configFile := filepath.Join(configDir, fileName)
	return r.BackupFile(configFile)
}

// BackupProjectFile 备份项目文件
func (r *RollbackManager) BackupProjectFile(filePath string) error {
	return r.BackupFile(filePath)
}
