package handlers

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"brand-config-api/database"
	"brand-config-api/models"

	"github.com/gin-gonic/gin"
)

// GetBrands 获取所有品牌
func GetBrands(c *gin.Context) {
	var brands []models.Brand

	result := database.DB.Find(&brands)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  brands,
		"total": len(brands),
	})
}

// GetBrand 获取单个品牌
func GetBrand(c *gin.Context) {
	id := c.Param("id")
	brandID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid brand ID"})
		return
	}

	var brand models.Brand
	result := database.DB.Preload("Clients").First(&brand, brandID)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Brand not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": brand})
}

// CreateBrand 创建品牌
func CreateBrand(c *gin.Context) {
	var req struct {
		Code string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "code is required"})
		return
	}

	// 校验唯一性
	var count int64
	database.DB.Model(&models.Brand{}).Where("code = ?", req.Code).Count(&count)
	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "code already exists"})
		return
	}

	brand := models.Brand{Code: req.Code}
	result := database.DB.Create(&brand)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	// 更新 _host.js 文件，添加新的 brand
	if err := updateHostFileForBrand(req.Code); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update _host.js: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": brand})
}

// updateHostFileForBrand 更新 _host.js 文件，添加新的 brand
func updateHostFileForBrand(brandCode string) error {
	fmt.Printf("🔄 Starting updateHostFileForBrand for brand: %s\n", brandCode)

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

// UpdateBrand 更新品牌
func UpdateBrand(c *gin.Context) {
	id := c.Param("id")
	brandID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid brand ID"})
		return
	}

	var brand models.Brand
	if err := database.DB.First(&brand, brandID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Brand not found"})
		return
	}

	var req struct {
		Code string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "code is required"})
		return
	}

	// 校验唯一性（排除自己）
	var count int64
	database.DB.Model(&models.Brand{}).Where("code = ? AND id != ?", req.Code, brandID).Count(&count)
	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "code already exists"})
		return
	}

	brand.Code = req.Code
	result := database.DB.Save(&brand)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": brand})
}

// DeleteBrand 删除品牌
func DeleteBrand(c *gin.Context) {
	id := c.Param("id")
	brandID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid brand ID"})
		return
	}

	result := database.DB.Delete(&models.Brand{}, brandID)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Brand not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Brand deleted successfully"})
}
