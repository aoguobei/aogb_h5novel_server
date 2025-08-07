package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"brand-config-api/database"
	"brand-config-api/models"

	"errors"

	"github.com/gin-gonic/gin"
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

// CreateWebsite 批量创建网站
func CreateWebsite(c *gin.Context) {
	var req CreateWebsiteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 开始事务
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. 验证品牌是否存在
	var brand models.Brand
	if err := tx.First(&brand, req.BasicInfo.BrandID).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"error": "Brand not found"})
		return
	}

	// 2. 验证host格式
	if req.BasicInfo.Host != "h5" && req.BasicInfo.Host != "tth5" && req.BasicInfo.Host != "ksh5" {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid host format"})
		return
	}

	// 3. 检查client是否已存在
	var existingClient models.Client
	if err := tx.Where("brand_id = ? AND host = ?", req.BasicInfo.BrandID, req.BasicInfo.Host).First(&existingClient).Error; err == nil {
		tx.Rollback()
		c.JSON(http.StatusConflict, gin.H{"error": "Client already exists for this brand and host"})
		return
	}

	// 4. 创建Client
	client := models.Client{
		BrandID: req.BasicInfo.BrandID,
		Host:    req.BasicInfo.Host,
	}
	if err := tx.Create(&client).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create client"})
		return
	}

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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create base config"})
		return
	}

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
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create extra client"})
			return
		}

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
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create extra base config"})
			return
		}
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create common config"})
		return
	}

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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create pay config"})
		return
	}

	// 9. 创建UIConfig
	uiConfig := models.UIConfig{
		ClientID:      client.ID,
		ThemeBgMain:   req.UIConfig.ThemeBgMain,
		ThemeBgSecond: req.UIConfig.ThemeBgSecond,
		ThemeTextMain: req.UIConfig.ThemeTextMain,
	}
	if err := tx.Create(&uiConfig).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create UI config"})
		return
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	// 10. 生成主BaseConfig文件
	if err := writeBaseConfigToFile(baseConfig, brand.Code, client.Host); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write base config file: " + err.Error()})
		return
	}

	// 11. 生成额外的BaseConfig文件（如果存在）
	if extraBaseConfig != nil && extraClient != nil {
		if err := writeBaseConfigToFile(*extraBaseConfig, brand.Code, extraClient.Host); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write extra base config file: " + err.Error()})
			return
		}
	}

	// 12. 写入其他配置文件
	formattedCommonConfig := formatCommonConfig(commonConfig)
	if err := writeConfigToFile("common", formattedCommonConfig, brand.Code, client.Host); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write common config file: " + err.Error()})
		return
	}

	formattedPayConfig := formatPayConfig(payConfig)
	if err := writeConfigToFile("pay", formattedPayConfig, brand.Code, client.Host); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write pay config file: " + err.Error()})
		return
	}

	formattedUIConfig := formatUIConfig(uiConfig)
	if err := writeConfigToFile("ui", formattedUIConfig, brand.Code, client.Host); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write UI config file: " + err.Error()})
		return
	}

	// 13. 更新项目配置文件
	if err := updateProjectConfigs(brand.Code, client.Host, commonConfig.ScriptBase); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update project configs: " + err.Error()})
		return
	}

	// 14. 更新 index.js 文件，添加新的配置导入
	if err := updateIndexFileForConfig(brand.Code); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update index.js: " + err.Error()})
		return
	}

	// 15. 创建prebuild文件
	if err := createPrebuildFiles(brand.Code, baseConfig.AppName, client.Host); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create prebuild files: " + err.Error()})
		return
	}

	// 16. 创建static图片目录
	if err := createStaticImageDirectory(brand.Code); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create static image directory: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Website created successfully",
		"data": gin.H{
			"client_id": client.ID,
			"brand_id":  brand.ID,
			"host":      client.Host,
		},
	})
}

