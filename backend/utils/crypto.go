package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"

	"github.com/007Secret/007Password/database"
	"github.com/007Secret/007Password/middleware"
)

// EncryptPassword 使用主密码加密密码字段
func EncryptPassword(password string) (string, error) {
	// 获取加密密钥 - 优先使用存储的加密密钥，如果不存在则使用主密码
	encryptionKey := ""

	// 检查数据库连接是否存在
	if database.DB == nil {
		log.Printf("加密失败: 数据库连接不存在")
		return "", errors.New("数据库连接不可用")
	}

	// 首先尝试获取存储的加密密钥
	storedKey, err := database.GetSetting("encryption_key")
	if err == nil && storedKey != "" {
		log.Printf("使用存储的加密密钥进行加密，长度: %d", len(storedKey))
		encryptionKey = storedKey
	} else {
		// 如果没有存储的加密密钥，则使用当前主密码
		masterPassword := middleware.GetMasterPassword()
		if masterPassword == "" {
			log.Printf("加密失败: 主密码为空且无存储的加密密钥")
			return "", errors.New("无可用的加密密钥")
		}

		log.Printf("未找到存储的加密密钥，使用当前主密码，长度: %d", len(masterPassword))
		encryptionKey = masterPassword

		// 首次使用时，将主密码保存为加密密钥
		if storedKey == "" && err != nil && err.Error() == "sql: no rows in result set" {
			saveErr := database.SetSetting("encryption_key", masterPassword)
			if saveErr == nil {
				log.Printf("首次使用，保存当前主密码作为加密密钥")
			} else {
				log.Printf("保存加密密钥失败: %v", saveErr)
			}
		}
	}

	// 获取存储的盐值
	salt, err := database.GetSetting("password_salt")
	if err != nil {
		// 如果盐值不存在，则创建一个新的
		if err.Error() == "sql: no rows in result set" {
			log.Printf("未找到密码盐值，创建新的盐值")
			salt = generateSalt()
			err = database.SetSetting("password_salt", salt)
			if err != nil {
				log.Printf("创建盐值失败: %v", err)
				return "", err
			}
			log.Printf("成功创建新的盐值，长度: %d", len(salt))
		} else {
			log.Printf("加密失败: 获取盐值出错: %v", err)
			return "", err
		}
	}

	log.Printf("加密密码: 找到盐值，长度: %d", len(salt))

	// 创建加密密钥
	key := deriveKey(encryptionKey, salt)

	// 生成随机IV
	iv := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		log.Printf("加密失败: 生成IV出错: %v", err)
		return "", err
	}

	// 创建加密器
	block, err := aes.NewCipher(key)
	if err != nil {
		log.Printf("加密失败: 创建AES加密器出错: %v", err)
		return "", err
	}

	// GCM模式
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		log.Printf("加密失败: 创建GCM模式出错: %v", err)
		return "", err
	}

	// 加密数据
	ciphertext := aesgcm.Seal(nil, iv, []byte(password), nil)

	// 将IV和密文组合并转为base64
	combined := append(iv, ciphertext...)
	encoded := base64.StdEncoding.EncodeToString(combined)
	log.Printf("密码加密成功，长度: %d", len(encoded))
	return encoded, nil
}

