package controllers

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/007Secret/007Password/database"
	"github.com/007Secret/007Password/middleware"
	"github.com/007Secret/007Password/models"
	"github.com/007Secret/007Password/utils"
	"github.com/gin-gonic/gin"
)

// GetAllPasswords 获取所有密码
func GetAllPasswords(c *gin.Context) {
	log.Printf("获取所有密码列表...")

	// 验证数据库连接是否有效
	if database.DB == nil {
		log.Printf("💥 数据库连接不存在，尝试重新初始化")
		// 使用当前主密码重新初始化数据库
		currentPassword := middleware.GetMasterPassword()
		if currentPassword != "" {
			log.Printf("⚡ 使用当前主密码 '%s' 重新初始化数据库", maskPassword(currentPassword))
			err := database.InitDBWithKey(currentPassword)
			if err != nil {
				log.Printf("💥 重新初始化数据库失败: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库连接丢失", "code": "DB_CONNECTION_LOST"})
				return
			}
			log.Printf("✅ 重新初始化数据库成功")
		} else {
			log.Printf("💥 无法重新初始化数据库: 主密码为空")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无可用的主密码", "code": "NO_MASTER_PASSWORD"})
			return
		}
	}

	// 执行数据库查询前先Ping测试
	err := database.DB.Ping()
	if err != nil {
		log.Printf("💥 数据库连接不可用: %v", err)

		// 尝试重新连接
		currentPassword := middleware.GetMasterPassword()
		if currentPassword != "" {
			log.Printf("⚡ 尝试重新连接数据库...")
			err := database.InitDBWithKey(currentPassword)
			if err != nil {
				log.Printf("💥 重新连接数据库失败: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库连接错误", "code": "DB_CONNECTION_ERROR"})
				return
			}
			log.Printf("✅ 重新连接数据库成功")
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库连接错误", "code": "DB_CONNECTION_ERROR"})
			return
		}
	}

	passwords, err := database.GetAllPasswords()
	if err != nil {
		log.Printf("💥 获取密码列表失败: %v", err)

		if err.Error() == "file is not a database" {
			// 数据库文件可能损坏，尝试重新初始化
			currentPassword := middleware.GetMasterPassword()
			if currentPassword != "" {
				log.Printf("⚡ 数据库文件可能损坏，尝试强制重新初始化...")

				// 关闭当前连接
				if database.DB != nil {
					database.DB.Close()
					database.DB = nil
				}

				// 删除可能损坏的数据库文件
				dbPath := filepath.Join(database.GetDBFolder(), "passwordManager.db")
				os.Remove(dbPath)

				// 重新初始化
				err := database.InitDBWithKey(currentPassword)
				if err != nil {
					log.Printf("💥 强制重新初始化数据库失败: %v", err)
					c.JSON(http.StatusInternalServerError, gin.H{
						"error": "数据库文件已损坏且无法重建",
						"code":  "DB_CORRUPTED",
					})
					return
				}

				log.Printf("✅ 强制重新初始化数据库成功，返回空密码列表")
				c.JSON(http.StatusOK, []interface{}{})
				return
			}
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取密码失败", "details": err.Error()})
		return
	}

	log.Printf("✅ 成功获取密码列表，数量: %d", len(passwords))

	// 如果没有密码记录，直接返回空数组
	if len(passwords) == 0 {
		c.JSON(http.StatusOK, []interface{}{})
		return
	}

	// 解密所有密码字段
	var decryptedPasswords []models.Password
	decryptFailCount := 0

	for _, pwd := range passwords {
		pwdCopy := pwd
		if pwd.Password != "" {
			log.Printf("🔐 尝试解密密码ID=%d 名称=%s", pwd.ID, pwd.Name)
			decrypted, err := utils.DecryptPassword(pwd.Password)
			if err != nil {
				decryptFailCount++
				log.Printf("⚠️ 解密密码失败 ID=%d 名称=%s: %v", pwd.ID, pwd.Name, err)
				// 设置一个特殊的错误消息，让前端知道解密失败的原因
				errorMsg := "解密失败: " + err.Error()
				if len(errorMsg) > 50 {
					errorMsg = errorMsg[:50] + "..."
				}
				pwdCopy.Password = errorMsg
			} else {
				pwdCopy.Password = decrypted
			}
		}
		decryptedPasswords = append(decryptedPasswords, pwdCopy)
	}

	if decryptFailCount > 0 {
		log.Printf("⚠️ 警告: %d/%d 个密码解密失败", decryptFailCount, len(passwords))
		if decryptFailCount == len(passwords) {
			log.Printf("💥 严重错误: 所有密码解密均失败，可能是主密码错误或数据库加密密钥不匹配")

			// 检查salt值是否存在且有效
			salt, saltErr := database.GetSetting("password_salt")
			if saltErr != nil || salt == "" {
				log.Printf("🔍 检测到盐值不存在或获取失败: %v", saltErr)
				// 创建新的盐值
				newSalt := utils.GenerateSalt()
				if err := database.SetSetting("password_salt", newSalt); err == nil {
					log.Printf("✅ 已创建新的盐值: %s", newSalt[:8]+"...")
				} else {
					log.Printf("❌ 创建新盐值失败: %v", err)
				}
			} else {
				log.Printf("✅ 盐值存在且有效，长度: %d", len(salt))
			}

			// 尝试重新初始化数据库
			currentPassword := middleware.GetMasterPassword()
			if currentPassword != "" {
				log.Printf("⚡ 尝试用当前主密码重新初始化数据库")
				err := database.InitDBWithKey(currentPassword)
				if err != nil {
					log.Printf("💥 重新初始化数据库失败: %v", err)
				} else {
					log.Printf("✅ 重新初始化数据库成功")
				}
			}
		}
	}

	log.Printf("✅ 密码列表解密完成，返回结果")
	c.JSON(http.StatusOK, decryptedPasswords)
}

