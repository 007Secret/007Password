package middleware

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/007Secret/007Password/database"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

const (
	secretKey = "your-secret-key-change-in-production" // 在生产环境中应当使用环境变量设置密钥
)

var (
	// masterPassword存储，用于加解密操作
	masterPassword     string
	masterPasswordLock sync.RWMutex
)

// Claims JWT的声明结构
type Claims struct {
	jwt.StandardClaims
}

// SetMasterPassword 设置主密码供加解密使用
func SetMasterPassword(password string) {
	log.Printf("设置主密码到内存中，长度: %d", len(password))
	masterPasswordLock.Lock()
	defer masterPasswordLock.Unlock()

	oldPassword := masterPassword
	masterPassword = password

	// 如果密码已更改，确保在数据库连接中也使用新密码
	if oldPassword != password && password != "" {
		// 通知数据库模块更新连接
		log.Printf("密码已更改，重新初始化数据库连接")

		// 异步重新初始化数据库，避免死锁
		go func() {
			err := database.InitDBWithKey(password)
			if err != nil {
				log.Printf("使用新密码重新初始化数据库失败: %v", err)
			} else {
				log.Printf("使用新密码重新初始化数据库成功")
			}
		}()
	}
}

// GetMasterPassword 获取主密码
func GetMasterPassword() string {
	masterPasswordLock.RLock()
	defer masterPasswordLock.RUnlock()
	passwordLen := 0
	if masterPassword != "" {
		passwordLen = len(masterPassword)
	}
	log.Printf("获取主密码，长度: %d, 是否为空: %v", passwordLen, masterPassword == "")
	return masterPassword
}

// GenerateToken 生成JWT令牌
func GenerateToken() (string, error) {
	// 创建声明
	claims := Claims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24).Unix(), // 24小时后过期
			IssuedAt:  time.Now().Unix(),
		},
	}

	// 使用密钥创建token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		log.Printf("Error generating token: %v", err)
		return "", err
	}

	log.Printf("Generated new token: %s...", tokenString[:10])
	return tokenString, nil
}

// AuthRequired 验证JWT令牌的中间件
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Printf("Processing auth for: %s %s", c.Request.Method, c.Request.URL.Path)

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			log.Printf("No Authorization header found")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			log.Printf("Invalid Authorization format: %s", authHeader)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization头格式必须是Bearer {token}"})
			c.Abort()
			return
		}

		tokenStr := parts[1]
		log.Printf("Found token: %s...", tokenStr[:min(10, len(tokenStr))])

		// 验证令牌
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			// 验证签名算法
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(secretKey), nil
		})

		if err != nil {
			log.Printf("Token validation error: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效或已过期的令牌"})
			c.Abort()
			return
		}

		if !token.Valid {
			log.Printf("Token is invalid")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的令牌"})
			c.Abort()
			return
		}

		// 检查主密码是否已设置在内存中
		if GetMasterPassword() == "" {
			log.Printf("主密码未设置在内存中，尝试从数据库重新加载用户凭据")
			// 这里可以实现从数据库重新加载主密码的逻辑
			// 或者引导用户重新登录
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Session expired, please login again", "code": "SESSION_EXPIRED"})
			c.Abort()
			return
		}

		log.Printf("Auth successful for: %s %s", c.Request.Method, c.Request.URL.Path)
		c.Next()
	}
}

// 辅助函数：获取较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
