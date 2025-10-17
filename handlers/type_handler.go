package handlers

import (
	"strconv"

	"brand-config-api/services"
	"brand-config-api/utils"

	"github.com/gin-gonic/gin"
)

// TypeHandler 类型控制器
type TypeHandler struct {
	typeService *services.TypeService
}

// NewTypeHandler 创建类型控制器
func NewTypeHandler() *TypeHandler {
	return &TypeHandler{
		typeService: services.NewTypeService(),
	}
}

// GetTypes 获取所有类型
func (h *TypeHandler) GetTypes(c *gin.Context) {
	types, err := h.typeService.GetAllTypes()
	if err != nil {
		utils.InternalServerError(c, "获取类型列表失败")
		return
	}

	utils.Success(c, gin.H{
		"data":  types,
		"total": len(types),
	}, "获取类型列表成功")
}

// GetType 获取单个类型
func (h *TypeHandler) GetType(c *gin.Context) {
	id := c.Param("id")
	typeID, err := strconv.Atoi(id)
	if err != nil {
		utils.BadRequest(c, "无效的类型ID")
		return
	}

	typeData, err := h.typeService.GetTypeByID(uint(typeID))
	if err != nil {
		utils.NotFound(c, "类型不存在")
		return
	}

	utils.Success(c, gin.H{"data": typeData}, "获取类型成功")
}

// CreateType 创建类型
func (h *TypeHandler) CreateType(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
		Code string `json:"code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}

	typeData, err := h.typeService.CreateType(req.Name, req.Code)
	if err != nil {
		utils.Conflict(c, err.Error())
		return
	}

	utils.Created(c, gin.H{"data": typeData}, "类型创建成功")
}

// UpdateType 更新类型
func (h *TypeHandler) UpdateType(c *gin.Context) {
	id := c.Param("id")
	typeID, err := strconv.Atoi(id)
	if err != nil {
		utils.BadRequest(c, "无效的类型ID")
		return
	}

	var req struct {
		Name string `json:"name" binding:"required"`
		Code string `json:"code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误")
		return
	}

	typeData, err := h.typeService.UpdateType(uint(typeID), req.Name, req.Code)
	if err != nil {
		utils.Conflict(c, err.Error())
		return
	}

	utils.Success(c, gin.H{"data": typeData}, "类型更新成功")
}

// DeleteType 删除类型
func (h *TypeHandler) DeleteType(c *gin.Context) {
	id := c.Param("id")
	typeID, err := strconv.Atoi(id)
	if err != nil {
		utils.BadRequest(c, "无效的类型ID")
		return
	}

	err = h.typeService.DeleteType(uint(typeID))
	if err != nil {
		if err.Error() == "cannot delete type with existing brands" {
			utils.Conflict(c, "无法删除类型：该类型下还有品牌")
		} else {
			utils.InternalServerError(c, "删除类型失败："+err.Error())
		}
		return
	}

	utils.Success(c, nil, "类型删除成功")
}
