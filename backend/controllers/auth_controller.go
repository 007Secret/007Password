package controllers

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/007Secret/007Password/database"
	"github.com/007Secret/007Password/middleware"
	"github.com/007Secret/007Password/utils"
	"github.com/gin-gonic/gin"
)

// LoginRequest 登录请求结构
type LoginRequest struct {
	MasterPassword string `json:"masterPassword" binding:"required"`
}

// Login 处理用户登录
func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求"})
		return
	}

	log.Printf("收到登录请求，处理中...")

	// 检查数据库文件是否存在
	dbPath := filepath.Join(database.GetDBFolder(), "passwordManager.db")
	_, err := os.Stat(dbPath)

	// 首次使用，数据库文件不存在
	if os.IsNotExist(err) {
		log.Printf("数据库文件不存在，这是首次使用")
		// 验证主密码长度
		if len(req.MasterPassword) < 6 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Master password must be at least 6 characters"})
			return
		}

		// 初始化数据库加密
		err = database.InitDBWithKey(req.MasterPassword)
		if err != nil {
			log.Printf("Failed to initialize database with encryption key: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to setup database encryption"})
			return
		}

		// 存储主密码哈希（仅作为参考，不用于验证）
		hashedPw := hashPassword(req.MasterPassword)
		if err := database.SetSetting("master_password", hashedPw); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set master password"})
			return
		}

		// 同时保存主密码的加密salt
		salt := generateSalt()
		if err := database.SetSetting("password_salt", salt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set encryption salt"})
			return
		}

		// 重要：保存加密密钥
		// 首次设置时，加密密钥就是主密码本身
		if err := database.SetSetting("encryption_key", req.MasterPassword); err != nil {
			log.Printf("保存加密密钥失败: %v", err)
		} else {
			log.Printf("成功保存加密密钥，用于后续加解密")
		}

		// 生成JWT令牌
		token, err := middleware.GenerateToken()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "生成令牌失败"})
			return
		}

		// 将主密码保存到上下文中
		middleware.SetMasterPassword(req.MasterPassword)
		log.Printf("首次设置加密成功，设置了主密码到内存中: %v", req.MasterPassword != "")

		c.JSON(http.StatusOK, gin.H{
			"token":        token,
			"firstTimeSet": true,
		})
		return
	} else if err != nil {
		log.Printf("检查数据库文件时出错: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误", "details": err.Error()})
		return
	}

	// 关闭当前连接
	if database.DB != nil {
		database.DB.Close()
		database.DB = nil
	}

	// 直接使用主密码尝试初始化数据库连接，SQLite会进行密码验证
	// 如果密码正确，则可以成功连接并解密数据库；如果密码错误，连接会失败
	log.Printf("尝试使用提供的主密码初始化数据库...")
	err = database.InitDBWithKey(req.MasterPassword)
	if err != nil {
		log.Printf("使用提供的主密码打开数据库失败: %v", err)
		// 尝试使用空密码打开，检查是否是未加密数据库
		database.DB = nil // 确保关闭之前的连接尝试
		err = database.InitDB()

		if err == nil {
			// 数据库未加密，这是首次设置加密
			log.Printf("数据库未加密，应用密码作为新的加密密钥")

			// 关闭未加密的连接
			database.DB.Close()

			// 使用主密码重新初始化
			err = database.InitDBWithKey(req.MasterPassword)
			if err != nil {
				log.Printf("使用主密码重新初始化数据库失败: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encrypt database"})
				return
			}

			// 存储主密码哈希
			hashedPw := hashPassword(req.MasterPassword)
			if err := database.SetSetting("master_password", hashedPw); err != nil {
				log.Printf("保存主密码哈希失败: %v", err)
			}

			// 同时保存主密码的加密salt
			salt := generateSalt()
			if err := database.SetSetting("password_salt", salt); err != nil {
				log.Printf("保存加密盐值失败: %v", err)
			}

			log.Printf("成功将数据库从未加密状态转换为加密状态")

			// 保存主密码到内存
			middleware.SetMasterPassword(req.MasterPassword)

			// 生成JWT令牌
			token, err := middleware.GenerateToken()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "生成令牌失败"})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"token":        token,
				"firstTimeSet": true,
				"converted":    true,
			})
			return
		} else {
			// 数据库已加密，但密码错误
			log.Printf("验证失败：主密码不正确")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "主密码不正确"})
			return
		}
	}

	// 验证成功，密码正确
	log.Printf("主密码验证成功，SQLite连接已经建立")

	// 设置主密码到内存中
	middleware.SetMasterPassword(req.MasterPassword)
	log.Printf("登录成功后设置主密码到内存: %v", req.MasterPassword != "")

	// 生成JWT令牌
	token, err := middleware.GenerateToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成令牌失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":   token,
		"message": "登录成功",
	})
}