// 辅助函数: 遮蔽密码用于日志输出
func maskPassword(password string) string {
	if len(password) <= 4 {
		return "****"
	}
	return password[:2] + "****" + password[len(password)-2:]
}

// GetPasswordByID 通过ID获取密码
func GetPasswordByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	password, err := database.GetPasswordByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "未找到密码"})
		return
	}

	// 解密密码
	if password.Password != "" {
		decrypted, err := utils.DecryptPassword(password.Password)
		if err != nil {
			log.Printf("Error decrypting password for %s: %v", password.Name, err)
			// 不返回错误，只是记录日志
		} else {
			password.Password = decrypted
		}
	}

	c.JSON(http.StatusOK, password)
}

// CreatePassword 创建新密码
func CreatePassword(c *gin.Context) {
	var password models.Password
	if err := c.ShouldBindJSON(&password); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
		return
	}

	// 加密密码字段
	if password.Password != "" {
		encrypted, err := utils.EncryptPassword(password.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "加密密码失败"})
			return
		}
		password.Password = encrypted
	}

	id, err := database.CreatePassword(password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建密码失败"})
		return
	}

	// 获取创建后的密码记录（带解密密码）
	createdPassword, err := database.GetPasswordByID(int(id))
	if err == nil && createdPassword.Password != "" {
		decrypted, err := utils.DecryptPassword(createdPassword.Password)
		if err == nil {
			createdPassword.Password = decrypted
		}
	}

	c.JSON(http.StatusCreated, createdPassword)
}

// UpdatePassword 更新密码
func UpdatePassword(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var password models.Password
	if err := c.ShouldBindJSON(&password); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
		return
	}

	// 加密密码字段
	if password.Password != "" {
		encrypted, err := utils.EncryptPassword(password.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "加密密码失败"})
			return
		}
		password.Password = encrypted
	}

	password.ID = id
	if err := database.UpdatePassword(password); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新密码失败"})
		return
	}

	// 获取更新后的密码记录（带解密密码）
	updatedPassword, err := database.GetPasswordByID(id)
	if err == nil && updatedPassword.Password != "" {
		decrypted, err := utils.DecryptPassword(updatedPassword.Password)
		if err == nil {
			updatedPassword.Password = decrypted
		}
	}

	c.JSON(http.StatusOK, updatedPassword)
}

// DeletePassword 删除密码
func DeletePassword(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	if err := database.DeletePassword(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除密码失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password deleted successfully"})
}

// SearchPasswords 搜索密码
func SearchPasswords(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		GetAllPasswords(c)
		return
	}

	passwords, err := database.SearchPasswordsByName(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "搜索密码失败"})
		return
	}

	// 解密所有密码字段
	for i, pwd := range passwords {
		if pwd.Password != "" {
			decrypted, err := utils.DecryptPassword(pwd.Password)
			if err != nil {
				log.Printf("Error decrypting password for %s: %v", pwd.Name, err)
				// 不返回错误，只是记录日志
			} else {
				passwords[i].Password = decrypted
			}
		}
	}

	c.JSON(http.StatusOK, passwords)
}
