package services

import (
	"fmt"

	"brand-config-api/config"
	"brand-config-api/database"
	"brand-config-api/models"
	"brand-config-api/utils"

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
	db                     *gorm.DB
	config                 *config.Config
	fileService            *FileService
	configGeneratorService *ConfigGeneratorService
}

// NewWebsiteService 创建网站服务实例
func NewWebsiteService() *WebsiteService {
	return &WebsiteService{
		db:                     database.DB,
		config:                 config.Load(),
		fileService:            NewFileService(),
		configGeneratorService: NewConfigGeneratorService(),
	}
}

// CreateWebsite 创建网站（带回滚功能）
func (s *WebsiteService) CreateWebsite(req *CreateWebsiteRequest) (map[string]interface{}, error) {
	// 创建回滚管理器
	rollbackManager := utils.NewRollbackManager(s.config)

	// 回滚函数
	rollback := func(err error) error {
		fmt.Printf("❌ 创建失败，开始回滚: %v\n", err)
		if rollbackErr := rollbackManager.Rollback(); rollbackErr != nil {
			return fmt.Errorf("回滚失败: %v, 原始错误: %v", rollbackErr, err)
		}
		return err
	}

	// 开始事务
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			rollback(fmt.Errorf("panic: %v", r))
		}
	}()

	// 1. 验证品牌是否存在
	var brand models.Brand
	if err := tx.First(&brand, req.BasicInfo.BrandID).Error; err != nil {
		tx.Rollback()
		return nil, rollback(fmt.Errorf("brand not found"))
	}

	// 2. 验证host格式
	if req.BasicInfo.Host != "h5" && req.BasicInfo.Host != "tth5" && req.BasicInfo.Host != "ksh5" {
		tx.Rollback()
		return nil, rollback(fmt.Errorf("invalid host format"))
	}

	// 3. 检查client是否已存在
	var existingClient models.Client
	if err := tx.Where("brand_id = ? AND host = ?", req.BasicInfo.BrandID, req.BasicInfo.Host).First(&existingClient).Error; err == nil {
		tx.Rollback()
		return nil, rollback(fmt.Errorf("client already exists for this brand and host"))
	}

	// 4. 创建Client
	client := models.Client{
		BrandID: req.BasicInfo.BrandID,
		Host:    req.BasicInfo.Host,
	}
	if err := tx.Create(&client).Error; err != nil {
		tx.Rollback()
		return nil, rollback(fmt.Errorf("failed to create client: %v", err))
	}
	rollbackManager.AddCreatedClient(int(client.ID))
	fmt.Printf("✅ 创建Client成功: ID=%d\n", client.ID)

	// 5. 创建主BaseConfig
	baseConfig := models.BaseConfig{
		ClientID: client.ID,
		Platform: req.BaseConfig.Platform,
		AppName:  req.BaseConfig.AppName,
		AppCode:  req.BaseConfig.AppCode,
		Product:  req.BaseConfig.Product,
		Customer: req.BaseConfig.Customer,
		AppID:    req.BaseConfig.AppID,
		Version:  req.BaseConfig.Version,
		CL:       req.BaseConfig.CL,
		UC:       req.BaseConfig.UC,
	}
	if err := tx.Create(&baseConfig).Error; err != nil {
		tx.Rollback()
		return nil, rollback(fmt.Errorf("failed to create base config: %v", err))
	}
	rollbackManager.AddCreatedBaseConfig(int(baseConfig.ID))
	fmt.Printf("✅ 创建BaseConfig成功: ID=%d\n", baseConfig.ID)

	// 6. 如果需要额外的BaseConfig，创建额外的Client和BaseConfig
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

		// 创建额外的Client
		extraClient = &models.Client{
			BrandID: brand.ID,
			Host:    extraHost,
		}
		if err := tx.Create(extraClient).Error; err != nil {
			tx.Rollback()
			return nil, rollback(fmt.Errorf("failed to create extra client: %v", err))
		}
		rollbackManager.AddCreatedExtraClient(int(extraClient.ID))
		fmt.Printf("✅ 创建额外Client成功: ID=%d\n", extraClient.ID)

		// 创建额外的BaseConfig
		extraBaseConfig = &models.BaseConfig{
			ClientID: extraClient.ID,
			Platform: req.ExtraBaseConfig.Platform,
			AppName:  req.ExtraBaseConfig.AppName,
			AppCode:  req.ExtraBaseConfig.AppCode,
			Product:  req.ExtraBaseConfig.Product,
			Customer: req.ExtraBaseConfig.Customer,
			AppID:    req.ExtraBaseConfig.AppID,
			Version:  req.ExtraBaseConfig.Version,
			CL:       req.ExtraBaseConfig.CL,
			UC:       req.ExtraBaseConfig.UC,
		}
		if err := tx.Create(extraBaseConfig).Error; err != nil {
			tx.Rollback()
			return nil, rollback(fmt.Errorf("failed to create extra base config: %v", err))
		}
		rollbackManager.AddCreatedExtraBaseConfig(int(extraBaseConfig.ID))
		fmt.Printf("✅ 创建额外BaseConfig成功: ID=%d\n", extraBaseConfig.ID)
	}

	// 7. 创建CommonConfig
	commonConfig := models.CommonConfig{
		ClientID:                client.ID,
		DeliverBusinessIDEnable: req.CommonConfig.DeliverBusinessIDEnable,
		DeliverBusinessID:       req.CommonConfig.DeliverBusinessID,
		DeliverSwitchIDEnable:   req.CommonConfig.DeliverSwitchIDEnable,
		DeliverSwitchID:         req.CommonConfig.DeliverSwitchID,
		ProtocolCompany:         req.CommonConfig.ProtocolCompany,
		ProtocolAbout:           req.CommonConfig.ProtocolAbout,
		ProtocolPrivacy:         req.CommonConfig.ProtocolPrivacy,
		ProtocolVod:             req.CommonConfig.ProtocolVod,
		ProtocolUserCancel:      req.CommonConfig.ProtocolUserCancel,
		ContactURL:              req.CommonConfig.ContactURL,
		ScriptBase:              req.CommonConfig.ScriptBase,
	}
	if err := tx.Create(&commonConfig).Error; err != nil {
		tx.Rollback()
		return nil, rollback(fmt.Errorf("failed to create common config: %v", err))
	}
	rollbackManager.AddCreatedCommonConfig(int(commonConfig.ID))
	fmt.Printf("✅ 创建CommonConfig成功: ID=%d\n", commonConfig.ID)

	// 8. 创建PayConfig
	payConfig := models.PayConfig{
		ClientID:                client.ID,
		NormalPayEnable:         req.PayConfig.NormalPayEnable,
		NormalPayGatewayAndroid: req.PayConfig.NormalPayGatewayAndroid,
		NormalPayGatewayIOS:     req.PayConfig.NormalPayGatewayIOS,
		RenewPayEnable:          req.PayConfig.RenewPayEnable,
		RenewPayGatewayAndroid:  req.PayConfig.RenewPayGatewayAndroid,
		RenewPayGatewayIOS:      req.PayConfig.RenewPayGatewayIOS,
	}
	if err := tx.Create(&payConfig).Error; err != nil {
		tx.Rollback()
		return nil, rollback(fmt.Errorf("failed to create pay config: %v", err))
	}
	rollbackManager.AddCreatedPayConfig(int(payConfig.ID))
	fmt.Printf("✅ 创建PayConfig成功: ID=%d\n", payConfig.ID)

	// 9. 创建UIConfig
	uiConfig := models.UIConfig{
		ClientID:      client.ID,
		ThemeBgMain:   req.UIConfig.ThemeBgMain,
		ThemeBgSecond: req.UIConfig.ThemeBgSecond,
		ThemeTextMain: req.UIConfig.ThemeTextMain,
	}
	if err := tx.Create(&uiConfig).Error; err != nil {
		tx.Rollback()
		return nil, rollback(fmt.Errorf("failed to create UI config: %v", err))
	}
	rollbackManager.AddCreatedUIConfig(int(uiConfig.ID))
	fmt.Printf("✅ 创建UIConfig成功: ID=%d\n", uiConfig.ID)

	// 10. 创建NovelConfig（如果提供）
	var novelConfig *models.NovelConfig
	if req.NovelConfig != nil {
		novelConfig = &models.NovelConfig{
			ClientID:              client.ID,
			TTJumpHomeUrl:         req.NovelConfig.TTJumpHomeUrl,
			TTLoginCallbackDomain: req.NovelConfig.TTLoginCallbackDomain,
		}
		if err := tx.Create(novelConfig).Error; err != nil {
			tx.Rollback()
			return nil, rollback(fmt.Errorf("failed to create novel config: %v", err))
		}
		rollbackManager.AddCreatedNovelConfig(int(novelConfig.ID))
		fmt.Printf("✅ 创建NovelConfig成功: ID=%d\n", novelConfig.ID)
	} else {
		fmt.Printf("ℹ️ 未提供NovelConfig，跳过创建\n")
	}

	// 提交数据库事务
	if err := tx.Commit().Error; err != nil {
		return nil, rollback(fmt.Errorf("failed to commit transaction: %v", err))
	}
	fmt.Printf("✅ 数据库事务提交成功\n")

	// 10. 备份项目文件
	if err := s.fileService.BackupProjectFiles(brand.Code, rollbackManager); err != nil {
		return nil, rollback(fmt.Errorf("failed to backup project files: %v", err))
	}

	// 11. 生成配置文件
	if err := s.configGeneratorService.GenerateConfigFiles(brand.Code, client.Host, baseConfig, extraBaseConfig, commonConfig, payConfig, uiConfig, novelConfig, rollbackManager); err != nil {
		return nil, rollback(fmt.Errorf("failed to generate config files: %v", err))
	}

	// 12. 更新项目配置文件
	if err := s.fileService.UpdateProjectConfigs(brand.Code, client.Host, commonConfig.ScriptBase, baseConfig.AppName, rollbackManager); err != nil {
		return nil, rollback(fmt.Errorf("failed to update project configs: %v", err))
	}

	// 13. 创建prebuild文件
	if err := s.fileService.CreatePrebuildFiles(brand.Code, baseConfig.AppName, client.Host, rollbackManager); err != nil {
		return nil, rollback(fmt.Errorf("failed to create prebuild files: %v", err))
	}

	// 14. 创建static图片目录
	if err := s.fileService.CreateStaticImageDirectory(brand.Code, rollbackManager); err != nil {
		return nil, rollback(fmt.Errorf("failed to create static image directory: %v", err))
	}

	fmt.Printf("✅ 网站创建完成，所有操作成功\n")

	return map[string]interface{}{
		"client_id": client.ID,
		"brand_id":  brand.ID,
		"host":      client.Host,
	}, nil
}
