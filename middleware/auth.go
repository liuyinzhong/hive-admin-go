package middleware

import (
	"hive-admin-go/models"
	"hive-admin-go/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, models.NewErrorResponse(nil, "authorization header is required"))
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, models.NewErrorResponse(nil, "invalid authorization header format"))
			c.Abort()
			return
		}

		tokenString := parts[1]

		if utils.IsTokenBlacklisted(tokenString) {
			c.JSON(http.StatusUnauthorized, models.NewErrorResponse(nil, "token has been invalidated"))
			c.Abort()
			return
		}

		claims, err := utils.ParseToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, models.NewErrorResponse(nil, "invalid or expired token"))
			c.Abort()
			return
		}

		c.Set("userId", claims.UserID)
		c.Set("token", tokenString)
		c.Next()
	}
}
