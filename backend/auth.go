package main

import (
	"log"
	"sync"

	"github.com/007Secret/007Password/database"
)

var (
	// 主密码，用于加解密数据库
	masterPassword string

	// 锁，用于保护主密码
	masterPasswordMutex sync.RWMutex
)

// SetMasterPassword 设置主密码并重新初始化数据库
func SetMasterPassword(password string) {
	if password == "" {
		log.Printf("警告: 尝试设置空的主密码")
		return
	}

	log.Printf("设置主密码，长度: %d", len(password))

	masterPasswordMutex.Lock()
	defer masterPasswordMutex.Unlock()

	// 保存密码
	masterPassword = password

	// 重新初始化数据库
	err := database.InitDBWithKey(password)
	if err != nil {
		log.Printf("使用主密码初始化数据库失败: %v", err)
	} else {
		log.Printf("使用主密码初始化数据库成功")
	}
}

// GetMasterPassword 获取当前的主密码
func GetMasterPassword() string {
	masterPasswordMutex.RLock()
	defer masterPasswordMutex.RUnlock()

	return masterPassword
}

// ValidateDatabaseConnection 验证数据库连接是否有效
func ValidateDatabaseConnection() bool {
	if database.DB == nil {
		log.Printf("数据库连接不存在")
		return false
	}

	err := database.DB.Ping()
	if err != nil {
		log.Printf("数据库连接失效: %v", err)

		// 尝试重新初始化数据库
		currentPassword := GetMasterPassword()
		if currentPassword != "" {
			log.Printf("尝试使用主密码重新初始化数据库")
			err := database.InitDBWithKey(currentPassword)
			if err != nil {
				log.Printf("重新初始化数据库失败: %v", err)
				return false
			}
			return true
		}

		return false
	}

	return true
}