// updateProjectConfigs 更新项目配置文件
func updateProjectConfigs(brandCode, host string, scriptBase string) error {
	// 3. 更新 vite.config.js 文件（对所有新配置）
	if err := updateViteConfigFile(brandCode, host, scriptBase); err != nil {
		return fmt.Errorf("failed to update vite.config.js: %v", err)
	}

	// 4. 更新 package.json 文件（对所有新配置）
	if err := updatePackageJSONFile(brandCode, host); err != nil {
		return fmt.Errorf("failed to update package.json: %v", err)
	}

	return nil
}

// updateHostFile 更新 _host.js 文件，添加新的 brand
func updateHostFile(brandCode string) error {
	fmt.Printf("🔄 Starting updateHostFile for brand: %s\n", brandCode)

	hostFilePath := "C:/F_explorer/h5projects/jianruiH5/novel_h5config/funNovel/src/appConfig/_host.js"

	content, err := os.ReadFile(hostFilePath)
	if err != nil {
		fmt.Printf("❌ Failed to read _host.js: %v\n", err)
		return fmt.Errorf("failed to read _host.js: %v", err)
	}

	contentStr := string(content)

	// 检查是否已经存在该 brand
	brandPattern := fmt.Sprintf("// #ifdef MP-%s", strings.ToUpper(brandCode))
	if strings.Contains(contentStr, brandPattern) {
		fmt.Printf("Brand %s already exists in _host.js\n", brandCode)
		return nil
	}

	// 在 getBrand_ 函数中添加新的 brand
	// 找到 getBrand_ 函数的位置
	funcStartIndex := strings.Index(contentStr, "function getBrand_()")
	if funcStartIndex == -1 {
		return fmt.Errorf("cannot find getBrand_ function in _host.js")
	}

	// 找到 getBrand_ 函数内部的最后一个 #endif
	funcContent := contentStr[funcStartIndex:]
	lastEndifIndex := strings.LastIndex(funcContent, "// #endif")
	if lastEndifIndex == -1 {
		return fmt.Errorf("cannot find last #endif in getBrand_ function")
	}
	lastEndifIndex += funcStartIndex

	// 插入新的 brand
	newBrandCode := fmt.Sprintf(`  // #ifdef MP-%s
  return '%s'
  // #endif
`, strings.ToUpper(brandCode), brandCode)

	// 在 getBrand_ 函数内部，最后一个 #endif 之前插入新代码
	newContent := contentStr[:lastEndifIndex] + newBrandCode + "\n" + contentStr[lastEndifIndex:]

	// 写回文件
	if err := os.WriteFile(hostFilePath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write _host.js: %v", err)
	}

	fmt.Printf("✅ Added brand %s to _host.js\n", brandCode)
	return nil
}

// updateIndexFile 更新 index.js 文件，添加新的配置导入
func updateIndexFile(brandCode string) error {
	fmt.Printf("🔄 Starting updateIndexFile for brand: %s\n", brandCode)

	indexFilePath := "C:/F_explorer/h5projects/jianruiH5/novel_h5config/funNovel/src/appConfig/index.js"

	content, err := os.ReadFile(indexFilePath)
	if err != nil {
		fmt.Printf("❌ Failed to read index.js: %v\n", err)
		return fmt.Errorf("failed to read index.js: %v", err)
	}

	contentStr := string(content)

	// 检查是否已经存在该 brand 的导入
	brandPattern := fmt.Sprintf("// #ifdef MP-%s", strings.ToUpper(brandCode))
	if strings.Contains(contentStr, brandPattern) {
		fmt.Printf("Brand %s already exists in index.js\n", brandCode)
		return nil
	}

	// 在文件开头添加新的导入（在其他宏的外面）
	newImportCode := fmt.Sprintf(`// #ifdef MP-%s
import baseConfig from './baseConfigs/%s.js'
import commonConfig from './commonConfigs/%s.js'
import payConfig from './payConfigs/%s.js'
import uiConfig from './uiConfigs/%s.js'
import localConfig from './localConfigs/base.js'
// #endif

`, strings.ToUpper(brandCode), brandCode, brandCode, brandCode, brandCode)

	// 在文件开头插入新代码
	newContent := newImportCode + contentStr

	// 写回文件
	if err := os.WriteFile(indexFilePath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write index.js: %v", err)
	}

	fmt.Printf("✅ Added brand %s imports to index.js\n", brandCode)
	return nil
}

