package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"hive-admin-go/database"
	"hive-admin-go/models"
)

// SystemUserMiddleware 将出诊排班管理限制为系统管理员用户。
// 当前项目尚无通用接口权限中间件，系统管理员由 sys_user.is_sys = 1 标识。
func SystemUserMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDValue, exists := c.Get("userId")
		userID, ok := userIDValue.(string)
		if !exists || !ok || userID == "" {
			c.JSON(http.StatusUnauthorized, models.NewErrorResponse(nil, "未登录或登录已失效"))
			c.Abort()
			return
		}

		var count int64
		if err := database.DB.Model(&models.SysUser{}).
			Where("user_id = ? AND is_sys = 1 AND status = 1 AND del_flag = 0", userID).
			Count(&count).Error; err != nil {
			c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, "管理员权限校验失败"))
			c.Abort()
			return
		}
		if count == 0 {
			c.JSON(http.StatusForbidden, models.NewErrorResponse(nil, "仅系统管理员可以管理出诊排班"))
			c.Abort()
			return
		}
		c.Next()
	}
}
