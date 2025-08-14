package services

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"brand-config-api/config"
	"brand-config-api/utils"
	"brand-config-api/utils/rollback"
)

// FileService 文件操作服务
type FileService struct {
	config    *config.Config
	fileUtils *utils.FileUtils // 保留对FileUtils的引用
	jsonUtils *utils.JSONUtils
}

// NewFileService 创建文件操作服务实例
func NewFileService() *FileService {
	return &FileService{
		config:    config.Load(),
		fileUtils: utils.NewFileUtils(), // 初始化fileUtils
		jsonUtils: utils.NewJSONUtils(), // 初始化jsonUtils
	}
}

// UpdateProjectConfigs 更新项目配置文件
func (s *FileService) UpdateProjectConfigs(brandCode, host string, scriptBase string, appName string, fileManager *rollback.FileRollback) error {
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
func (s *FileService) CreatePrebuildFiles(brandCode string, appName string, host string, fileManager *rollback.FileRollback) error {
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
func (s *FileService) CreateStaticImageDirectory(brandCode string, fileManager *rollback.FileRollback) error {
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
	if err := s.fileUtils.CopyDirectory(sourceDir, targetDir); err != nil {
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

	// 插入新的配置（确保换行）
	newConfigEntry := fmt.Sprintf(`  '%s-%s': '%s',`, host, brandCode, scriptBase)

	// 在basePathMap对象内部插入新条目，确保在反括号前换行
	newContent := contentStr[:mapEndIndex] + newConfigEntry + "\n" + contentStr[mapEndIndex:]

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

	// 1. 添加scripts - 直接在 "scripts": { 后一行添加
	scriptsStartIndex := strings.Index(contentStr, `"scripts": {`)
	if scriptsStartIndex == -1 {
		return fmt.Errorf("cannot find scripts block in package.json")
	}

	// 找到 "scripts": { 行的结束位置
	scriptsLineEndIndex := strings.Index(contentStr[scriptsStartIndex:], "\n")
	if scriptsLineEndIndex == -1 {
		return fmt.Errorf("cannot find end of scripts line")
	}
	insertPosition := scriptsStartIndex + scriptsLineEndIndex + 1

	// 添加新脚本（带逗号）
	newScripts := fmt.Sprintf(`    "dev:%s": "uni -p %s --minify",
    "build:%s": "cross-env UNI_UTS_PLATFORM=%s npm run prebuild && uni build -p %s --minify",
`, platformKey, platformKey, platformKey, platformKey, platformKey)

	// 插入新脚本
	contentStr = contentStr[:insertPosition] + newScripts + contentStr[insertPosition:]

	// 2. 添加uni-app.scripts - 直接在 "scripts": { 后一行添加
	uniAppScriptsStartIndex := strings.Index(contentStr, `"uni-app": {`)
	if uniAppScriptsStartIndex == -1 {
		return fmt.Errorf("cannot find uni-app block in package.json")
	}

	// 在uni-app块中查找 "scripts": {
	uniAppScriptsStart := strings.Index(contentStr[uniAppScriptsStartIndex:], `"scripts": {`)
	if uniAppScriptsStart == -1 {
		return fmt.Errorf("cannot find uni-app.scripts block in package.json")
	}

	uniAppScriptsStart += uniAppScriptsStartIndex
	uniAppScriptsLineEndIndex := strings.Index(contentStr[uniAppScriptsStart:], "\n")
	if uniAppScriptsLineEndIndex == -1 {
		return fmt.Errorf("cannot find end of uni-app.scripts line")
	}
	uniAppInsertPosition := uniAppScriptsStart + uniAppScriptsLineEndIndex + 1

	// 添加新uni-app脚本（带逗号，正确的缩进）
	newUniAppScript := fmt.Sprintf(`      "%s": {
        "env": {
          "UNI_PLATFORM": "%s"
        },
        "define": {
          "MP-%s": true`, platformKey, s.fileUtils.GetUniPlatform(host), strings.ToUpper(brandCode))

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
      },
`, appName)

	// 插入新uni-app脚本
	contentStr = contentStr[:uniAppInsertPosition] + newUniAppScript + contentStr[uniAppInsertPosition:]

	// 写回文件
	if err := os.WriteFile(s.config.File.PackageFile, []byte(contentStr), 0644); err != nil {
		return fmt.Errorf("failed to write package.json: %v", err)
	}

	fmt.Printf("✅ Updated package.json for brand %s with host %s\n", brandCode, host)
	return nil
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

	// 准备要插入的新 brand 代码（不包含额外的空行）
	newBrandCode := fmt.Sprintf(`  // #ifdef MP-%s
  return '%s'
  // #endif`, strings.ToUpper(brandCode), brandCode)

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

	// 生成新的配置块（不包含空行）
	newBlock := []string{
		fmt.Sprintf("// #ifdef MP-%s", upperBrandCode),
		fmt.Sprintf("import baseConfig from './baseConfigs/%s.js'", brandCode),
		fmt.Sprintf("import commonConfig from './commonConfigs/%s.js'", brandCode),
		fmt.Sprintf("import payConfig from './payConfigs/%s.js'", brandCode),
		fmt.Sprintf("import uiConfig from './uiConfigs/%s.js'", brandCode),
		"import localConfig from './localConfigs/base.js'",
		"// #endif",
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

// RemoveProjectConfigs 删除项目配置文件中的相关配置
func (s *FileService) RemoveProjectConfigs(brandCode, host string, fileManager *rollback.FileRollback) error {
	log.Printf("🗑️ 开始删除项目配置文件: brand=%s, host=%s", brandCode, host)

	// 1. 删除vite.config.js中的配置
	if err := s.removeViteConfigEntry(brandCode, host, fileManager); err != nil {
		return fmt.Errorf("failed to remove vite config entry: %v", err)
	}

	// 2. 删除package.json中的配置
	if err := s.removePackageJSONEntries(brandCode, host, fileManager); err != nil {
		return fmt.Errorf("failed to remove package.json entries: %v", err)
	}

	// 3. 删除prebuild目录下的pages-host.json文件
	if err := s.removePrebuildPagesFile(brandCode, host, fileManager); err != nil {
		return fmt.Errorf("failed to remove prebuild pages file: %v", err)
	}

	log.Printf("✅ 项目配置文件删除完成: brand=%s, host=%s", brandCode, host)
	return nil
}

// removeViteConfigEntry 删除vite.config.js中的配置条目
func (s *FileService) removeViteConfigEntry(brandCode, host string, fileManager *rollback.FileRollback) error {
	content, err := os.ReadFile(s.config.File.ViteConfigFile)
	if err != nil {
		return fmt.Errorf("failed to read vite.config.js: %v", err)
	}

	contentStr := string(content)

	// 生成要删除的配置键
	scriptKey := fmt.Sprintf("'%s-%s'", host, brandCode)

	// 检查是否存在该配置
	if !strings.Contains(contentStr, scriptKey) {
		log.Printf("⚠️ vite.config.js中不存在配置: %s", scriptKey)
		return nil
	}

	// 备份文件
	if err := fileManager.Backup(s.config.File.ViteConfigFile, ""); err != nil {
		return fmt.Errorf("failed to backup vite.config.js: %v", err)
	}

	// 找到并删除配置行
	lines := strings.Split(contentStr, "\n")
	var newLines []string

	for _, line := range lines {
		if strings.Contains(line, scriptKey) {
			// 删除包含配置键的行
			log.Printf("🗑️ 删除vite.config.js配置行: %s", strings.TrimSpace(line))
			continue
		}

		newLines = append(newLines, line)
	}

	// 写回文件
	newContent := strings.Join(newLines, "\n")
	if err := os.WriteFile(s.config.File.ViteConfigFile, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write vite.config.js: %v", err)
	}

	log.Printf("✅ 删除vite.config.js配置成功: %s", scriptKey)
	return nil
}

// removePackageJSONEntries 删除package.json中的配置条目
func (s *FileService) removePackageJSONEntries(brandCode, host string, fileManager *rollback.FileRollback) error {
	content, err := os.ReadFile(s.config.File.PackageFile)
	if err != nil {
		return fmt.Errorf("failed to read package.json: %v", err)
	}

	contentStr := string(content)

	// 生成平台标识
	platformKey := fmt.Sprintf("%s-%s", host, brandCode)

	// 检查是否存在该配置
	if !strings.Contains(contentStr, fmt.Sprintf(`"dev:%s"`, platformKey)) {
		log.Printf("⚠️ package.json中不存在配置: %s", platformKey)
		return nil
	}

	// 备份文件
	if err := fileManager.Backup(s.config.File.PackageFile, ""); err != nil {
		return fmt.Errorf("failed to backup package.json: %v", err)
	}

	// 按行删除包含该平台标识的所有行
	lines := strings.Split(contentStr, "\n")
	var newLines []string
	skipMode := false
	braceCount := 0

	for _, line := range lines {

		// 检查是否进入删除模式
		if strings.Contains(line, fmt.Sprintf(`"%s": {`, platformKey)) {
			skipMode = true
			braceCount = 1
			log.Printf("🗑️ 开始删除uni-app.scripts配置块: %s", platformKey)
			continue
		}

		// 如果在删除模式中
		if skipMode {
			// 计算大括号数量
			for _, char := range line {
				if char == '{' {
					braceCount++
				} else if char == '}' {
					braceCount--
					// 当大括号数量归零时，退出删除模式
					if braceCount == 0 {
						skipMode = false
						log.Printf("🗑️ 完成删除uni-app.scripts配置块: %s", platformKey)
						continue
					}
				}
			}
			// 跳过当前行
			continue
		}

		// 检查是否包含该平台标识的行（scripts部分）
		if strings.Contains(line, platformKey) {
			log.Printf("🗑️ 删除package.json行: %s", strings.TrimSpace(line))
			continue
		}

		newLines = append(newLines, line)
	}

	// 写回文件
	newContent := strings.Join(newLines, "\n")
	if err := os.WriteFile(s.config.File.PackageFile, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write package.json: %v", err)
	}

	log.Printf("✅ 删除package.json配置成功: %s", platformKey)
	return nil
}

// removePrebuildPagesFile 删除prebuild目录下的pages-host.json文件
func (s *FileService) removePrebuildPagesFile(brandCode, host string, fileManager *rollback.FileRollback) error {
	pagesFile := filepath.Join(s.config.File.PrebuildDir, brandCode, fmt.Sprintf("pages-%s.json", host))

	// 检查文件是否存在
	if _, err := os.Stat(pagesFile); os.IsNotExist(err) {
		log.Printf("⚠️ prebuild pages文件不存在: %s", pagesFile)
		return nil
	}

	// 备份文件
	if err := fileManager.Backup(pagesFile, ""); err != nil {
		return fmt.Errorf("failed to backup pages file: %v", err)
	}

	// 删除文件
	if err := os.Remove(pagesFile); err != nil {
		return fmt.Errorf("failed to delete pages file: %v", err)
	}

	log.Printf("✅ 删除prebuild pages文件成功: %s", pagesFile)
	return nil
}

// findBrandConfigEnd 找到品牌配置块的结束位置
func findBrandConfigEnd(content string, startIndex int) int {
	braceCount := 0
	for i := startIndex; i < len(content); i++ {
		if content[i] == '{' {
			braceCount++
		} else if content[i] == '}' {
			braceCount--
			if braceCount == 0 {
				return i + 1
			}
		}
	}
	return -1
}

// removeNovelConfigBrandBlock 删除novelconfig.js中该品牌的整个配置块
func (s *FileService) removeNovelConfigBrandBlock(brandCode string, fileManager *rollback.FileRollback) error {
	novelConfigFile := filepath.Join(s.config.File.LocalConfigsDir, "novelConfig.js")

	// 检查文件是否存在
	if _, err := os.Stat(novelConfigFile); os.IsNotExist(err) {
		log.Printf("⚠️ novelconfig.js文件不存在: %s", novelConfigFile)
		return nil
	}

	content, err := os.ReadFile(novelConfigFile)
	if err != nil {
		return fmt.Errorf("failed to read novelconfig.js: %v", err)
	}

	contentStr := string(content)

	// 检查是否存在该品牌的配置
	brandPattern := fmt.Sprintf(`"%s": {`, brandCode)
	if !strings.Contains(contentStr, brandPattern) {
		log.Printf("⚠️ novelconfig.js中不存在品牌配置: %s", brandCode)
		return nil
	}

	// 备份文件
	if err := fileManager.Backup(novelConfigFile, ""); err != nil {
		return fmt.Errorf("failed to backup novelconfig.js: %v", err)
	}

	// 删除该品牌的整个配置块
	brandStart := strings.Index(contentStr, brandPattern)
	if brandStart == -1 {
		return nil
	}

	// 找到品牌配置块的结束位置
	brandEnd := findBrandConfigEnd(contentStr, brandStart)
	if brandEnd == -1 {
		return nil
	}

	// 删除品牌配置块
	newContent := contentStr[:brandStart] + contentStr[brandEnd:]

	// 写回文件
	if err := os.WriteFile(novelConfigFile, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write novelconfig.js: %v", err)
	}

	log.Printf("✅ 删除novelconfig.js品牌配置块成功: %s", brandCode)
	return nil
}

// RemoveBrandFiles 删除品牌相关的所有文件
func (s *FileService) RemoveBrandFiles(brandCode, host string, fileManager *rollback.FileRollback) error {
	log.Printf("🗑️ 开始删除品牌相关文件: %s, host: %s", brandCode, host)

	// 1. 删除prebuild目录下对应的host子目录
	if err := s.removePrebuildHostDir(brandCode, host, fileManager); err != nil {
		return fmt.Errorf("failed to remove prebuild host directory: %v", err)
	}

	// 2. 删除各个配置目录下的 brandCode.js 文件
	if err := s.removeConfigFiles(brandCode, fileManager); err != nil {
		return fmt.Errorf("failed to remove config files: %v", err)
	}

	// 3. 删除prebuild目录下的 brandCode 目录
	if err := s.removePrebuildBrandDir(brandCode, fileManager); err != nil {
		return fmt.Errorf("failed to remove prebuild brand directory: %v", err)
	}

	// 4. 删除static图片目录（如果为空）
	if err := s.removeStaticImageDir(brandCode, fileManager); err != nil {
		log.Printf("⚠️ 删除static图片目录失败（可能不为空）: %v", err)
	}

	// 5. 删除index.js中该品牌的配置
	if err := s.removeIndexFileBrandConfig(brandCode, fileManager); err != nil {
		return fmt.Errorf("failed to remove index.js brand config: %v", err)
	}

	// 6. 删除_host.js中该品牌的配置
	if err := s.removeHostFileBrandConfig(brandCode, fileManager); err != nil {
		return fmt.Errorf("failed to remove _host.js brand config: %v", err)
	}

	// 7. 删除novelconfig.js中该品牌的整个配置块
	if err := s.removeNovelConfigBrandBlock(brandCode, fileManager); err != nil {
		return fmt.Errorf("failed to remove novelconfig.js brand block: %v", err)
	}

	log.Printf("✅ 品牌相关文件删除完成: %s", brandCode)
	return nil
}

// removeConfigFiles 删除各个配置目录下的 brandCode.js 文件
func (s *FileService) removeConfigFiles(brandCode string, fileManager *rollback.FileRollback) error {
	configFiles := []string{
		filepath.Join(s.config.File.BaseConfigsDir, brandCode+".js"),
		filepath.Join(s.config.File.CommonConfigsDir, brandCode+".js"),
		filepath.Join(s.config.File.PayConfigsDir, brandCode+".js"),
		filepath.Join(s.config.File.UIConfigsDir, brandCode+".js"),
	}

	for _, configFile := range configFiles {
		if err := s.removeFileIfExists(configFile, fileManager); err != nil {
			return fmt.Errorf("failed to remove config file %s: %v", configFile, err)
		}
	}

	return nil
}

// removePrebuildBrandDir 删除prebuild目录下的 brandCode 目录
func (s *FileService) removePrebuildBrandDir(brandCode string, fileManager *rollback.FileRollback) error {
	brandDir := filepath.Join(s.config.File.PrebuildDir, brandCode)

	if _, err := os.Stat(brandDir); os.IsNotExist(err) {
		log.Printf("⚠️ prebuild品牌目录不存在: %s", brandDir)
		return nil
	}

	// 备份目录
	if err := fileManager.Backup(brandDir, ""); err != nil {
		return fmt.Errorf("failed to backup brand directory: %v", err)
	}

	// 删除目录
	if err := os.RemoveAll(brandDir); err != nil {
		return fmt.Errorf("failed to delete brand directory: %v", err)
	}

	log.Printf("✅ 删除prebuild品牌目录成功: %s", brandDir)
	return nil
}

// removePrebuildHostDir 删除prebuild目录下对应的host子目录
func (s *FileService) removePrebuildHostDir(brandCode, host string, fileManager *rollback.FileRollback) error {
	hostDir := filepath.Join(s.config.File.PrebuildDir, brandCode, host)

	if _, err := os.Stat(hostDir); os.IsNotExist(err) {
		log.Printf("⚠️ prebuild host目录不存在: %s", hostDir)
		return nil
	}

	// 备份目录
	if err := fileManager.Backup(hostDir, ""); err != nil {
		return fmt.Errorf("failed to backup host directory: %v", err)
	}

	// 删除目录
	if err := os.RemoveAll(hostDir); err != nil {
		return fmt.Errorf("failed to delete host directory: %v", err)
	}

	log.Printf("✅ 删除prebuild host目录成功: %s", hostDir)
	return nil
}

// removeStaticImageDir 删除static图片目录
func (s *FileService) removeStaticImageDir(brandCode string, fileManager *rollback.FileRollback) error {
	staticImageDir := filepath.Join(s.config.File.StaticDir, "img-"+brandCode)

	if _, err := os.Stat(staticImageDir); os.IsNotExist(err) {
		log.Printf("⚠️ static图片目录不存在: %s", staticImageDir)
		return nil
	}

	// 备份目录
	if err := fileManager.Backup(staticImageDir, ""); err != nil {
		return fmt.Errorf("failed to backup static image directory: %v", err)
	}

	// 删除目录及其所有内容
	if err := os.RemoveAll(staticImageDir); err != nil {
		return fmt.Errorf("failed to delete static image directory: %v", err)
	}

	log.Printf("✅ 删除static图片目录成功: %s", staticImageDir)
	return nil
}

// removeIndexFileBrandConfig 删除index.js中该品牌的配置
func (s *FileService) removeIndexFileBrandConfig(brandCode string, fileManager *rollback.FileRollback) error {
	content, err := os.ReadFile(s.config.File.IndexFile)
	if err != nil {
		return fmt.Errorf("failed to read index.js: %v", err)
	}

	contentStr := string(content)
	upperBrandCode := strings.ToUpper(brandCode)

	// 检查是否存在该品牌的配置
	if !strings.Contains(contentStr, "#ifdef MP-"+upperBrandCode) {
		log.Printf("⚠️ index.js中不存在品牌配置: %s", brandCode)
		return nil
	}

	// 备份文件
	if err := fileManager.Backup(s.config.File.IndexFile, ""); err != nil {
		return fmt.Errorf("failed to backup index.js: %v", err)
	}

	// 删除该品牌的配置块
	lines := strings.Split(contentStr, "\n")
	var newLines []string
	inBrandBlock := false
	skipLines := 0

	for _, line := range lines {
		if strings.Contains(line, "#ifdef MP-"+upperBrandCode) {
			inBrandBlock = true
			skipLines = 0
			continue
		}

		if inBrandBlock {
			if strings.Contains(line, "#endif") {
				inBrandBlock = false
				continue
			}
			skipLines++
			continue
		}

		newLines = append(newLines, line)
	}

	// 写回文件
	newContent := strings.Join(newLines, "\n")
	if err := os.WriteFile(s.config.File.IndexFile, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write index.js: %v", err)
	}

	log.Printf("✅ 删除index.js品牌配置成功: %s", brandCode)
	return nil
}

// removeHostFileBrandConfig 删除_host.js中该品牌的配置
func (s *FileService) removeHostFileBrandConfig(brandCode string, fileManager *rollback.FileRollback) error {
	content, err := os.ReadFile(s.config.File.HostFile)
	if err != nil {
		return fmt.Errorf("failed to read _host.js: %v", err)
	}

	contentStr := string(content)
	upperBrandCode := strings.ToUpper(brandCode)

	// 检查是否存在该品牌的配置
	brandPattern := fmt.Sprintf("// #ifdef MP-%s", upperBrandCode)
	if !strings.Contains(contentStr, brandPattern) {
		log.Printf("⚠️ _host.js中不存在品牌配置: %s", brandCode)
		return nil
	}

	// 备份文件
	if err := fileManager.Backup(s.config.File.HostFile, ""); err != nil {
		return fmt.Errorf("failed to backup _host.js: %v", err)
	}

	// 删除该品牌的配置块
	lines := strings.Split(contentStr, "\n")
	var newLines []string
	inBrandBlock := false

	for _, line := range lines {
		if strings.Contains(line, brandPattern) {
			inBrandBlock = true
			continue
		}

		if inBrandBlock {
			if strings.Contains(line, "// #endif") {
				inBrandBlock = false
				continue
			}
			continue
		}

		newLines = append(newLines, line)
	}

	// 写回文件
	newContent := strings.Join(newLines, "\n")
	if err := os.WriteFile(s.config.File.HostFile, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write _host.js: %v", err)
	}

	log.Printf("✅ 删除_host.js品牌配置成功: %s", brandCode)
	return nil
}

// removeFileIfExists 删除文件（如果存在）
func (s *FileService) removeFileIfExists(filePath string, fileManager *rollback.FileRollback) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Printf("⚠️ 文件不存在，跳过删除: %s", filePath)
		return nil
	}

	// 检查文件是否已被备份（用于回滚）
	if fileManager.HasBackup(filePath) {
		log.Printf("✅ 文件已被备份，可以直接删除: %s", filePath)
	} else {
		log.Printf("⚠️ 文件未被备份，删除后无法回滚: %s", filePath)
	}

	// 删除文件
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to delete file: %v", err)
	}

	log.Printf("🗑️ 文件删除成功: %s", filePath)
	return nil
}
