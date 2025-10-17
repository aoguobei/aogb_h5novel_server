package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"os/exec"
	"time"

	"brand-config-api/config"

	"log"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

type DatabaseHandler struct {
	db *gorm.DB
}

type ExportDatabaseRequest struct {
	Password string `json:"password" binding:"required"`
}

// NewDatabaseHandler 创建数据库处理器
func NewDatabaseHandler(db *gorm.DB) *DatabaseHandler {
	return &DatabaseHandler{
		db: db,
	}
}

// ExportDatabase 导出数据库
func (h *DatabaseHandler) ExportDatabase(c *gin.Context) {
	var req ExportDatabaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "请求参数错误",
		})
		return
	}

	// 再次验证密码
	if err := h.testDatabaseConnection(req.Password); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "密码验证失败",
		})
		return
	}

	// 执行数据库导出
	log.Printf("开始导出数据库，用户: %s", config.Load().Database.User)
	sqlContent, err := h.exportDatabaseToSQL(req.Password)
	if err != nil {
		log.Printf("数据库导出失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "数据库导出失败: " + err.Error(),
		})
		return
	}

	log.Printf("数据库导出成功，内容长度: %d", len(sqlContent))

	// 设置响应头，直接返回文件
	filename := fmt.Sprintf("database_export_%s.sql", time.Now().Format("20060102_150405"))
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Length", fmt.Sprintf("%d", len(sqlContent)))
	c.Data(http.StatusOK, "application/octet-stream", []byte(sqlContent))
}

// testDatabaseConnection 测试数据库连接
func (h *DatabaseHandler) testDatabaseConnection(password string) error {
	// 获取数据库配置
	cfg := config.Load()

	// 构建测试连接字符串
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.Database.User,
		password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name)

	// 尝试连接
	testDB, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}
	defer testDB.Close()

	// 测试连接
	if err := testDB.Ping(); err != nil {
		return err
	}

	return nil
}

// exportDatabaseToSQL 导出数据库为SQL
func (h *DatabaseHandler) exportDatabaseToSQL(password string) (string, error) {
	// 获取数据库配置
	cfg := config.Load()

	// 构建mysqldump命令
	args := []string{
		"-u", cfg.Database.User,
		"-p" + password,
		"-h", cfg.Database.Host,
		"-P", cfg.Database.Port,
		"--single-transaction",
		"--routines",
		"--triggers",
		cfg.Database.Name,
	}

	cmd := exec.Command("mysqldump", args...)

	// 获取命令输出和错误
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("mysqldump执行失败: %v, 输出: %s", err, string(output))
	}
	return string(output), nil
}
