package middleware

import (
	"hive-admin-go/database"
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

		claims, err := utils.ParseToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, models.NewErrorResponse(nil, "invalid or expired token"))
			c.Abort()
			return
		}

		// 令牌撤销：登出后的凭证按 jti 拦截到其原过期时刻为止；查询出错按已失效处理
		if claims.JTI != "" {
			var revokedCount int64
			if err := database.DB.Model(&models.SysTokenBlacklist{}).
				Where("jti = ?", claims.JTI).
				Count(&revokedCount).Error; err != nil || revokedCount > 0 {
				c.JSON(http.StatusUnauthorized, models.NewErrorResponse(nil, "token has been invalidated"))
				c.Abort()
				return
			}
		}

		var user models.SysUser
		if err := database.DB.Where("user_id = ?", claims.UserID).First(&user).Error; err != nil {
			c.JSON(http.StatusUnauthorized, models.NewErrorResponse(nil, "user no longer exists"))
			c.Abort()
			return
		}

		if user.DelFlag == 1 || user.Status == 0 {
			c.JSON(http.StatusUnauthorized, models.NewErrorResponse(nil, "account has been disabled"))
			c.Abort()
			return
		}

		// 密码修改或重置后版本号递增，旧世代凭证在此统一失效
		if user.PwdVersion != claims.PwdVersion {
			c.JSON(http.StatusUnauthorized, models.NewErrorResponse(nil, "token has been invalidated"))
			c.Abort()
			return
		}

		c.Set("userId", claims.UserID)
		c.Set("token", tokenString)
		c.Next()
	}
}
