package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 这里应该实现实际的认证逻辑
		// 现在我们只是返回一个示例响应
		c.JSON(http.StatusUnauthorized, gin.H{"error": "认证失败"})
		c.Abort()
	}
}
