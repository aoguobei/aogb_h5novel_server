package services

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"brand-config-api/config"
	"brand-config-api/database"
	"brand-config-api/models"
	"brand-config-api/utils/rollback"

	"gorm.io/gorm"
)

// CreateWebsiteRequest 创建网站的请求结构
type CreateWebsiteRequest struct {
	BasicInfo       BasicInfoRequest    `json:"basic_info"`
	BaseConfig      BaseConfigRequest   `json:"base_config"`
	ExtraBaseConfig *BaseConfigRequest  `json:"extra_base_config"`
	CommonConfig    CommonConfigRequest `json:"common_config"`
	PayConfig       PayConfigRequest    `json:"pay_config"`
	UIConfig        UIConfigRequest     `json:"ui_config"`
	NovelConfig     *NovelConfigRequest `json:"novel_config"`
}

type BasicInfoRequest struct {
	BrandID int    `json:"brand_id"`
	Host    string `json:"host"`
}

type BaseConfigRequest struct {
	AppName  string `json:"app_name"`
	Platform string `json:"platform"`
	AppCode  string `json:"app_code"`
	Product  string `json:"product"`
	Customer string `json:"customer"`
	AppID    string `json:"appid"`
	Version  string `json:"version"`
	CL       string `json:"cl"`
	UC       string `json:"uc"`
}

type CommonConfigRequest struct {
	DeliverBusinessIDEnable bool   `json:"deliver_business_id_enable"`
	DeliverBusinessID       string `json:"deliver_business_id"`
	DeliverSwitchIDEnable   bool   `json:"deliver_switch_id_enable"`
	DeliverSwitchID         string `json:"deliver_switch_id"`
	ProtocolCompany         string `json:"protocol_company"`
	ProtocolAbout           string `json:"protocol_about"`
	ProtocolPrivacy         string `json:"protocol_privacy"`
	ProtocolVod             string `json:"protocol_vod"`
	ProtocolUserCancel      string `json:"protocol_user_cancel"`
	ContactURL              string `json:"contact_url"`
	ScriptBase              string `json:"script_base"`
}

type PayConfigRequest struct {
	NormalPayEnable         bool `json:"normal_pay_enable"`
	NormalPayGatewayAndroid *int `json:"normal_pay_gateway_android"`
	NormalPayGatewayIOS     *int `json:"normal_pay_gateway_ios"`
	RenewPayEnable          bool `json:"renew_pay_enable"`
	RenewPayGatewayAndroid  *int `json:"renew_pay_gateway_android"`
	RenewPayGatewayIOS      *int `json:"renew_pay_gateway_ios"`
}

type UIConfigRequest struct {
	ThemeBgMain   string  `json:"theme_bg_main"`
	ThemeBgSecond string  `json:"theme_bg_second"`
	ThemeTextMain *string `json:"theme_text_main"`
}

type NovelConfigRequest struct {
	TTJumpHomeUrl         string `json:"tt_jump_home_url"`
	TTLoginCallbackDomain string `json:"tt_login_callback_domain"`
}

// WebsiteService 网站服务
type WebsiteService struct {
	db          *gorm.DB
	config      *config.Config
	fileService *FileService
}

// NewWebsiteService 创建网站服务实例
func NewWebsiteService() *WebsiteService {
	return &WebsiteService{
		db:          database.DB,
		config:      config.Load(),
		fileService: NewFileService(),
	}
}

