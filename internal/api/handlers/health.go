package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"services": map[string]string{
			"mongodb": "connected",
			"redis":   "connected",
		},
	})
}
