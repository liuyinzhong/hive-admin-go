package middleware

import (
	"net/http"

	"hive-admin-go/models"

	"github.com/gin-gonic/gin"
)

type PermissionChecker interface {
	HasCode(userID, code string) bool
}

type PermissionGuard struct {
	checker PermissionChecker
}

func NewPermissionGuard(checker PermissionChecker) *PermissionGuard {
	return &PermissionGuard{checker: checker}
}

func (g *PermissionGuard) Require(code string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !g.checker.HasCode(c.GetString("userId"), code) {
			c.JSON(http.StatusForbidden, models.NewErrorResponse(nil, "无接口访问权限"))
			c.Abort()
			return
		}
		c.Next()
	}
}