// CreateWebsite 创建网站
func (s *WebsiteService) CreateWebsite(req *CreateWebsiteRequest) (map[string]interface{}, error) {
	// 创建回滚管理器
	rollbackManager := rollback.NewRollbackManager(s.db, s.config)

	var result map[string]interface{}

	err := rollbackManager.ExecuteWithTransaction(func(ctx *rollback.TransactionContext) error {
		// 创建客户端
		client, err := s.createClient(ctx.DB, req.BasicInfo)
		if err != nil {
			return fmt.Errorf("failed to create client: %v", err)
		}

		// 创建基础配置（包含文件生成）
		baseConfigService := NewBaseConfigService()
		baseConfig, err := baseConfigService.CreateBaseConfigWithFile(ctx, req.BaseConfig, int(client.ID), client.Brand.Code, client.Host)
		if err != nil {
			return fmt.Errorf("failed to create base config: %v", err)
		}

		// 创建通用配置（包含文件生成）
		commonConfigService := NewCommonConfigService()
		commonConfig, err := commonConfigService.CreateCommonConfigWithFile(ctx, req.CommonConfig, int(client.ID), client.Brand.Code, client.Host)
		if err != nil {
			return fmt.Errorf("failed to create common config: %v", err)
		}

		// 创建支付配置（包含文件生成）
		payConfigService := NewPayConfigService()
		payConfig, err := payConfigService.CreatePayConfigWithFile(ctx, req.PayConfig, int(client.ID), client.Brand.Code, client.Host)
		if err != nil {
			return fmt.Errorf("failed to create pay config: %v", err)
		}

		// 创建UI配置（包含文件生成）
		uiConfigService := NewUIConfigService()
		uiConfig, err := uiConfigService.CreateUIConfigWithFile(ctx, req.UIConfig, int(client.ID), client.Brand.Code, client.Host)
		if err != nil {
			return fmt.Errorf("failed to create UI config: %v", err)
		}

		// 创建额外客户端和基础配置（如果需要）
		var extraClient *models.Client
		var extraBaseConfig *models.BaseConfig
		if req.ExtraBaseConfig != nil {
			// 确定额外的host类型
			var extraHost string
			if client.Host == "tth5" {
				extraHost = "tt"
			} else if client.Host == "ksh5" {
				extraHost = "ks"
			}

			if extraHost != "" {
				// 创建额外的客户端
				extraClient, err = s.createExtraClient(ctx.DB, client.Brand.ID, extraHost)
				if err != nil {
					return fmt.Errorf("failed to create extra client: %v", err)
				}

				// 创建额外的基础配置（包含文件生成）
				extraBaseConfig, err = baseConfigService.CreateBaseConfigWithFile(ctx, *req.ExtraBaseConfig, int(extraClient.ID), client.Brand.Code, extraHost)
				if err != nil {
					return fmt.Errorf("failed to create extra base config: %v", err)
				}
			}
		}

		// 创建小说配置（如果存在，包含文件生成）
		var novelConfig *models.NovelConfig
		if req.NovelConfig != nil {
			novelConfigService := NewNovelConfigService()
			novelConfig, err = novelConfigService.CreateNovelConfigWithFile(ctx, *req.NovelConfig, int(client.ID), client.Brand.Code, client.Host)
			if err != nil {
				return fmt.Errorf("failed to create novel config: %v", err)
			}
		}

		// 更新项目配置文件
		if err := s.fileService.UpdateProjectConfigs(client.Brand.Code, client.Host, commonConfig.ScriptBase, baseConfig.AppName, ctx.Files); err != nil {
			return fmt.Errorf("failed to update project configs: %v", err)
		}

		// 创建prebuild文件
		if err := s.fileService.CreatePrebuildFiles(client.Brand.Code, baseConfig.AppName, client.Host, ctx.Files); err != nil {
			return fmt.Errorf("failed to create prebuild files: %v", err)
		}

		// 创建static图片目录
		if err := s.fileService.CreateStaticImageDirectory(client.Brand.Code, ctx.Files); err != nil {
			return fmt.Errorf("failed to create static image directory: %v", err)
		}

		// 构建返回结果
		result = map[string]interface{}{
			"client_id":        client.ID,
			"base_config_id":   baseConfig.ID,
			"common_config_id": commonConfig.ID,
			"pay_config_id":    payConfig.ID,
			"ui_config_id":     uiConfig.ID,
		}

		if novelConfig != nil {
			result["novel_config_id"] = novelConfig.ID
		}

		if extraClient != nil {
			result["extra_client_id"] = extraClient.ID
		}

		if extraBaseConfig != nil {
			result["extra_base_config_id"] = extraBaseConfig.ID
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetWebsiteConfig 获取网站完整配置
func (s *WebsiteService) GetWebsiteConfig(clientID int) (map[string]interface{}, error) {
	// 查询Client信息
	var client models.Client
	if err := s.db.Preload("Brand").First(&client, clientID).Error; err != nil {
		return nil, err
	}

	// 查询BaseConfig
	var baseConfig models.BaseConfig
	s.db.Where("client_id = ?", clientID).First(&baseConfig)

	// 查询CommonConfig
	var commonConfig models.CommonConfig
	s.db.Where("client_id = ?", clientID).First(&commonConfig)

	// 查询PayConfig
	var payConfig models.PayConfig
	s.db.Where("client_id = ?", clientID).First(&payConfig)

	// 查询UIConfig
	var uiConfig models.UIConfig
	s.db.Where("client_id = ?", clientID).First(&uiConfig)

	// 查询NovelConfig
	var novelConfig models.NovelConfig
	if err := s.db.Where("client_id = ?", clientID).First(&novelConfig).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// NovelConfig不存在，使用空结构
			novelConfig = models.NovelConfig{}
		} else {
			return nil, err
		}
	}

	// 添加调试日志
	log.Printf("🔍 NovelConfig 查询结果: ID=%d, ClientID=%d, TTJumpHomeUrl=%s, TTLoginCallbackDomain=%s",
		novelConfig.ID, novelConfig.ClientID, novelConfig.TTJumpHomeUrl, novelConfig.TTLoginCallbackDomain)

	// 构建响应数据
	response := map[string]interface{}{
		"client": map[string]interface{}{
			"id":         client.ID,
			"host":       client.Host,
			"created_at": client.CreatedAt,
			"updated_at": client.UpdatedAt,
			"brand": map[string]interface{}{
				"id":   client.Brand.ID,
				"code": client.Brand.Code,
			},
		},
		"base_config":   baseConfig,
		"common_config": commonConfig,
		"pay_config":    payConfig,
		"ui_config":     uiConfig,
		"novel_config":  novelConfig,
	}

	return response, nil
}

// createClient 创建客户端
func (s *WebsiteService) createClient(tx *gorm.DB, basicInfo BasicInfoRequest) (*models.Client, error) {
	// 使用ClientService在事务中创建客户端
	clientService := NewClientService()
	return clientService.CreateClientWithTx(tx, basicInfo.BrandID, basicInfo.Host)
}

// createExtraClient 创建额外客户端
func (s *WebsiteService) createExtraClient(tx *gorm.DB, brandID int, extraHost string) (*models.Client, error) {
	client := &models.Client{
		BrandID: brandID,
		Host:    extraHost,
	}

	if err := tx.Create(client).Error; err != nil {
		return nil, err
	}

	return client, nil
}

// DeleteWebsite 删除网站（原子操作）
func (s *WebsiteService) DeleteWebsite(clientID int) error {
	// 创建回滚管理器
	rollbackManager := rollback.NewRollbackManager(s.db, s.config)

	return rollbackManager.ExecuteWithTransaction(func(ctx *rollback.TransactionContext) error {
		// 获取客户端信息
		var client models.Client
		if err := ctx.DB.Preload("Brand").Where("id = ?", clientID).First(&client).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return fmt.Errorf("client with ID %d not found", clientID)
			}
			return fmt.Errorf("failed to find client: %v", err)
		}

		log.Printf("🗑️ 开始删除网站: clientID=%d, brand=%s, host=%s", clientID, client.Brand.Code, client.Host)

		// 1. 删除基础配置
		baseConfigService := NewBaseConfigService()
		if err := baseConfigService.deleteBaseConfigInternal(ctx, clientID); err != nil {
			return fmt.Errorf("failed to delete base config: %v", err)
		}

		// 2. 删除通用配置
		commonConfigService := NewCommonConfigService()
		if err := commonConfigService.deleteCommonConfigInternal(ctx, clientID); err != nil {
			return fmt.Errorf("failed to delete common config: %v", err)
		}

		// 3. 删除支付配置
		payConfigService := NewPayConfigService()
		if err := payConfigService.deletePayConfigInternal(ctx, clientID); err != nil {
			return fmt.Errorf("failed to delete pay config: %v", err)
		}

		// 4. 删除UI配置
		uiConfigService := NewUIConfigService()
		if err := uiConfigService.deleteUIConfigInternal(ctx, clientID); err != nil {
			return fmt.Errorf("failed to delete UI config: %v", err)
		}

		// 5. 删除小说配置（如果存在）
		novelConfigService := NewNovelConfigService()
		if err := novelConfigService.deleteNovelConfigInternal(ctx, clientID); err != nil {
			// 小说配置可能不存在，忽略错误
			log.Printf("⚠️ 小说配置删除失败（可能不存在）: %v", err)
		}

		// 6. 删除客户端
		if err := ctx.DB.Delete(&client).Error; err != nil {
			return fmt.Errorf("failed to delete client: %v", err)
		}
		log.Printf("✅ 客户端删除成功")

		// 8. 删除相关文件（项目配置文件、prebuild文件等）
		if err := s.deleteWebsiteFiles(ctx.Files, client.Brand.Code, client.Host); err != nil {
			return fmt.Errorf("failed to delete website files: %v", err)
		}

		// 9. 检查是否是该品牌的最后一个客户端，如果是则删除品牌相关文件
		if err := s.checkAndDeleteBrandFiles(ctx, client.Brand.Code, client.Host); err != nil {
			return fmt.Errorf("failed to check and delete brand files: %v", err)
		}

		log.Printf("✅ 网站删除成功: clientID=%d, brand=%s, host=%s", clientID, client.Brand.Code, client.Host)
		return nil
	})
}