// updateViteConfigFile 更新 vite.config.js 文件，添加新的 script_base
func updateViteConfigFile(brandCode, host string, scriptBase string) error {
	viteConfigPath := "C:/F_explorer/h5projects/jianruiH5/novel_h5config/funNovel/vite.config.js"

	content, err := os.ReadFile(viteConfigPath)
	if err != nil {
		return fmt.Errorf("failed to read vite.config.js: %v", err)
	}

	contentStr := string(content)

	// 检查是否已经存在该配置
	scriptKey := fmt.Sprintf("'%s-%s'", host, brandCode)
	if strings.Contains(contentStr, scriptKey) {
		fmt.Printf("Script %s-%s already exists in vite.config.js\n", host, brandCode)
		return nil
	}

	// 找到 basePathMap 对象的位置
	mapStartIndex := strings.Index(contentStr, "const basePathMap = {")
	if mapStartIndex == -1 {
		return fmt.Errorf("cannot find basePathMap in vite.config.js")
	}

	// 找到 basePathMap 对象的结束位置
	mapEndIndex := strings.Index(contentStr[mapStartIndex:], "}")
	if mapEndIndex == -1 {
		return fmt.Errorf("cannot find basePathMap end in vite.config.js")
	}
	mapEndIndex += mapStartIndex

	// 插入新的配置，使用用户输入的 scriptBase
	newConfigEntry := fmt.Sprintf(`  '%s-%s': '%s',`, host, brandCode, scriptBase)

	// 在 basePathMap 对象内部插入新条目
	newContent := contentStr[:mapEndIndex] + "\n" + newConfigEntry + contentStr[mapEndIndex:]

	// 写回文件
	if err := os.WriteFile(viteConfigPath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write vite.config.js: %v", err)
	}

	fmt.Printf("✅ Added script %s-%s to vite.config.js\n", host, brandCode)
	return nil
}

// updatePackageJSONFile 更新 package.json 文件，添加新的编译配置
func updatePackageJSONFile(brandCode, host string) error {
	packageJSONPath := "C:/F_explorer/h5projects/jianruiH5/novel_h5config/funNovel/package.json"

	content, err := os.ReadFile(packageJSONPath)
	if err != nil {
		return fmt.Errorf("failed to read package.json: %v", err)
	}

	// 解析 JSON
	var packageData map[string]interface{}
	if err := json.Unmarshal(content, &packageData); err != nil {
		return fmt.Errorf("failed to parse package.json: %v", err)
	}

	// 检查 scripts 部分
	scripts, ok := packageData["scripts"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("scripts section not found in package.json")
	}

	// 添加 dev 脚本
	devScriptKey := fmt.Sprintf("dev:%s-%s", host, brandCode)
	if _, exists := scripts[devScriptKey]; !exists {
		scripts[devScriptKey] = fmt.Sprintf("uni -p %s-%s --minify", host, brandCode)
	}

	// 添加 build 脚本
	buildScriptKey := fmt.Sprintf("build:%s-%s", host, brandCode)
	if _, exists := scripts[buildScriptKey]; !exists {
		scripts[buildScriptKey] = fmt.Sprintf("cross-env UNI_UTS_PLATFORM=%s-%s npm run prebuild && uni build -p %s-%s --minify", host, brandCode, host, brandCode)
	}

	// 检查 uni-app 配置
	uniApp, ok := packageData["uni-app"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("uni-app section not found in package.json")
	}

	scriptsConfig, ok := uniApp["scripts"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("uni-app.scripts section not found in package.json")
	}

	// 添加编译配置
	buildConfigKey := fmt.Sprintf("%s-%s", host, brandCode)
	if _, exists := scriptsConfig[buildConfigKey]; !exists {
		// 根据 host 类型设置正确的 define 字段
		var defineConfig map[string]interface{}
		if host == "tth5" {
			defineConfig = map[string]interface{}{
				"MP-TTH5": true,
				fmt.Sprintf("MP-%s", strings.ToUpper(brandCode)): true,
			}
		} else if host == "ksh5" {
			defineConfig = map[string]interface{}{
				"MP-KSH5": true,
				fmt.Sprintf("MP-%s", strings.ToUpper(brandCode)): true,
			}
		} else if host == "h5" {
			defineConfig = map[string]interface{}{
				"MP-H5": true,
				fmt.Sprintf("MP-%s", strings.ToUpper(brandCode)): true,
			}
		}

		scriptsConfig[buildConfigKey] = map[string]interface{}{
			"define": defineConfig,
			"env": map[string]interface{}{
				"UNI_PLATFORM": "h5",
			},
			"title": fmt.Sprintf("h5%s", brandCode),
		}
	}

	// 写回文件
	newContent, err := json.MarshalIndent(packageData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal package.json: %v", err)
	}

	if err := os.WriteFile(packageJSONPath, newContent, 0644); err != nil {
		return fmt.Errorf("failed to write package.json: %v", err)
	}

	fmt.Printf("✅ Added build config %s-%s to package.json\n", host, brandCode)
	return nil
}

