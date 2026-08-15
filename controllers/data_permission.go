package controllers

import (
	"net/http"

	"hive-admin-go/datapermission"
	"hive-admin-go/middleware"
	"hive-admin-go/models"

	"github.com/gin-gonic/gin"
)

func currentDataPermission(c *gin.Context) datapermission.Permission {
	permission, ok := middleware.GetDataPermission(c)
	if ok {
		return permission
	}
	return datapermission.Permission{UserID: c.GetString("userId")}
}

func requireAllDataPermission(c *gin.Context) bool {
	if currentDataPermission(c).All {
		return true
	}
	c.JSON(http.StatusForbidden, models.NewErrorResponse(nil, "该操作需要全部数据权限"))
	return false
}
