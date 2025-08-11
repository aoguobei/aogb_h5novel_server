package services

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"brand-config-api/config"
	"brand-config-api/utils"
)

// FileService 文件操作服务
type FileService struct {
	config *config.Config
}

// NewFileService 创建文件操作服务实例
func NewFileService() *FileService {
	return &FileService{
		config: config.Load(),
	}
}

// BackupProjectFiles 备份项目文件
func (s *FileService) BackupProjectFiles(brandCode string, rollbackManager *utils.RollbackManager) error {
	// 备份vite.config.js
	if err := rollbackManager.BackupProjectFile(s.config.File.ViteConfigFile); err != nil {
		return err
	}

	// 备份package.json
	if err := rollbackManager.BackupProjectFile(s.config.File.PackageFile); err != nil {
		return err
	}

	// 备份配置文件
	if err := rollbackManager.BackupConfigFile("base", brandCode); err != nil {
		return err
	}
	if err := rollbackManager.BackupConfigFile("common", brandCode); err != nil {
		return err
	}
	if err := rollbackManager.BackupConfigFile("pay", brandCode); err != nil {
		return err
	}
	if err := rollbackManager.BackupConfigFile("ui", brandCode); err != nil {
		return err
	}

	return nil
}

// UpdateProjectConfigs 更新项目配置文件
func (s *FileService) UpdateProjectConfigs(brandCode, host string, scriptBase string, appName string, rollbackManager *utils.RollbackManager) error {
	// 注意：_host.js 只在新建品牌时更新，不在新建网站时更新
	// 所以这里不调用 updateHostFileForBrand

	// 更新index.js文件
	if err := s.updateIndexFileForBrand(brandCode); err != nil {
		return err
	}

	// 更新vite.config.js
	if err := s.updateViteConfigFile(brandCode, host, scriptBase); err != nil {
		return err
	}

	// 更新package.json
	if err := s.updatePackageJSONFile(brandCode, host, appName); err != nil {
		return err
	}

	return nil
}

// CreatePrebuildFiles 创建prebuild文件
func (s *FileService) CreatePrebuildFiles(brandCode string, appName string, host string, rollbackManager *utils.RollbackManager) error {
	brandDir := s.config.GetPrebuildPath(brandCode)

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
	"minPlatformVersion": 1062,
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
	}

	return nil
}

// CreateStaticImageDirectory 创建static图片目录
func (s *FileService) CreateStaticImageDirectory(brandCode string, rollbackManager *utils.RollbackManager) error {
	sourceDir := filepath.Join(s.config.File.StaticDir, "img-jinse")
	targetDir := s.config.GetStaticPath(brandCode)

	// 检查源目录是否存在
	if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
		return fmt.Errorf("source directory %s does not exist", sourceDir)
	}

	// 检查目标目录是否已存在
	if _, err := os.Stat(targetDir); err == nil {
		return nil
	}

	// 创建目标目录
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %v", err)
	}

	// 拷贝目录内容
	if err := s.copyDirectory(sourceDir, targetDir); err != nil {
		return fmt.Errorf("failed to copy directory content: %v", err)
	}

	return nil
}

// updateViteConfigFile 更新vite.config.js文件
func (s *FileService) updateViteConfigFile(brandCode, host string, scriptBase string) error {
	content, err := os.ReadFile(s.config.File.ViteConfigFile)
	if err != nil {
		return fmt.Errorf("failed to read vite.config.js: %v", err)
	}

	contentStr := string(content)

	// 检查是否已经存在该配置
	scriptKey := fmt.Sprintf("'%s-%s'", host, brandCode)
	if strings.Contains(contentStr, scriptKey) {
		return nil
	}

	// 找到basePathMap对象的位置
	mapStartIndex := strings.Index(contentStr, "const basePathMap = {")
	if mapStartIndex == -1 {
		return fmt.Errorf("cannot find basePathMap in vite.config.js")
	}

	// 找到basePathMap对象的结束位置
	mapEndIndex := strings.Index(contentStr[mapStartIndex:], "}")
	if mapEndIndex == -1 {
		return fmt.Errorf("cannot find basePathMap end in vite.config.js")
	}
	mapEndIndex += mapStartIndex

	// 插入新的配置
	newConfigEntry := fmt.Sprintf(`  '%s-%s': '%s',`, host, brandCode, scriptBase)

	// 在basePathMap对象内部插入新条目
	newContent := contentStr[:mapEndIndex] + "\n" + newConfigEntry + contentStr[mapEndIndex:]

	// 写回文件
	if err := os.WriteFile(s.config.File.ViteConfigFile, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write vite.config.js: %v", err)
	}

	return nil
}