// deleteWebsiteFiles 删除网站相关文件
func (s *WebsiteService) deleteWebsiteFiles(fileManager *rollback.FileRollback, brandCode, host string) error {
	log.Printf("🗑️ 开始删除网站文件: brand=%s, host=%s", brandCode, host)

	// 删除项目配置文件中的相关配置（vite.config.js, package.json, pages-host.json, novelconfig.js）
	if err := s.fileService.RemoveProjectConfigs(brandCode, host, fileManager); err != nil {
		return fmt.Errorf("failed to remove project configs: %v", err)
	}

	// 删除prebuild文件
	prebuildDir := filepath.Join(s.config.File.PrebuildDir, brandCode, host)
	if err := s.deleteDirectoryIfExists(fileManager, prebuildDir); err != nil {
		return fmt.Errorf("failed to delete prebuild directory: %v", err)
	}

	// 删除static图片目录（如果为空）
	staticImageDir := filepath.Join(s.config.File.StaticDir, brandCode)
	if err := s.deleteDirectoryIfEmpty(fileManager, staticImageDir); err != nil {
		log.Printf("⚠️ 删除static图片目录失败（可能不为空）: %v", err)
	}

	log.Printf("✅ 网站文件删除完成: brand=%s, host=%s", brandCode, host)
	return nil
}