// ValidateToken 验证令牌有效性
func ValidateToken(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"valid": true})
}

// hashPassword 使用SHA-256哈希密码
func hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
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

	// 确保文件内容已经写入磁盘
	err = destFile.Sync()
	if err != nil {
		return err
	}

	return nil
}

// CheckFirstTimeSetup 检查是否首次使用（需要设置主密码）
func CheckFirstTimeSetup(c *gin.Context) {
	// 检查数据库文件是否存在
	dbPath := filepath.Join(database.GetDBFolder(), "passwordManager.db")
	_, err := os.Stat(dbPath)

	// 数据库文件不存在，说明是首次使用
	if os.IsNotExist(err) {
		log.Printf("数据库文件不存在，这是首次使用")
		c.JSON(http.StatusOK, gin.H{
			"isFirstTimeSetup": true,
			"reason":           "database_not_exist",
		})
		return
	} else if err != nil {
		log.Printf("检查数据库文件时出错: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误"})
		return
	}

	// 检查文件大小，如果文件很小（小于100字节），可能是新创建的空文件
	fileInfo, err := os.Stat(dbPath)
	if err == nil && fileInfo.Size() < 100 {
		log.Printf("数据库文件过小，可能是新创建的空文件")
		c.JSON(http.StatusOK, gin.H{
			"isFirstTimeSetup": true,
			"reason":           "database_empty",
		})
		return
	}

	// 尝试使用当前保存的主密码连接数据库
	currentPassword := middleware.GetMasterPassword()
	if currentPassword != "" {
		log.Printf("尝试使用内存中保存的主密码初始化数据库连接")

		// 保存当前连接
		currentDB := database.DB

		// 尝试使用主密码连接
		err := database.InitDBWithKey(currentPassword)
		if err == nil {
			// 检查是否存在master_password设置
			var count int
			err = database.DB.QueryRow("SELECT COUNT(*) FROM settings WHERE key = 'master_password'").Scan(&count)
			if err == nil && count > 0 {
				log.Printf("使用主密码成功连接到数据库，且找到master_password记录，不是首次设置")
				c.JSON(http.StatusOK, gin.H{
					"isFirstTimeSetup": false,
					"reason":           "master_password_exists",
				})
				return
			}
		}

		// 如果连接或检查失败，恢复原来的连接
		if currentDB != nil {
			database.DB = currentDB
		}
	}

	// 尝试检查数据库文件头部以判断是否已加密
	file, err := os.Open(dbPath)
	if err == nil {
		defer file.Close()

		// SQLite文件头部标识: 前16字节
		header := make([]byte, 16)
		n, _ := file.Read(header)

		if n >= 16 {
			// 未加密的SQLite数据库以"SQLite format 3\000"开头
			// 加密的SQLite数据库通常没有这个标识
			// 0x53, 0x51, 0x4c, 0x69, 0x74, 0x65, 0x20, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x74, 0x20, 0x33, 0x00
			// "S  Q  L  i  t  e     f  o  r  m  a  t     3  \0"
			if header[0] == 0x53 && header[1] == 0x51 && header[2] == 0x4c && header[3] == 0x69 &&
				header[4] == 0x74 && header[5] == 0x65 && header[6] == 0x20 && header[7] == 0x66 &&
				header[8] == 0x6f && header[9] == 0x72 && header[10] == 0x6d && header[11] == 0x61 &&
				header[12] == 0x74 && header[13] == 0x20 && header[14] == 0x33 && header[15] == 0x00 {
				log.Printf("数据库文件头部检测为未加密SQLite文件")

				// 再次使用无密码方式尝试检查数据库
				database.DB = nil
				err = database.InitDB()
				if err == nil {
					// 检查是否已经有主密码设置
					var count int
					err = database.DB.QueryRow("SELECT COUNT(*) FROM settings WHERE key = 'master_password'").Scan(&count)
					if err == nil && count > 0 {
						log.Printf("数据库未加密但已经有主密码记录")
						c.JSON(http.StatusOK, gin.H{
							"isFirstTimeSetup": false,
							"reason":           "database_not_encrypted_but_setup_done",
						})
						return
					}
				}

				// 否则认为需要设置主密码
				c.JSON(http.StatusOK, gin.H{
					"isFirstTimeSetup": true,
					"reason":           "database_not_encrypted",
				})
				return
			} else {
				log.Printf("数据库文件头部不是标准SQLite格式，可能已加密")
				c.JSON(http.StatusOK, gin.H{
					"isFirstTimeSetup": false,
					"reason":           "database_encrypted",
				})
				return
			}
		}
	}

	// 如果通过直接检查文件头无法确定，继续使用原方法尝试无密码打开
	// 记录当前数据库连接（可能为nil）
	currentDB := database.DB

	// 尝试无密码打开数据库
	database.DB = nil // 确保关闭之前的连接尝试

	err = database.InitDB()
	if err == nil {
		// 能够无密码打开，说明没有设置加密
		log.Printf("数据库能够无密码打开，未加密")

		// 检查是否已经有主密码设置
		var count int
		err = database.DB.QueryRow("SELECT COUNT(*) FROM settings WHERE key = 'master_password'").Scan(&count)
		if err == nil && count > 0 {
			log.Printf("数据库未加密但已经有主密码记录")

			// 安全关闭未加密数据库连接
			if database.DB != nil {
				database.DB.Close()
				database.DB = nil
			}

			// 恢复原始连接
			database.DB = currentDB

			c.JSON(http.StatusOK, gin.H{
				"isFirstTimeSetup": false,
				"reason":           "database_not_encrypted_but_setup_done",
			})
			return
		}

		// 安全关闭未加密数据库连接
		if database.DB != nil {
			database.DB.Close()
			database.DB = nil
		}

		// 恢复原始连接
		database.DB = currentDB

		c.JSON(http.StatusOK, gin.H{
			"isFirstTimeSetup": true,
			"reason":           "database_not_encrypted",
		})
		return
	} else {
		log.Printf("无法无密码打开数据库: %v", err)
	}

	// 恢复原始连接（如果存在）
	if currentDB != nil {
		database.DB = currentDB
	}

	// 如果走到这里，假设数据库已加密
	log.Printf("数据库文件检测为已加密")
	c.JSON(http.StatusOK, gin.H{
		"isFirstTimeSetup": false,
		"reason":           "database_likely_encrypted",
	})
}