// updatePackageJSONFile 更新package.json文件
func (s *FileService) updatePackageJSONFile(brandCode, host string, appName string) error {
	content, err := os.ReadFile(s.config.File.PackageFile)
	if err != nil {
		return fmt.Errorf("failed to read package.json: %v", err)
	}

	contentStr := string(content)

	// 生成平台标识
	platformKey := fmt.Sprintf("%s-%s", host, brandCode)

	// 检查是否已经存在该配置
	if strings.Contains(contentStr, fmt.Sprintf(`"dev:%s"`, platformKey)) {
		fmt.Printf("⚠️  Platform %s already exists in package.json, skipping...\n", platformKey)
		return nil
	}

	// 1. 添加scripts - 在scripts块的末尾添加
	scriptsEndIndex := s.findScriptsEndIndex(contentStr)
	if scriptsEndIndex > 0 {
		fmt.Printf("🔍 Found scripts block end position: %d\n", scriptsEndIndex)

		// 找到scripts块中最后一个脚本的结束位置
		lastScriptEndIndex := s.findLastScriptEndIndex(contentStr, scriptsEndIndex)
		if lastScriptEndIndex > 0 {
			fmt.Printf("🔍 Found last script end position: %d\n", lastScriptEndIndex)

			// 在最后一个脚本后面添加逗号和新脚本
			newScripts := fmt.Sprintf(`,
    "dev:%s": "uni -p %s --minify",
    "build:%s": "cross-env UNI_UTS_PLATFORM=%s npm run prebuild && uni build -p %s --minify"`,
				platformKey, platformKey, platformKey, platformKey, platformKey)

			// 在最后一个脚本后面插入新内容
			contentStr = contentStr[:lastScriptEndIndex] + newScripts + contentStr[lastScriptEndIndex:]
			fmt.Println("✅ Added scripts configuration")
		} else {
			fmt.Println("❌ Could not find last script position")
			return fmt.Errorf("failed to find last script position")
		}
	} else {
		fmt.Println("❌ Could not find scripts block")
		return fmt.Errorf("failed to find scripts block")
	}

	// 2. 添加uni-app.scripts - 在uni-app.scripts块的末尾添加
	uniAppScriptsEndIndex := s.findUniAppScriptsEndIndex(contentStr)
	if uniAppScriptsEndIndex > 0 {
		fmt.Printf("🔍 Found uni-app.scripts block end position: %d\n", uniAppScriptsEndIndex)

		// 找到uni-app.scripts块中最后一个配置的结束位置
		lastUniAppScriptEndIndex := s.findLastUniAppScriptEndIndex(contentStr, uniAppScriptsEndIndex)
		if lastUniAppScriptEndIndex > 0 {
			fmt.Printf("🔍 Found last uni-app script end position: %d\n", lastUniAppScriptEndIndex)

			// 在最后一个配置后面添加逗号和新配置
			newUniAppScript := fmt.Sprintf(`,
    "%s": {
      "env": {
        "UNI_PLATFORM": "%s"
      },
      "define": {
        "MP-%s": true`, platformKey, s.getUniPlatform(host), strings.ToUpper(brandCode))

			// 根据host类型设置对应的平台宏
			switch host {
			case "tth5":
				newUniAppScript += `,
        "MP-TTH5": true`
			case "ksh5":
				newUniAppScript += `,
        "MP-KSH5": true`
			case "h5":
				newUniAppScript += `,
        "MP-H5": true`
			}

			newUniAppScript += fmt.Sprintf(`
      },
      "title": "h5%s"
    }`, appName)

			// 在最后一个配置后面插入新内容
			contentStr = contentStr[:lastUniAppScriptEndIndex] + newUniAppScript + contentStr[lastUniAppScriptEndIndex:]
			fmt.Println("✅ Added uni-app.scripts configuration")
		} else {
			fmt.Println("❌ Could not find last uni-app script position")
			return fmt.Errorf("failed to find last uni-app script position")
		}
	} else {
		fmt.Println("❌ Could not find uni-app.scripts block")
		return fmt.Errorf("failed to find uni-app.scripts block")
	}

	// 写回文件，保持原有格式
	if err := os.WriteFile(s.config.File.PackageFile, []byte(contentStr), 0644); err != nil {
		return fmt.Errorf("failed to write package.json: %v", err)
	}

	fmt.Printf("✅ Updated package.json for brand %s with host %s\n", brandCode, host)
	return nil
}

