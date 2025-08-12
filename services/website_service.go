package services

import (
	"fmt"

	"brand-config-api/config"
	"brand-config-api/database"
	"brand-config-api/models"
	"brand-config-api/utils"
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

		// 创建基础配置
		baseConfig, err := s.createBaseConfig(ctx.DB, req.BaseConfig, int(client.ID))
		if err != nil {
			return fmt.Errorf("failed to create base config: %v", err)
		}

		// 创建通用配置
		commonConfig, err := s.createCommonConfig(ctx.DB, req.CommonConfig, int(client.ID))
		if err != nil {
			return fmt.Errorf("failed to create common config: %v", err)
		}

		// 创建支付配置
		payConfig, err := s.createPayConfig(ctx.DB, req.PayConfig, int(client.ID))
		if err != nil {
			return fmt.Errorf("failed to create pay config: %v", err)
		}

		// 创建UI配置
		uiConfig, err := s.createUIConfig(ctx.DB, req.UIConfig, int(client.ID))
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

				// 创建额外的基础配置
				extraBaseConfig, err = s.createBaseConfig(ctx.DB, *req.ExtraBaseConfig, int(extraClient.ID))
				if err != nil {
					return fmt.Errorf("failed to create extra base config: %v", err)
				}
			}
		}

		// 创建小说配置（如果存在）
		var novelConfig *models.NovelConfig
		if req.NovelConfig != nil {
			novelConfig, err = s.createNovelConfig(ctx.DB, *req.NovelConfig, int(client.ID))
			if err != nil {
				return fmt.Errorf("failed to create novel config: %v", err)
			}
		}

		// 生成配置文件
		if err := s.generateConfigFiles(rollbackManager, client, baseConfig, commonConfig, payConfig, uiConfig, novelConfig, extraBaseConfig); err != nil {
			return fmt.Errorf("failed to generate config files: %v", err)
		}

		// 更新项目配置文件
		if err := s.fileService.UpdateProjectConfigs(client.Brand.Code, client.Host, commonConfig.ScriptBase, baseConfig.AppName, rollbackManager); err != nil {
			return fmt.Errorf("failed to update project configs: %v", err)
		}

		// 创建prebuild文件
		if err := s.fileService.CreatePrebuildFiles(client.Brand.Code, baseConfig.AppName, client.Host, rollbackManager); err != nil {
			return fmt.Errorf("failed to create prebuild files: %v", err)
		}

		// 创建static图片目录
		if err := s.fileService.CreateStaticImageDirectory(client.Brand.Code, rollbackManager); err != nil {
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

// createBaseConfig 创建基础配置
func (s *WebsiteService) createBaseConfig(tx *gorm.DB, baseConfigReq BaseConfigRequest, clientID int) (*models.BaseConfig, error) {
	// 使用BaseConfigService的带事务方法
	baseConfigService := NewBaseConfigService()
	return baseConfigService.CreateBaseConfigFromRequestWithTx(tx, baseConfigReq, clientID)
}

// createCommonConfig 创建通用配置
func (s *WebsiteService) createCommonConfig(tx *gorm.DB, commonConfigReq CommonConfigRequest, clientID int) (*models.CommonConfig, error) {
	// 使用CommonConfigService的带事务方法
	commonConfigService := NewCommonConfigService()
	return commonConfigService.CreateCommonConfigFromRequestWithTx(tx, commonConfigReq, clientID)
}

// createPayConfig 创建支付配置
func (s *WebsiteService) createPayConfig(tx *gorm.DB, payConfigReq PayConfigRequest, clientID int) (*models.PayConfig, error) {
	// 使用PayConfigService的带事务方法
	payConfigService := NewPayConfigService()
	return payConfigService.CreatePayConfigFromRequestWithTx(tx, payConfigReq, clientID)
}

// createUIConfig 创建UI配置
func (s *WebsiteService) createUIConfig(tx *gorm.DB, uiConfigReq UIConfigRequest, clientID int) (*models.UIConfig, error) {
	// 使用UIConfigService的带事务方法
	uiConfigService := NewUIConfigService()
	return uiConfigService.CreateUIConfigFromRequestWithTx(tx, uiConfigReq, clientID)
}

// createNovelConfig 创建小说配置
func (s *WebsiteService) createNovelConfig(tx *gorm.DB, novelConfigReq NovelConfigRequest, clientID int) (*models.NovelConfig, error) {
	// 使用NovelConfigService的带事务方法
	novelConfigService := NewNovelConfigService()
	return novelConfigService.CreateNovelConfigFromRequestWithTx(tx, novelConfigReq, clientID)
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

// generateConfigFiles 生成配置文件
func (s *WebsiteService) generateConfigFiles(rollbackManager *rollback.RollbackManager, client *models.Client, baseConfig *models.BaseConfig, commonConfig *models.CommonConfig, payConfig *models.PayConfig, uiConfig *models.UIConfig, novelConfig *models.NovelConfig, extraBaseConfig *models.BaseConfig) error {
	// 确保有品牌代码
	if client.Brand.Code == "" {
		return fmt.Errorf("brand code is empty for client %d", client.ID)
	}

	fmt.Printf("📁 开始生成配置文件: brand=%s, host=%s\n", client.Brand.Code, client.Host)

	// 创建配置文件写入工具
	configfileManager := utils.NewConfigFileManager()

	// 创建各种配置服务实例
	baseConfigService := NewBaseConfigService()
	commonConfigService := NewCommonConfigService()
	payConfigService := NewPayConfigService()
	uiConfigService := NewUIConfigService()
	novelConfigService := NewNovelConfigService()

	// 生成主BaseConfig文件
	formattedBaseConfig := baseConfigService.FormatBaseConfig(*baseConfig)
	if err := configfileManager.WriteConfigToFile("base", formattedBaseConfig, client.Brand.Code, client.Host, rollbackManager); err != nil {
		return fmt.Errorf("failed to write base config file: %v", err)
	}

	// 生成额外的BaseConfig文件（如果存在）
	if extraBaseConfig != nil {
		var extraHost string
		if client.Host == "tth5" {
			extraHost = "tt"
		} else if client.Host == "ksh5" {
			extraHost = "ks"
		}

		formattedExtraBaseConfig := baseConfigService.FormatBaseConfig(*extraBaseConfig)
		if err := configfileManager.WriteConfigToFile("base", formattedExtraBaseConfig, client.Brand.Code, extraHost, rollbackManager); err != nil {
			return fmt.Errorf("failed to write extra base config file: %v", err)
		}
	}

	// 写入其他配置文件
	formattedCommonConfig := commonConfigService.FormatCommonConfig(*commonConfig)
	if err := configfileManager.WriteConfigToFile("common", formattedCommonConfig, client.Brand.Code, client.Host, rollbackManager); err != nil {
		return fmt.Errorf("failed to write common config file: %v", err)
	}

	formattedPayConfig := payConfigService.FormatPayConfig(*payConfig)
	if err := configfileManager.WriteConfigToFile("pay", formattedPayConfig, client.Brand.Code, client.Host, rollbackManager); err != nil {
		return fmt.Errorf("failed to write pay config file: %v", err)
	}

	formattedUIConfig := uiConfigService.FormatUIConfig(*uiConfig)
	if err := configfileManager.WriteConfigToFile("ui", formattedUIConfig, client.Brand.Code, client.Host, rollbackManager); err != nil {
		return fmt.Errorf("failed to write ui config file: %v", err)
	}

	// 生成NovelConfig文件（如果存在）
	if novelConfig != nil {
		formattedNovelConfig := novelConfigService.FormatNovelConfig(*novelConfig)
		if err := configfileManager.WriteConfigToFile("novel", formattedNovelConfig, client.Brand.Code, client.Host, rollbackManager); err != nil {
			return fmt.Errorf("failed to write novel config file: %v", err)
		}
	}

	fmt.Printf("✅ 配置文件生成完成: brand=%s, host=%s\n", client.Brand.Code, client.Host)
	return nil
}