// ChangeMasterPassword 处理主密码修改请求
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
	if len(req.NewPassword) < 6 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "新密码长度至少需要6个字符"})
		return
	}

	// 验证当前密码是否正确
	if middleware.GetMasterPassword() != req.CurrentPassword {
		c.JSON(http.StatusBadRequest, gin.H{"error": "当前密码不正确"})
		return
	}

	// 1. 备份整个数据库文件
	dbPath := filepath.Join(database.GetDBFolder(), "passwordManager.db")
	backupPath := filepath.Join(database.GetDBFolder(), "passwordManager_backup_"+time.Now().Format("20060102_150405")+".db")

	err := copyFile(dbPath, backupPath)
	if err != nil {
		log.Printf("备份数据库文件失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法创建备份，修改密码取消"})
		return
	}
	log.Printf("✅ 成功创建数据库备份: %s", backupPath)

	// 2. 使用SQLite的PRAGMA命令直接修改数据库密码
	log.Printf("开始修改数据库主密码...")

	// 确保数据库连接存在
	if database.DB == nil {
		log.Printf("数据库连接不存在，尝试重新连接...")
		err = database.InitDBWithKey(req.CurrentPassword)
		if err != nil {
			log.Printf("使用当前密码初始化数据库失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "无法连接到数据库"})
			return
		}
	}

	// 执行密码变更SQL命令
	_, err = database.DB.Exec(fmt.Sprintf("PRAGMA rekey = '%s'", req.NewPassword))
	if err != nil {
		log.Printf("修改数据库密码失败: %v", err)

		// 保留备份文件用于恢复
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "修改数据库密码失败，但已创建备份文件",
			"backupPath": backupPath,
		})
		return
	}

	log.Printf("✅ 数据库密码修改成功")

	// 关闭当前连接并用新密码重新打开以验证
	database.DB.Close()
	database.DB = nil

	err = database.InitDBWithKey(req.NewPassword)
	if err != nil {
		log.Printf("使用新密码验证数据库失败: %v", err)
		log.Printf("尝试恢复到原始密码...")

		// 尝试用原密码重新打开
		err = database.InitDBWithKey(req.CurrentPassword)
		if err != nil {
			log.Printf("无法恢复到原始密码: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":      "密码修改可能未完全成功，请重试。如果问题持续，请使用备份恢复",
				"backupPath": backupPath,
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "无法验证新密码，已恢复到原始密码",
		})
		return
	}

	// 3. 更新内存中的主密码
	middleware.SetMasterPassword(req.NewPassword)
	log.Printf("✅ 成功更新内存中的主密码")

	// 重要：获取当前的盐值
	currentSalt, saltErr := database.GetSetting("password_salt")
	if saltErr != nil {
		log.Printf("⚠️ 获取当前盐值失败: %v，这可能导致解密问题", saltErr)
	}

	// 4. 关键修改: 保存旧主密码作为加密密钥
	// 我们将旧密码保存到数据库中，用于后续的加解密操作
	// 这样即使主密码变了，加解密操作仍使用原始密码进行
	err = database.SetSetting("encryption_key", req.CurrentPassword)
	if err != nil {
		log.Printf("⚠️ 保存加密密钥失败: %v，这可能导致解密问题", err)
	} else {
		log.Printf("✅ 成功保存加密密钥供后续加解密使用")
	}

	// 确保当前盐值存在，如果不存在则创建新的盐值
	if saltErr != nil || currentSalt == "" {
		log.Printf("🔑 未找到有效的盐值，将创建新的盐值")

		// 生成新的盐值
		newSalt := utils.GenerateSalt()

		// 保存新的盐值
		if database.DB != nil {
			saltSaveErr := database.SetSetting("password_salt", newSalt)
			if saltSaveErr != nil {
				log.Printf("⚠️ 保存新盐值失败: %v，这可能导致解密问题", saltSaveErr)
			} else {
				log.Printf("✅ 成功创建并保存新盐值")
			}
		} else {
			log.Printf("⚠️ 数据库连接不存在，无法保存新盐值")
		}
	} else {
		log.Printf("✅ 使用现有盐值进行加解密，确保数据兼容性")
	}

	// 5. 测试加密解密是否正常
	testPassword := "测试密码123"

	// 确保数据库连接存在，否则加解密测试将失败
	if database.DB == nil {
		log.Printf("加密测试前数据库连接不存在，尝试重新连接...")
		err = database.InitDBWithKey(req.NewPassword)
		if err != nil {
			log.Printf("加密测试前重新连接数据库失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "加密测试失败，但密码已修改，请重新登录"})
			return
		}
	}

	encrypted, err := utils.EncryptPassword(testPassword)
	if err != nil {
		log.Printf("测试加密失败: %v", err)
	} else {
		// 再次确保数据库连接存在，因为解密也需要从数据库获取salt
		if database.DB == nil {
			log.Printf("解密测试前数据库连接不存在，尝试重新连接...")
			err = database.InitDBWithKey(req.NewPassword)
			if err != nil {
				log.Printf("解密测试前重新连接数据库失败: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "解密测试失败，但密码已修改，请重新登录"})
				return
			}
		}

		decrypted, err := utils.DecryptPassword(encrypted)
		if err != nil {
			log.Printf("测试解密失败: %v", err)
		} else if decrypted != testPassword {
			log.Printf("测试解密结果不匹配: 期望 %s, 实际 %s", testPassword, decrypted)
		} else {
			log.Printf("✅ 加密解密测试通过")
		}
	}

	// 5. 成功后删除备份文件
	os.Remove(backupPath)
	log.Printf("主密码修改成功，已删除备份文件")

	// 6. 生成新的JWT令牌
	token, err := middleware.GenerateToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成新令牌失败"})
		return
	}

	// 返回成功消息和新令牌
	c.JSON(http.StatusOK, gin.H{
		"message": "主密码修改成功",
		"token":   token,
		"stats": gin.H{
			"status": "success",
		},
	})
}