// getUniPlatform 根据host获取对应的uni平台标识
func (s *FileService) getUniPlatform(host string) string {
	switch host {
	case "tt":
		return "mp-toutiao"
	case "ks":
		return "mp-kuaishou"
	case "wx":
		return "mp-weixin"
	case "bd":
		return "mp-baidu"
	default:
		return "h5"
	}
}

// copyDirectory 递归拷贝目录内容
func (s *FileService) copyDirectory(src, dst string) error {
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
			if err := s.copyDirectory(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			// 拷贝文件
			if err := s.copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// copyFile 拷贝单个文件
func (s *FileService) copyFile(src, dst string) error {
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

// UpdateHostFileForBrand 更新 _host.js 文件，添加新的 brand
func (s *FileService) UpdateHostFileForBrand(brandCode string) error {
	fmt.Printf("🔄 Starting updateHostFileForBrand for brand: %s\n", brandCode)

	hostFilePath := s.config.File.HostFile

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
	funcStartStr := "function getBrand_()"
	funcStartIndex := strings.Index(contentStr, funcStartStr)
	if funcStartIndex == -1 {
		return fmt.Errorf("cannot find getBrand_ function in _host.js")
	}

	// 计算函数声明后的插入位置（函数声明行的下一行）
	// 找到函数声明行的结尾
	funcLineEndIndex := funcStartIndex + len(funcStartStr)
	// 找到下一个换行符
	nextNewlineIndex := strings.Index(contentStr[funcLineEndIndex:], "\n")
	if nextNewlineIndex == -1 {
		return fmt.Errorf("cannot find end of getBrand_ function declaration line")
	}
	// 插入位置是函数声明行的下一行开头
	insertPosition := funcLineEndIndex + nextNewlineIndex + 1

	// 准备要插入的新 brand 代码
	newBrandCode := fmt.Sprintf(`  // #ifdef MP-%s
  return '%s'
  // #endif
`, strings.ToUpper(brandCode), brandCode)

	// 在 getBrand_ 函数声明的下一行插入新代码
	newContent := contentStr[:insertPosition] + newBrandCode + "\n" + contentStr[insertPosition:]

	// 写回文件
	if err := os.WriteFile(hostFilePath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write _host.js: %v", err)
	}

	fmt.Printf("✅ Added brand %s to _host.js\n", brandCode)
	return nil
}

// updateIndexFileForBrand 更新 index.js 文件，添加新的 brand 配置
func (s *FileService) updateIndexFileForBrand(brandCode string) error {
	fmt.Printf("🔄 Starting updateIndexFileForBrand for brand: %s\n", brandCode)

	// 使用配置文件中的IndexFile路径

	content, err := os.ReadFile(s.config.File.IndexFile)
	if err != nil {
		fmt.Printf("❌ Failed to read index.js: %v\n", err)
		return fmt.Errorf("failed to read index.js: %v", err)
	}

	contentStr := string(content)
	lines := strings.Split(contentStr, "\n")

	// 检查是否已存在该品牌的配置块
	upperBrandCode := strings.ToUpper(brandCode)
	existingBlock := false
	for _, line := range lines {
		if strings.Contains(line, "#ifdef MP-"+upperBrandCode) {
			existingBlock = true
			break
		}
	}

	if existingBlock {
		fmt.Printf("✅ Brand %s already exists in index.js\n", brandCode)
		return nil
	}

	// 生成新的配置块
	newBlock := []string{
		fmt.Sprintf("// #ifdef MP-%s", upperBrandCode),
		fmt.Sprintf("import baseConfig from './baseConfigs/%s.js'", brandCode),
		fmt.Sprintf("import commonConfig from './commonConfigs/%s.js'", brandCode),
		fmt.Sprintf("import payConfig from './payConfigs/%s.js'", brandCode),
		fmt.Sprintf("import uiConfig from './uiConfigs/%s.js'", brandCode),
		"import localConfig from './localConfigs/base.js'",
		"// #endif",
		"",
	}

	// 找到第一个 #ifdef 的位置，在其前面插入新块
	insertIndex := -1
	for i, line := range lines {
		if strings.Contains(line, "#ifdef MP-") {
			insertIndex = i
			break
		}
	}

	if insertIndex == -1 {
		// 如果没有找到 #ifdef，则在文件开头插入
		insertIndex = 0
	}

	// 插入新块
	lines = append(lines[:insertIndex], append(newBlock, lines[insertIndex:]...)...)

	// 写回文件
	newContent := strings.Join(lines, "\n")
	if err := os.WriteFile(s.config.File.IndexFile, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write index.js: %v", err)
	}

	fmt.Printf("✅ Updated index.js for brand %s\n", brandCode)
	return nil
}

// findScriptsEndIndex 找到scripts块的结束位置
func (s *FileService) findScriptsEndIndex(content string) int {
	// 查找 "scripts": { 的开始位置
	scriptsStart := strings.Index(content, `"scripts": {`)
	if scriptsStart == -1 {
		return -1
	}

	// 从scripts开始位置向后查找对应的结束大括号
	braceCount := 0

	for i := scriptsStart; i < len(content); i++ {
		if content[i] == '{' {
			braceCount++
		} else if content[i] == '}' {
			braceCount--
			if braceCount == 0 {
				return i
			}
		}
	}

	return -1
}

// findLastScriptEndIndex 找到scripts块中最后一个脚本的结束位置
func (s *FileService) findLastScriptEndIndex(content string, scriptsEndIndex int) int {
	// 从scripts块开始位置向后查找
	scriptsStart := strings.Index(content, `"scripts": {`)
	if scriptsStart == -1 {
		return -1
	}

	// 在scripts块中查找最后一个脚本的结束位置
	// 从scripts结束位置向前查找最后一个脚本的结束引号
	for i := scriptsEndIndex - 1; i > scriptsStart; i-- {
		// 查找最后一个脚本的结束引号
		if content[i] == '"' {
			// 向前查找这个脚本的开始引号
			for j := i - 1; j > scriptsStart; j-- {
				if content[j] == '"' {
					// 找到了脚本值的开始引号，现在需要找到这个值的结束引号
					// 从开始引号向后查找结束引号
					for k := j + 1; k < scriptsEndIndex; k++ {
						if content[k] == '"' {
							// 找到了值的结束引号，返回这个位置
							return k + 1
						}
					}
				}
			}
		}
	}

	return -1
}

// findUniAppScriptsEndIndex 找到uni-app.scripts块的结束位置
func (s *FileService) findUniAppScriptsEndIndex(content string) int {
	// 查找 "uni-app": { 的开始位置
	uniAppStart := strings.Index(content, `"uni-app": {`)
	if uniAppStart == -1 {
		return -1
	}

	// 从uni-app开始位置向后查找 "scripts": { 的开始位置
	scriptsStart := strings.Index(content[uniAppStart:], `"scripts": {`)
	if scriptsStart == -1 {
		return -1
	}

	scriptsStart += uniAppStart

	// 从scripts开始位置向后查找对应的结束大括号
	braceCount := 0

	for i := scriptsStart; i < len(content); i++ {
		if content[i] == '{' {
			braceCount++
		} else if content[i] == '}' {
			braceCount--
			if braceCount == 0 {
				return i
			}
		}
	}

	return -1
}

// findLastUniAppScriptEndIndex 找到uni-app.scripts块中最后一个配置的结束位置
func (s *FileService) findLastUniAppScriptEndIndex(content string, uniAppScriptsEndIndex int) int {
	// 从uni-app.scripts块开始位置向后查找
	uniAppStart := strings.Index(content, `"uni-app": {`)
	if uniAppStart == -1 {
		return -1
	}

	// 从uni-app开始位置向后查找 "scripts": { 的开始位置
	scriptsStart := strings.Index(content[uniAppStart:], `"scripts": {`)
	if scriptsStart == -1 {
		return -1
	}

	scriptsStart += uniAppStart

	// 在uni-app.scripts块中查找最后一个配置的结束位置
	// 从scripts结束位置向前查找最后一个配置的结束大括号
	for i := uniAppScriptsEndIndex - 1; i > scriptsStart; i-- {
		// 查找最后一个配置的结束大括号
		if content[i] == '}' {
			// 向前查找这个配置的开始大括号
			braceCount := 1
			for j := i - 1; j > scriptsStart; j-- {
				if content[j] == '}' {
					braceCount++
				} else if content[j] == '{' {
					braceCount--
					if braceCount == 0 {
						// 找到了配置的开始大括号，返回结束大括号后的位置
						return i + 1
					}
				}
			}
		}
	}

	return -1
}
