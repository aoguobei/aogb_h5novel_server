package handlers

import (
	"log"
	"strconv"

	"brand-config-api/services"
	"brand-config-api/utils"

	"github.com/gin-gonic/gin"
)

// BrandHandler 品牌控制器
type BrandHandler struct {
	brandService *services.BrandService
}

// NewBrandHandler 创建品牌控制器
func NewBrandHandler() *BrandHandler {
	return &BrandHandler{
		brandService: services.NewBrandService(),
	}
}

// GetBrands 获取所有品牌
func (h *BrandHandler) GetBrands(c *gin.Context) {
	brands, err := h.brandService.GetAllBrands()
	if err != nil {
		utils.InternalServerError(c, "获取品牌列表失败")
		return
	}

	utils.Success(c, gin.H{
		"data":  brands,
		"total": len(brands),
	}, "获取品牌列表成功")
}

// GetBrand 获取单个品牌
func (h *BrandHandler) GetBrand(c *gin.Context) {
	id := c.Param("id")
	brandID, err := strconv.Atoi(id)
	if err != nil {
		utils.BadRequest(c, "无效的品牌ID")
		return
	}

	brand, err := h.brandService.GetBrandByID(brandID)
	if err != nil {
		utils.NotFound(c, "品牌不存在")
		return
	}

	utils.Success(c, gin.H{"data": brand}, "获取品牌成功")
}

// CreateBrand 创建品牌
func (h *BrandHandler) CreateBrand(c *gin.Context) {
	var req struct {
		Code   string `json:"code" binding:"required"`
		TypeID uint   `json:"type_id" binding:"required"`
	}

	// 先绑定参数再打印
	if err := c.ShouldBindJSON(&req); err != nil {
		// 打印错误详情和原始请求体，便于调试
		body, _ := c.GetRawData()
		log.Printf("参数解析错误: %v, 原始请求体: %s", err, body)
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	brand, err := h.brandService.CreateBrand(req.Code, req.TypeID)
	if err != nil {
		utils.Conflict(c, err.Error())
		return
	}

	utils.Created(c, gin.H{"data": brand}, "品牌创建成功")
}

// UpdateBrand 更新品牌
func (h *BrandHandler) UpdateBrand(c *gin.Context) {
	id := c.Param("id")
	brandID, err := strconv.Atoi(id)
	if err != nil {
		utils.BadRequest(c, "无效的品牌ID")
		return
	}

	var req struct {
		Code   string `json:"code" binding:"required"`
		TypeID uint   `json:"type_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}

	brand, err := h.brandService.UpdateBrand(brandID, req.Code, req.TypeID)
	if err != nil {
		utils.Conflict(c, err.Error())
		return
	}

	utils.Success(c, gin.H{"data": brand}, "品牌更新成功")
}

// DeleteBrand 删除品牌
func (h *BrandHandler) DeleteBrand(c *gin.Context) {
	id := c.Param("id")
	brandID, err := strconv.Atoi(id)
	if err != nil {
		utils.BadRequest(c, "无效的品牌ID")
		return
	}

	err = h.brandService.DeleteBrand(brandID)
	if err != nil {
		if err.Error() == "cannot delete brand with existing clients" {
			utils.Conflict(c, "无法删除品牌：该品牌下还有客户端")
		} else {
			utils.InternalServerError(c, "删除品牌失败："+err.Error())
		}
		return
	}

	utils.Success(c, nil, "品牌删除成功")
}
