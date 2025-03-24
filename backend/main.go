package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/007Secret/007Password/controllers"
	"github.com/007Secret/007Password/database"
	"github.com/007Secret/007Password/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	log.Println("启动007Password管理器服务...")

	// 创建数据目录
	dataFolder := database.GetDBFolder()
	log.Printf("创建数据目录: %s", dataFolder)
	os.MkdirAll(dataFolder, 0755)

	// 初始化数据库
	dbPath := filepath.Join(dataFolder, "passwordManager.db")
	_, err := os.Stat(dbPath)
	if os.IsNotExist(err) {
		// 新数据库，使用无加密初始化
		log.Printf("数据库文件不存在，使用无加密方式初始化")
		if err := database.InitDB(); err != nil {
			log.Fatalf("数据库初始化失败: %v", err)
		}
	} else {
		// 已存在的数据库，尝试无加密方式打开
		log.Printf("数据库文件已存在，尝试使用无加密方式打开")
		if err := database.InitDB(); err != nil {
			log.Printf("无加密方式打开失败，数据库可能已加密: %v", err)
			log.Printf("请通过登录API提供主密码")
		}
	}
	log.Printf("数据库初始化完成")

	// 创建Gin路由
	r := gin.Default()

	// 添加路由日志中间件
	r.Use(func(c *gin.Context) {
		log.Printf("[req-%s] %s %s", strings.TrimPrefix(c.Request.URL.Path, "/api"), c.Request.Method, c.Request.URL.Path)
		c.Next()
	})

	// 配置CORS
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Authorization", "Accept"}
	r.Use(cors.New(config))

	// 添加请求日志记录中间件
	r.Use(func(c *gin.Context) {
		// 记录请求的路径和方法
		path := c.Request.URL.Path
		// 提取路径的最后一部分，用作请求标识
		parts := strings.Split(path, "/")
		requestID := "req"
		if len(parts) > 0 {
			requestID = "req-" + parts[len(parts)-1]
		}

		// 记录请求信息
		log.Printf("[%s] %s %s", requestID, c.Request.Method, path)

		// 继续处理请求
		c.Next()

		// 请求完成后记录状态码
		//log.Printf("[%s] 完成 %s %s 状态码: %d", requestID, c.Request.Method, path, c.Writer.Status())
	})

	// 公开路由组
	public := r.Group("/api/auth")
	{
		public.POST("/login", controllers.Login)
		public.GET("/validate", middleware.AuthRequired(), controllers.ValidateToken)
		public.POST("/setup", controllers.SetMasterPassword)
		public.GET("/check-first-time", controllers.CheckFirstTimeSetup)
		// 修改主密码需要授权
		public.POST("/change-password", middleware.AuthRequired(), controllers.ChangeMasterPassword)
	}

	// 需要授权的API
	authorized := r.Group("/api")
	authorized.Use(middleware.AuthRequired())
	{
		// 密码管理API
		authorized.GET("/passwords", controllers.GetAllPasswords)
		authorized.GET("/passwords/:id", controllers.GetPasswordByID)
		authorized.POST("/passwords", controllers.CreatePassword)
		authorized.PUT("/passwords/:id", controllers.UpdatePassword)
		authorized.DELETE("/passwords/:id", controllers.DeletePassword)
		authorized.GET("/passwords/search", controllers.SearchPasswords)
	}

	// 启动服务
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("服务启动在 :%s 端口", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