// updateIndexFileForConfig 更新 index.js 文件，添加新的配置导入
func updateIndexFileForConfig(brandCode string) error {
	fmt.Printf("🔄 Starting updateIndexFileForConfig for brand: %s\n", brandCode)

	indexFilePath := "C:/F_explorer/h5projects/jianruiH5/novel_h5config/funNovel/src/appConfig/index.js"

	content, err := os.ReadFile(indexFilePath)
	if err != nil {
		fmt.Printf("❌ Failed to read index.js: %v\n", err)
		return fmt.Errorf("failed to read index.js: %v", err)
	}

	contentStr := string(content)

	// 检查是否已经存在该 brand 的导入
	brandPattern := fmt.Sprintf("// #ifdef MP-%s", strings.ToUpper(brandCode))
	if strings.Contains(contentStr, brandPattern) {
		fmt.Printf("Brand %s already exists in index.js\n", brandCode)
		return nil
	}

	// 在文件开头添加新的导入（在其他宏的外面）
	newImportCode := fmt.Sprintf(`// #ifdef MP-%s
import baseConfig from './baseConfigs/%s.js'
import commonConfig from './commonConfigs/%s.js'
import payConfig from './payConfigs/%s.js'
import uiConfig from './uiConfigs/%s.js'
import localConfig from './localConfigs/base.js'
// #endif

`, strings.ToUpper(brandCode), brandCode, brandCode, brandCode, brandCode)

	// 在文件开头插入新代码
	newContent := newImportCode + contentStr

	// 写回文件
	if err := os.WriteFile(indexFilePath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write index.js: %v", err)
	}

	fmt.Printf("✅ Added brand %s imports to index.js\n", brandCode)
	return nil
}