// deleteBrandRecord 删除品牌数据库记录
func (s *WebsiteService) deleteBrandRecord(ctx *rollback.TransactionContext, brandCode string) error {
	// 查找品牌记录
	var brand models.Brand
	if err := ctx.DB.Where("code = ?", brandCode).First(&brand).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Printf("⚠️ 品牌记录不存在: %s", brandCode)
			return nil
		}
		return fmt.Errorf("failed to find brand record: %v", err)
	}

	// 删除品牌记录
	if err := ctx.DB.Delete(&brand).Error; err != nil {
		return fmt.Errorf("failed to delete brand record: %v", err)
	}

	log.Printf("✅ 品牌数据库记录删除成功: %s", brandCode)
	return nil
}

// deleteFileIfExists 删除文件（如果存在）
func (s *WebsiteService) deleteFileIfExists(fileManager *rollback.FileRollback, filePath string) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Printf("⚠️ 文件不存在，跳过删除: %s", filePath)
		return nil
	}

	// 备份文件
	if err := fileManager.Backup(filePath, ""); err != nil {
		return fmt.Errorf("failed to backup file: %v", err)
	}

	// 删除文件
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to delete file: %v", err)
	}

	log.Printf("🗑️ 文件删除成功: %s", filePath)
	return nil
}

