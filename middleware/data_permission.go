package middleware

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"hive-admin-go/datapermission"
	"hive-admin-go/models"
)

const dataPermissionContextKey = "dataPermission"

type DataPermissionResolver interface {
	Resolve(ctx context.Context, userID string) (datapermission.Permission, error)
}

func DataPermissionMiddleware(resolver DataPermissionResolver) gin.HandlerFunc {
	return func(c *gin.Context) {
		permission, err := resolver.Resolve(c.Request.Context(), c.GetString("userId"))
		if err != nil {
			if errors.Is(err, datapermission.ErrUserUnavailable) {
				c.JSON(http.StatusUnauthorized, models.NewErrorResponse(nil, "用户已失效，请重新登录"))
			} else {
				c.JSON(http.StatusInternalServerError, models.NewErrorResponse(nil, "数据权限解析失败"))
			}
			c.Abort()
			return
		}
		c.Set(dataPermissionContextKey, permission)
		c.Next()
	}
}

func GetDataPermission(c *gin.Context) (datapermission.Permission, bool) {
	value, exists := c.Get(dataPermissionContextKey)
	permission, ok := value.(datapermission.Permission)
	return permission, exists && ok
}