// writeBaseConfigToFile 将BaseConfig写入本地文件，使用host作为key
func writeBaseConfigToFile(baseConfig models.BaseConfig, brandCode string, host string) error {
	// BaseConfig文件路径
	configDir := "C:/F_explorer/h5projects/jianruiH5/novel_h5config/funNovel/src/appConfig/baseConfigs"
	configFile := filepath.Join(configDir, brandCode+".js")

	// 确保目录存在
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %v", configDir, err)
	}

	// 检查文件是否存在
	_, err := os.Stat(configFile)
	fileExists := err == nil

	var configData map[string]interface{}

	if fileExists {
		// 读取现有文件
		content, err := os.ReadFile(configFile)
		if err != nil {
			return fmt.Errorf("failed to read existing config file: %v", err)
		}

		// 解析现有配置（移除export default前缀）
		contentStr := string(content)
		if strings.HasPrefix(contentStr, "export default ") {
			contentStr = strings.TrimPrefix(contentStr, "export default ")
		}

		// 将单引号替换为双引号，使其符合JSON格式
		contentStr = strings.ReplaceAll(contentStr, "'", `"`)

		if err := json.Unmarshal([]byte(contentStr), &configData); err != nil {
			return fmt.Errorf("failed to parse existing config file: %v", err)
		}
	} else {
		// 创建新的配置对象
		configData = make(map[string]interface{})
	}

	// 构建host配置
	hostConfig := map[string]interface{}{
		"app_name": baseConfig.AppName,
		"platform": baseConfig.Platform,
		"app_code": baseConfig.AppCode,
		"product":  baseConfig.Product,
		"customer": baseConfig.Customer,
		"appid":    baseConfig.AppID,
		"version":  baseConfig.Version,
		"cl":       baseConfig.CL,
		"uc":       baseConfig.UC,
	}

	// 使用host作为key
	configData[host] = hostConfig

	// 转换为JSON并写入文件
	configJSON, err := json.MarshalIndent(configData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config to JSON: %v", err)
	}

	// 添加export default前缀
	content := fmt.Sprintf("export default %s\n", string(configJSON))

	if err := os.WriteFile(configFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write config file %s: %v", configFile, err)
	}

	fmt.Printf("✅ BaseConfig written to: %s with host key: %s\n", configFile, host)
	return nil
}

// createPrebuildFiles 创建prebuild目录和文件
func createPrebuildFiles(brandCode string, appName string, host string) error {
	prebuildDir := "C:/F_explorer/h5projects/jianruiH5/novel_h5config/funNovel/prebuild/build"
	brandDir := filepath.Join(prebuildDir, brandCode)

	// 确保品牌目录存在
	if err := os.MkdirAll(brandDir, 0755); err != nil {
		return fmt.Errorf("failed to create brand directory: %v", err)
	}

	// 检查manifest.json是否存在
	manifestFile := filepath.Join(brandDir, "manifest.json")
	manifestExists := false
	if _, err := os.Stat(manifestFile); err == nil {
		manifestExists = true
	}

	// 如果manifest.json不存在，创建它
	if !manifestExists {
		manifestContent := fmt.Sprintf(`{
	"name": "%s",
	"appid": "",
	"description": "",
	"icon": "static/imgs/mine/head.png",
	"package": "com.example.demo",
	"minPlatformVersion": 1062, //最小平台支持 vivo/oppo >= 1062
	"versionName": "1.0.0",
	"versionCode": "100",
	"transformPx": false,
	"uniStatistics": {
		"enable": false
	},
	"vueVersion": "2",
	"h5" : {
	    "template" : "index.html",
	    "router" : {
	        "mode" : "history"
	    },
	    "title" : "%s"
	}
}`, brandCode, appName)

		if err := os.WriteFile(manifestFile, []byte(manifestContent), 0644); err != nil {
			return fmt.Errorf("failed to create manifest.json: %v", err)
		}
		fmt.Printf("✅ Created manifest.json for brand: %s\n", brandCode)
	}

	// 创建或更新pages-host.json文件
	pagesFile := filepath.Join(brandDir, fmt.Sprintf("pages-%s.json", host))

	// 检查文件是否存在
	pagesExists := false
	if _, err := os.Stat(pagesFile); err == nil {
		pagesExists = true
	}

	// 如果pages-host.json不存在，创建它
	if !pagesExists {
		pagesContent := fmt.Sprintf(`{
  "pages": [
    {
      "path": "pages/readerPage/readerPage",
      "style": {
        "navigationBarTitleText": "%s",
        "onReachBottomDistance": 50,
        "enablePullDownRefresh": true
      }
    },
    {
      "path": "pages/loginCallback/loginCallback",
      "style": {
        "navigationBarTitleText": "%s-登陆回调"
      }
    },
    {
      "path": "pages/userInfo/userInfo",
      "style": {
        "navigationBarTitleText": "%s-用户信息"
      }
    },
    {
      "path": "pages/testJump/testJump",
      "style": {
        "navigationBarTitleText": "%s-用户信息"
      }
    },
    {
      "path": "pages/webView/webView",
      "style": {
        "navigationBarTitleText": "%s"
      }
    }
  ],
  "globalStyle": {
    "navigationBarTitleText": "%s",
    "navigationStyle": "custom"
  }
}`, appName, appName, appName, appName, appName, appName)

		if err := os.WriteFile(pagesFile, []byte(pagesContent), 0644); err != nil {
			return fmt.Errorf("failed to create pages-%s.json: %v", host, err)
		}
		fmt.Printf("✅ Created pages-%s.json for brand: %s\n", host, brandCode)
	}

	return nil
}