// deleteDirectoryIfExists 删除目录（如果存在）
func (s *WebsiteService) deleteDirectoryIfExists(fileManager *rollback.FileRollback, dirPath string) error {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		log.Printf("⚠️ 目录不存在，跳过删除: %s", dirPath)
		return nil
	}

	// 备份目录
	if err := fileManager.Backup(dirPath, ""); err != nil {
		return fmt.Errorf("failed to backup directory: %v", err)
	}

	// 删除目录
	if err := os.RemoveAll(dirPath); err != nil {
		return fmt.Errorf("failed to delete directory: %v", err)
	}

	log.Printf("🗑️ 目录删除成功: %s", dirPath)
	return nil
}

// deleteDirectoryIfEmpty 删除目录（如果为空）
func (s *WebsiteService) deleteDirectoryIfEmpty(fileManager *rollback.FileRollback, dirPath string) error {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		log.Printf("⚠️ 目录不存在，跳过删除: %s", dirPath)
		return nil
	}

	// 检查目录是否为空
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("failed to read directory: %v", err)
	}

	if len(entries) > 0 {
		log.Printf("⚠️ 目录不为空，跳过删除: %s", dirPath)
		return nil
	}

	// 备份目录
	if err := fileManager.Backup(dirPath, ""); err != nil {
		return fmt.Errorf("failed to backup directory: %v", err)
	}

	// 删除目录
	if err := os.Remove(dirPath); err != nil {
		return fmt.Errorf("failed to delete directory: %v", err)
	}

	log.Printf("🗑️ 空目录删除成功: %s", dirPath)
	return nil
}

// checkAndDeleteBrandFiles 检查是否是该品牌的最后一个客户端，如果是则删除品牌相关文件
func (s *WebsiteService) checkAndDeleteBrandFiles(ctx *rollback.TransactionContext, brandCode, host string) error {
	// 检查该品牌是否还有其他客户端
	var count int64
	if err := ctx.DB.Model(&models.Client{}).Joins("JOIN brands ON clients.brand_id = brands.id").Where("brands.code = ?", brandCode).Count(&count).Error; err != nil {
		return fmt.Errorf("failed to check remaining clients: %v", err)
	}

	if count > 0 {
		log.Printf("⚠️ 品牌 %s 还有其他 %d 个客户端，跳过删除品牌文件", brandCode, count)
		return nil
	}

	log.Printf("🗑️ 品牌 %s 没有其他客户端，开始删除品牌相关文件", brandCode)

	// 删除品牌相关文件
	if err := s.fileService.RemoveBrandFiles(brandCode, host, ctx.Files); err != nil {
		return fmt.Errorf("failed to remove brand files: %v", err)
	}

	// 删除品牌数据库记录
	if err := s.deleteBrandRecord(ctx, brandCode); err != nil {
		return fmt.Errorf("failed to delete brand record: %v", err)
	}

	log.Printf("✅ 品牌相关文件和数据库记录删除完成: %s", brandCode)
	return nil
}