// DecryptPassword 使用主密码解密密码字段
func DecryptPassword(encryptedPassword string) (string, error) {
	if encryptedPassword == "" {
		log.Printf("解密失败: 加密的密码为空")
		return "", errors.New("加密的密码为空")
	}

	// 获取解密密钥 - 优先使用存储的加密密钥，如果不存在则使用主密码
	decryptionKey := ""

	// 检查数据库连接是否存在
	if database.DB == nil {
		log.Printf("解密失败: 数据库连接不存在")
		return "", errors.New("数据库连接不可用")
	}

	// 首先尝试获取存储的加密密钥
	storedKey, err := database.GetSetting("encryption_key")
	if err == nil && storedKey != "" {
		log.Printf("使用存储的加密密钥进行解密，长度: %d", len(storedKey))
		decryptionKey = storedKey
	} else {
		// 如果没有存储的加密密钥，则使用当前主密码
		masterPassword := middleware.GetMasterPassword()
		if masterPassword == "" {
			log.Printf("解密失败: 主密码为空且无存储的加密密钥")
			return "", errors.New("无可用的解密密钥")
		}

		log.Printf("未找到存储的加密密钥，使用当前主密码，长度: %d", len(masterPassword))
		decryptionKey = masterPassword
	}

	// 获取存储的盐值
	salt, err := database.GetSetting("password_salt")
	if err != nil {
		log.Printf("解密失败: 获取盐值出错: %v", err)
		return "", fmt.Errorf("获取盐值失败: %w", err)
	}

	if salt == "" {
		log.Printf("⚠️ 警告: 盐值为空，这可能导致解密失败")
		// 尝试创建新的盐值
		salt = generateSalt()
		saveErr := database.SetSetting("password_salt", salt)
		if saveErr != nil {
			log.Printf("⚠️ 创建新盐值失败: %v", saveErr)
		} else {
			log.Printf("✅ 创建了新的盐值: %s", salt[:8]+"...")
		}
	}

	log.Printf("解密密码: 找到盐值，长度: %d", len(salt))

	// 创建加密密钥
	key := deriveKey(decryptionKey, salt)
	keyPrefix := fmt.Sprintf("%x", key[:4])
	log.Printf("🔑 使用密钥前缀 %s... 进行解密", keyPrefix)

	// 解码base64
	combined, err := base64.StdEncoding.DecodeString(encryptedPassword)
	if err != nil {
		log.Printf("解密失败: base64解码出错: %v, 加密字符串: %s", err, encryptedPassword)
		return "", fmt.Errorf("base64解码失败: %w", err)
	}

	// 提取IV和密文
	if len(combined) < 12 {
		log.Printf("解密失败: 无效的加密格式，长度过短: %d", len(combined))
		return "", fmt.Errorf("无效的加密密码格式，长度过短: %d", len(combined))
	}
	iv := combined[:12]
	ciphertext := combined[12:]

	// 创建加密器
	block, err := aes.NewCipher(key)
	if err != nil {
		log.Printf("解密失败: 创建AES加密器出错: %v", err)
		return "", fmt.Errorf("创建AES加密器失败: %w", err)
	}

	// GCM模式
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		log.Printf("解密失败: 创建GCM模式出错: %v", err)
		return "", fmt.Errorf("创建GCM模式失败: %w", err)
	}

	// 解密数据
	plaintext, err := aesgcm.Open(nil, iv, ciphertext, nil)
	if err != nil {
		log.Printf("解密失败: GCM.Open出错: %v", err)
		return "", fmt.Errorf("GCM解密失败，可能是密钥或salt不匹配: %w", err)
	}

	log.Printf("密码解密成功")
	return string(plaintext), nil
}

// deriveKey 根据主密码和盐值生成加密密钥
func deriveKey(masterPassword, salt string) []byte {
	// 简单的密钥派生，实际应用中可以使用PBKDF2或Argon2
	saltBytes, err := hex.DecodeString(salt)
	if err != nil {
		log.Printf("密钥派生出错: 盐值解码失败: %v", err)
		// 防止程序崩溃，使用空盐值继续
		saltBytes = []byte{}
	}
	combined := append([]byte(masterPassword), saltBytes...)
	hash := sha256.Sum256(combined)
	return hash[:]
}

// generateSalt 生成随机盐值
func generateSalt() string {
	salt := make([]byte, 16)
	// 使用安全随机数生成器
	_, err := io.ReadFull(rand.Reader, salt)
	if err != nil {
		// 如果随机生成失败，使用备用方法
		log.Printf("随机生成盐值失败，使用备用方法: %v", err)
		for i := 0; i < 16; i++ {
			salt[i] = byte(i + 1)
		}
	}
	return hex.EncodeToString(salt)
}

// GenerateSalt 生成随机盐值（公开API）
func GenerateSalt() string {
	return generateSalt()
}
