package api

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"qinniu/internal/api/handlers"
	"qinniu/internal/api/middleware"
	"time"
)

func SetupRoutes(r *gin.Engine) {
	// 添加健康检查路由
	r.GET("/health", handlers.HealthCheck)

	api := r.Group("/api")
	{
		api.GET("/developers/:id", handlers.GetDeveloper)
		api.GET("/search", handlers.SearchDevelopers)

		// 需要认证的路由
		authorized := api.Group("/")
		authorized.Use(middleware.AuthRequired())
		{
			authorized.POST("/developers", handlers.CreateDeveloper)
			authorized.PUT("/developers/:id", handlers.UpdateDeveloper)
			authorized.DELETE("/developers/:id", handlers.DeleteDeveloper)
		}

		api.GET("/nations", handlers.GetAllNations)
	}
}

func CrossOriginMiddleware() gin.HandlerFunc {

	return cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"PUT", "PATCH", "POST", "GET", "OPTIONS", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Content-Length", "Token", "Authorization", "Access-Control-Allow-Headers"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			return origin == "https://github.com"
		},
		MaxAge: 12 * time.Hour,
	})
}