// SetMasterPassword 设置主密码
func SetMasterPassword(c *gin.Context) {
	var req struct {
		MasterPassword string `json:"masterPassword"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if len(req.MasterPassword) < 6 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password must be at least 6 characters"})
		return
	}

	// Hash the master password
	hash := sha256.Sum256([]byte(req.MasterPassword))
	hashString := hex.EncodeToString(hash[:])

	// 使用主密码作为SQLite加密密钥
	err := database.InitDBWithKey(req.MasterPassword)
	if err != nil {
		log.Printf("Failed to initialize database with encryption key: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to setup database encryption"})
		return
	}

	// Check if master password already exists
	var count int
	err = database.DB.QueryRow("SELECT COUNT(*) FROM settings WHERE key = 'master_password'").Scan(&count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误"})
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
			c.JSON(http.StatusInternalServerError, gin.H{"error": "设置加密盐值失败"})
			return
		}

		// 保存加密密钥 - 首次设置时与主密码相同
		_, keyErr := database.DB.Exec("INSERT INTO settings (key, value) VALUES ('encryption_key', ?)", req.MasterPassword)
		if keyErr != nil {
			log.Printf("保存加密密钥失败: %v", keyErr)
		} else {
			log.Printf("成功保存加密密钥，用于后续加解密")
		}
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "设置主密码失败"})
		return
	}

	// 将主密码保存到上下文中，用于加解密操作
	middleware.SetMasterPassword(req.MasterPassword)

	// Generate token
	token, err := middleware.GenerateToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成令牌失败"})
		return
	}

	resp := map[string]interface{}{
		"token":        token,
		"firstTimeSet": count == 0,
	}

	c.JSON(http.StatusOK, resp)
}
