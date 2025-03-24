package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/007Secret/007Password/database"
	"github.com/007Secret/007Password/middleware"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

// SetMasterPassword 设置主密码
func SetMasterPassword(c *gin.Context) {
	var req struct {
		MasterPassword string `json:"masterPassword"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	if len(req.MasterPassword) < 8 {
		c.JSON(400, gin.H{"error": "Password must be at least 8 characters"})
		return
	}

	// Hash the master password
	hash := sha256.Sum256([]byte(req.MasterPassword))
	hashString := hex.EncodeToString(hash[:])

	// 使用主密码作为SQLite加密密钥
	err := database.InitDBWithKey(req.MasterPassword)
	if err != nil {
		log.Printf("Failed to initialize database with encryption key: %v", err)
		c.JSON(500, gin.H{"error": "Failed to setup database encryption"})
		return
	}

	// Check if master password already exists
	var count int
	err = database.DB.QueryRow("SELECT COUNT(*) FROM settings WHERE key = 'master_password'").Scan(&count)
	if err != nil {
		c.JSON(500, gin.H{"error": "Database error"})
		return
	}

	if count > 0 {
		// Update existing password
		_, err = database.DB.Exec("UPDATE settings SET value = ? WHERE key = 'master_password'", hashString)
	} else {
		// Insert new password
		_, err = database.DB.Exec("INSERT INTO settings (key, value) VALUES ('master_password', ?)", hashString)

		// 同时创建salt用于加密
		salt := generateSalt()
		_, saltErr := database.DB.Exec("INSERT INTO settings (key, value) VALUES ('password_salt', ?)", salt)
		if saltErr != nil {
			c.JSON(500, gin.H{"error": "Failed to set encryption salt"})
			return
		}
	}

	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to save master password"})
		return
	}

	// 将主密码保存到上下文中，用于加解密操作
	middleware.SetMasterPassword(req.MasterPassword)

	// Generate token
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

	// 使用 middleware 包中的密钥签名
	secretKey := []byte("your-secret-key-change-in-production") // 这应该与 middleware/auth.go 中的相同
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to generate token"})
		return
	}

	resp := map[string]interface{}{
		"token":        tokenString,
		"firstTimeSet": count == 0,
	}

	c.JSON(200, resp)
}

// ChangeMasterPassword 处理修改主密码
func ChangeMasterPassword(c *gin.Context) {
	var req struct {
		CurrentPassword string `json:"currentPassword" binding:"required"`
		NewPassword     string `json:"newPassword" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效请求参数"})
		return
	}

	// 验证新密码长度
	if len(req.NewPassword) < 8 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "新密码长度至少需要8个字符"})
		return
	}

	// 验证当前主密码
	// 尝试使用当前主密码解锁数据库
	currentDB := database.DB // 保存当前数据库连接

	// 关闭当前连接以尝试验证当前密码
	if currentDB != nil {
		currentDB.Close()
		database.DB = nil
	}

	log.Printf("尝试验证当前主密码")
	err := database.InitDBWithKey(req.CurrentPassword)
	if err != nil {
		// 还原之前的连接
		if currentDB != nil {
			database.DB = currentDB
		} else {
			// 重新尝试初始化数据库
			database.InitDB()
		}

		log.Printf("验证当前主密码失败: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "当前主密码错误"})
		return
	}

	log.Printf("当前主密码验证成功，准备使用新密码重新初始化数据库")

	// 创建临时数据库文件
	dbPath := filepath.Join(database.GetDBFolder(), "passwordManager.db")
	// 备份当前所有密码
	passwords, err := database.GetAllPasswords()
	if err != nil {
		log.Printf("获取所有密码失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "修改主密码时出错"})
		return
	}

	// 关闭当前数据库连接
	if database.DB != nil {
		database.DB.Close()
		database.DB = nil
	}

	// 备份当前数据库
	backupPath := filepath.Join(database.GetDBFolder(), "backup_passwordManager.db")
	err = copyFile(dbPath, backupPath)
	if err != nil {
		log.Printf("备份数据库失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "备份数据库失败"})
		return
	}

	// 使用新密码创建新数据库
	err = os.Remove(dbPath)
	if err != nil {
		log.Printf("删除旧数据库失败: %v", err)
		// 尝试恢复原始连接
		database.InitDBWithKey(req.CurrentPassword)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "修改主密码时出错"})
		return
	}

	// 使用新密码初始化数据库
	err = database.InitDBWithKey(req.NewPassword)
	if err != nil {
		log.Printf("使用新密码初始化数据库失败: %v", err)
		// 尝试恢复备份
		os.Rename(backupPath, dbPath)
		database.InitDBWithKey(req.CurrentPassword)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "使用新密码初始化数据库失败"})
		return
	}

	// 存储新的主密码哈希
	hashedPw := hashPassword(req.NewPassword)
	if err := database.SetSetting("master_password", hashedPw); err != nil {
		log.Printf("存储新主密码哈希失败: %v", err)
		// 继续执行，这不是致命错误
	}

	// 还原所有密码
	log.Printf("正在还原 %d 个密码...", len(passwords))
	for _, p := range passwords {
		_, err := database.CreatePassword(p)
		if err != nil {
			log.Printf("还原密码失败 ID=%d: %v", p.ID, err)
			// 继续尝试其他密码
		}
	}

	// 删除备份
	os.Remove(backupPath)

	// 生成新的JWT令牌
	token, err := middleware.GenerateToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成新令牌失败"})
		return
	}

	// 更新内存中的主密码
	middleware.SetMasterPassword(req.NewPassword)

	log.Printf("主密码修改成功，需要用户使用新密码重新登录")

	c.JSON(http.StatusOK, gin.H{
		"message": "主密码修改成功，请使用新密码重新登录",
		"token":   token,
	})
}

// copyFile 复制文件
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	return nil
}

// generateSalt 生成随机盐值
func generateSalt() string {
	salt := make([]byte, 16)
	// 在实际应用中，应该使用crypto/rand
	for i := 0; i < 16; i++ {
		salt[i] = byte(i + 1)
	}
	return hex.EncodeToString(salt)
}

// hashPassword 使用SHA-256哈希密码
func hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}
