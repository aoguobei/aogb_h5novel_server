package services

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"brand-config-api/config"
	"brand-config-api/models"
	"brand-config-api/utils"
)

// ConfigGeneratorService 配置生成服务
type ConfigGeneratorService struct {
	config *config.Config
}

// NewConfigGeneratorService 创建配置生成服务实例
func NewConfigGeneratorService() *ConfigGeneratorService {
	return &ConfigGeneratorService{
		config: config.Load(),
	}
}

// GenerateConfigFiles 生成配置文件
func (s *ConfigGeneratorService) GenerateConfigFiles(brandCode, host string, baseConfig models.BaseConfig, extraBaseConfig *models.BaseConfig, commonConfig models.CommonConfig, payConfig models.PayConfig, uiConfig models.UIConfig, rollbackManager *utils.RollbackManager) error {
	// 生成主BaseConfig文件
	if err := s.writeBaseConfigToFile(baseConfig, brandCode, host, rollbackManager); err != nil {
		return err
	}

	// 生成额外的BaseConfig文件（如果存在）
	if extraBaseConfig != nil {
		var extraHost string
		if host == "tth5" {
			extraHost = "tt"
		} else if host == "ksh5" {
			extraHost = "ks"
		}
		if err := s.writeBaseConfigToFile(*extraBaseConfig, brandCode, extraHost, rollbackManager); err != nil {
			return err
		}
	}

	// 写入其他配置文件
	formattedCommonConfig := s.formatCommonConfig(commonConfig)
	if err := s.writeConfigToFile("common", formattedCommonConfig, brandCode, host, rollbackManager); err != nil {
		return err
	}

	formattedPayConfig := s.formatPayConfig(payConfig)
	if err := s.writeConfigToFile("pay", formattedPayConfig, brandCode, host, rollbackManager); err != nil {
		return err
	}

	formattedUIConfig := s.formatUIConfig(uiConfig)
	if err := s.writeConfigToFile("ui", formattedUIConfig, brandCode, host, rollbackManager); err != nil {
		return err
	}

	return nil
}

// writeBaseConfigToFile 将BaseConfig写入文件
func (s *ConfigGeneratorService) writeBaseConfigToFile(baseConfig models.BaseConfig, brandCode string, host string, rollbackManager *utils.RollbackManager) error {
	// 根据配置类型确定文件路径
	configDir := "C:/F_explorer/h5projects/jianruiH5/novel_h5config/funNovel/src/appConfig/baseConfigs"
	fileName := brandCode + ".js"
	configFile := filepath.Join(configDir, fileName)

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

		// 解析现有配置
		contentStr := string(content)
		if strings.HasPrefix(contentStr, "export default ") {
			contentStr = strings.TrimPrefix(contentStr, "export default ")
		}

		contentStr = strings.ReplaceAll(contentStr, "'", `"`)

		if err := json.Unmarshal([]byte(contentStr), &configData); err != nil {
			return fmt.Errorf("failed to parse existing config file: %v", err)
		}
	} else {
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

	configData[host] = hostConfig

	// 转换为JSON并写入文件
	configJSON, err := json.MarshalIndent(configData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config to JSON: %v", err)
	}

	content := fmt.Sprintf("export default %s\n", string(configJSON))

	if err := os.WriteFile(configFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write config file %s: %v", configFile, err)
	}

	fmt.Printf("✅ base config written to: %s with host key: %s\n", configFile, host)
	return nil
}

// writeConfigToFile 写入配置文件
func (s *ConfigGeneratorService) writeConfigToFile(configType string, formattedConfig map[string]interface{}, brandCode, host string, rollbackManager *utils.RollbackManager) error {
	// 根据配置类型确定文件路径
	var configDir string
	var fileName string

	switch configType {
	case "common":
		configDir = "C:/F_explorer/h5projects/jianruiH5/novel_h5config/funNovel/src/appConfig/commonConfigs"
		fileName = brandCode + ".js"
	case "pay":
		configDir = "C:/F_explorer/h5projects/jianruiH5/novel_h5config/funNovel/src/appConfig/payConfigs"
		fileName = brandCode + ".js"
	case "ui":
		configDir = "C:/F_explorer/h5projects/jianruiH5/novel_h5config/funNovel/src/appConfig/uiConfigs"
		fileName = brandCode + ".js"
	default:
		return fmt.Errorf("unknown config type: %s", configType)
	}

	// 确保目录存在
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %v", configDir, err)
	}

	// 构建文件路径
	configFile := filepath.Join(configDir, fileName)

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

	// 使用host作为key，添加配置
	configData[host] = formattedConfig

	// 转换为JSON并添加export default前缀
	configJSON, err := json.MarshalIndent(configData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config to JSON: %v", err)
	}

	content := fmt.Sprintf("export default %s\n", string(configJSON))

	if err := os.WriteFile(configFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write config file %s: %v", configFile, err)
	}

	fmt.Printf("✅ %s config written to: %s with host key: %s\n", configType, configFile, host)
	return nil
}

// formatCommonConfig 格式化通用配置
func (s *ConfigGeneratorService) formatCommonConfig(commonConfig models.CommonConfig) map[string]interface{} {
	return map[string]interface{}{
		"deliver": map[string]interface{}{
			"business_id": map[string]interface{}{
				"value":  commonConfig.DeliverBusinessID,
				"enable": commonConfig.DeliverBusinessIDEnable,
			},
			"switch_id": map[string]interface{}{
				"value":  commonConfig.DeliverSwitchID,
				"enable": commonConfig.DeliverSwitchIDEnable,
			},
		},
		"protocol": map[string]interface{}{
			"company":    commonConfig.ProtocolCompany,
			"about":      commonConfig.ProtocolAbout,
			"privacy":    commonConfig.ProtocolPrivacy,
			"vod":        commonConfig.ProtocolVod,
			"userCancel": commonConfig.ProtocolUserCancel,
		},
		"contact": commonConfig.ContactURL,
		"script": map[string]interface{}{
			"base": commonConfig.ScriptBase,
		},
	}
}

// formatPayConfig 格式化支付配置
func (s *ConfigGeneratorService) formatPayConfig(payConfig models.PayConfig) map[string]interface{} {
	return map[string]interface{}{
		"normal_pay": map[string]interface{}{
			"enable": payConfig.NormalPayEnable,
			"gateway_id": map[string]interface{}{
				"android": payConfig.NormalPayGatewayAndroid,
				"ios":     payConfig.NormalPayGatewayIOS,
			},
		},
		"renew_pay": map[string]interface{}{
			"enable": payConfig.RenewPayEnable,
			"gateway_id": map[string]interface{}{
				"android": payConfig.RenewPayGatewayAndroid,
				"ios":     payConfig.RenewPayGatewayIOS,
			},
		},
	}
}

// formatUIConfig 格式化UI配置
func (s *ConfigGeneratorService) formatUIConfig(uiConfig models.UIConfig) map[string]interface{} {
	result := map[string]interface{}{
		"bgStyle": map[string]interface{}{
			"main":   uiConfig.ThemeBgMain,
			"second": uiConfig.ThemeBgSecond,
		},
	}

	if uiConfig.ThemeTextMain != nil && *uiConfig.ThemeTextMain != "" {
		result["textColor"] = map[string]interface{}{
			"main": *uiConfig.ThemeTextMain,
		}
	}

	return result
}