// createStaticImageDirectory 创建static图片目录并拷贝内容
func createStaticImageDirectory(brandCode string) error {
	staticDir := "C:/F_explorer/h5projects/jianruiH5/novel_h5config/funNovel/src/static"
	sourceDir := filepath.Join(staticDir, "img-xingchen")
	targetDir := filepath.Join(staticDir, fmt.Sprintf("img-%s", brandCode))

	// 检查源目录是否存在
	if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
		return fmt.Errorf("source directory %s does not exist", sourceDir)
	}

	// 检查目标目录是否已存在
	if _, err := os.Stat(targetDir); err == nil {
		fmt.Printf("✅ Target directory already exists: %s\n", targetDir)
		return nil
	}

	// 创建目标目录
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %v", err)
	}

	// 拷贝目录内容
	if err := copyDirectory(sourceDir, targetDir); err != nil {
		return fmt.Errorf("failed to copy directory content: %v", err)
	}

	fmt.Printf("✅ Created static image directory: %s\n", targetDir)
	return nil
}

// copyDirectory 递归拷贝目录内容
func copyDirectory(src, dst string) error {
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
			if err := copyDirectory(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			// 拷贝文件
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// copyFile 拷贝单个文件
func copyFile(src, dst string) error {
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

// GetWebsiteConfig 获取网站配置
func GetWebsiteConfig(c *gin.Context) {
	clientID := c.Param("clientId")
	if clientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Client ID is required"})
		return
	}

	// 解析clientID为整数
	clientIDInt, err := strconv.Atoi(clientID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid client ID"})
		return
	}

	// 查询Client信息
	var client models.Client
	if err := database.DB.Preload("Brand").First(&client, clientIDInt).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Client not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch client"})
		}
		return
	}

	// 查询BaseConfig
	var baseConfig models.BaseConfig
	if err := database.DB.Where("client_id = ?", clientIDInt).First(&baseConfig).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch base config"})
		return
	}

	// 查询CommonConfig
	var commonConfig models.CommonConfig
	if err := database.DB.Where("client_id = ?", clientIDInt).First(&commonConfig).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch common config"})
		return
	}

	// 查询PayConfig
	var payConfig models.PayConfig
	if err := database.DB.Where("client_id = ?", clientIDInt).First(&payConfig).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch pay config"})
		return
	}

	// 查询UIConfig
	var uiConfig models.UIConfig
	if err := database.DB.Where("client_id = ?", clientIDInt).First(&uiConfig).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch UI config"})
		return
	}

	// 构建响应数据
	response := gin.H{
		"client": gin.H{
			"id":         client.ID,
			"host":       client.Host,
			"created_at": client.CreatedAt,
			"updated_at": client.UpdatedAt,
			"brand": gin.H{
				"id":   client.Brand.ID,
				"code": client.Brand.Code,
			},
		},
		"base_config":   baseConfig,
		"common_config": commonConfig,
		"pay_config":    payConfig,
		"ui_config":     uiConfig,
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Website config retrieved successfully",
		"data":    response,
	})
}
