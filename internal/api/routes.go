package api

import (
	"qinniu/internal/api/handlers"
	"qinniu/internal/api/middleware"

	"github.com/gin-gonic/gin"
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
			authorized.POST("/run-crawler", handlers.RunCrawlerHandler)
		}

		api.GET("/nations", handlers.GetAllNations)
	}
}
