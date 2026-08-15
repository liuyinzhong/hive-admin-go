package services

import (
	"context"

	"hive-admin-go/database"
	"hive-admin-go/datapermission"
)

func resolveDataPermission(userID string) (datapermission.Permission, error) {
	resolver := datapermission.NewResolver(datapermission.NewGormAssignmentStore(database.DB))
	return resolver.Resolve(context.Background(), userID)
}
