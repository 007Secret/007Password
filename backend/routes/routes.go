package routes

import (
	"github.com/007Secret/007Password/controllers"
	"github.com/007Secret/007Password/middleware"
	"github.com/gin-gonic/gin"
)

// SetupRoutes 配置API路由
func SetupRoutes(r *gin.Engine) {
	// 认证API
	authGroup := r.Group("/api/auth")
	{
		authGroup.POST("/login", controllers.Login)
		authGroup.GET("/validate", middleware.AuthRequired(), controllers.ValidateToken)
		authGroup.POST("/setup", controllers.SetMasterPassword)
		authGroup.GET("/check-first-time", controllers.CheckFirstTimeSetup)
		authGroup.POST("/change-password", middleware.AuthRequired(), controllers.ChangeMasterPassword)
	}

	// 密码管理API
	passwordGroup := r.Group("/api/passwords", middleware.AuthRequired())
	{
		passwordGroup.GET("", controllers.GetAllPasswords)
		passwordGroup.GET("/:id", controllers.GetPasswordByID)
		passwordGroup.POST("", controllers.CreatePassword)
		passwordGroup.PUT("/:id", controllers.UpdatePassword)
		passwordGroup.DELETE("/:id", controllers.DeletePassword)
		passwordGroup.GET("/search", controllers.SearchPasswords)
	}
}
