package services

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"hive-admin-go/database"
	"hive-admin-go/models"
	"hive-admin-go/utils"
)

type normalizedErpWarehouseZoneSave struct {
	ZoneName           string
	ZoneNameNormalized string
	ZoneType           string
	Remark             *string
	ExpectedRowVersion int
}

type normalizedErpWarehouseLocationSave struct {
	LocationName           string
	LocationNameNormalized string
	Remark                 *string
	ExpectedRowVersion     int
}

func (s *ErpWarehouseService) GetWarehouseZoneList(warehouseID string, req models.ErpWarehouseZoneListRequest) (*utils.PaginationResponse, error) {
	if err := validateErpWarehouseUUID(warehouseID, "仓库ID"); err != nil {
		return nil, err
	}
	if err := s.ensureWarehouseExists(database.DB, warehouseID); err != nil {
		return nil, err
	}

	page, pageSize := normalizeErpWarehousePage(req.Page, req.PageSize, 20, 100)
	query := database.DB.Model(&models.ErpWarehouseZone{}).Where("warehouse_id = ? AND del_flag = 0", warehouseID)
	var err error
	query, err = s.applyWarehouseZoneFilters(query, req.Keyword, req.ZoneType)
	if err != nil {
		return nil, err
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	order := utils.BuildOrderBy(req.Sorts, map[string]string{
		"zoneCode":   "zone_code",
		"zoneName":   "zone_name",
		"zoneType":   "zone_type",
		"createDate": "create_date",
		"updateDate": "update_date",
	})
	if order == "" {
		order = "zone_code asc"
	}

	var zones []models.ErpWarehouseZone
	if err := query.Order(order).Offset((page - 1) * pageSize).Limit(pageSize).Find(&zones).Error; err != nil {
		return nil, err
	}

	responses := erpWarehouseZonesToResponses(zones)
	if err := s.fillWarehouseZoneLocationCounts(database.DB, responses); err != nil {
		return nil, err
	}
	return &utils.PaginationResponse{Items: responses, Total: total}, nil
}

func (s *ErpWarehouseService) GetWarehouseZoneDetail(warehouseID, zoneID string) (*models.ErpWarehouseZoneResponse, error) {
	if err := validateErpWarehouseUUID(warehouseID, "仓库ID"); err != nil {
		return nil, err
	}
	if err := validateErpWarehouseUUID(zoneID, "库区ID"); err != nil {
		return nil, err
	}

	zone, err := s.getWarehouseZone(database.DB, warehouseID, zoneID, false)
	if err != nil {
		return nil, err
	}
	response := erpWarehouseZoneToResponse(*zone)
	responses := []models.ErpWarehouseZoneResponse{*response}
	if err := s.fillWarehouseZoneLocationCounts(database.DB, responses); err != nil {
		return nil, err
	}
	response.LocationCount = responses[0].LocationCount
	return response, nil
}

func (s *ErpWarehouseService) GetWarehouseZoneOptions(warehouseID string, req models.ErpWarehouseZoneOptionsRequest) ([]models.ErpWarehouseZoneOptionResponse, error) {
	if err := validateErpWarehouseUUID(warehouseID, "仓库ID"); err != nil {
		return nil, err
	}
	if err := s.ensureWarehouseExists(database.DB, warehouseID); err != nil {
		return nil, err
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	query := database.DB.Model(&models.ErpWarehouseZone{}).Where("warehouse_id = ? AND del_flag = 0", warehouseID)
	if strings.TrimSpace(req.Keyword) != "" {
		keyword := "%" + strings.ToLower(strings.TrimSpace(req.Keyword)) + "%"
		query = query.Where("LOWER(zone_code) LIKE ? OR LOWER(zone_name) LIKE ?", keyword, keyword)
	}

	var zones []models.ErpWarehouseZone
	if err := query.Order("zone_code asc").Limit(pageSize).Find(&zones).Error; err != nil {
		return nil, err
	}
	options := make([]models.ErpWarehouseZoneOptionResponse, 0, len(zones))
	for _, zone := range zones {
		options = append(options, models.ErpWarehouseZoneOptionResponse{
			ZoneID:      zone.ZoneID,
			WarehouseID: zone.WarehouseID,
			ZoneCode:    zone.ZoneCode,
			ZoneName:    zone.ZoneName,
			ZoneType:    zone.ZoneType,
		})
	}
	return options, nil
}

func (s *ErpWarehouseService) CreateWarehouseZone(warehouseID string, req models.SaveErpWarehouseZoneRequest, operatorID string) (*models.ErpWarehouseZoneResponse, error) {
	if err := validateErpWarehouseUUID(warehouseID, "仓库ID"); err != nil {
		return nil, err
	}
	normalized, err := s.normalizeZoneSaveRequest(req, false)
	if err != nil {
		return nil, err
	}

	var createdID string
	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := s.ensureWarehouseExists(tx, warehouseID); err != nil {
			return err
		}
		if err := s.ensureWarehouseZoneNameUnique(tx, warehouseID, normalized.ZoneNameNormalized, ""); err != nil {
			return err
		}
		code, err := s.nextWarehouseZoneCode(tx)
		if err != nil {
			return err
		}

		now := time.Now()
		createdID = utils.GenerateUUID()
		zone := models.ErpWarehouseZone{
			ZoneID:             createdID,
			WarehouseID:        warehouseID,
			ZoneCode:           code,
			ZoneName:           normalized.ZoneName,
			ZoneNameNormalized: normalized.ZoneNameNormalized,
			ZoneType:           normalized.ZoneType,
			Remark:             normalized.Remark,
			RowVersion:         1,
			CreatorID:          optionalErpWarehouseOperatorID(operatorID),
			UpdaterID:          optionalErpWarehouseOperatorID(operatorID),
			CreateDate:         &now,
			UpdateDate:         &now,
			DelFlag:            0,
		}
		return tx.Create(&zone).Error
	}); err != nil {
		return nil, err
	}

	return s.GetWarehouseZoneDetail(warehouseID, createdID)
}

func (s *ErpWarehouseService) UpdateWarehouseZone(warehouseID, zoneID string, req models.SaveErpWarehouseZoneRequest, operatorID string) (*models.ErpWarehouseZoneResponse, error) {
	if err := validateErpWarehouseUUID(warehouseID, "仓库ID"); err != nil {
		return nil, err
	}
	if err := validateErpWarehouseUUID(zoneID, "库区ID"); err != nil {
		return nil, err
	}
	normalized, err := s.normalizeZoneSaveRequest(req, true)
	if err != nil {
		return nil, err
	}

	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		current, err := s.getWarehouseZone(tx, warehouseID, zoneID, true)
		if err != nil {
			return err
		}
		if current.RowVersion != normalized.ExpectedRowVersion {
			return fmt.Errorf("%w: 库区已被其他人修改，请刷新后重试", ErrErpWarehouseConflict)
		}
		if err := s.ensureWarehouseZoneNameUnique(tx, warehouseID, normalized.ZoneNameNormalized, zoneID); err != nil {
			return err
		}

		now := time.Now()
		return tx.Model(&models.ErpWarehouseZone{}).Where("zone_id = ?", zoneID).Updates(map[string]interface{}{
			"zone_name":            normalized.ZoneName,
			"zone_name_normalized": normalized.ZoneNameNormalized,
			"zone_type":            normalized.ZoneType,
			"remark":               normalized.Remark,
			"row_version":          current.RowVersion + 1,
			"updater_id":           optionalErpWarehouseOperatorID(operatorID),
			"update_date":          now,
		}).Error
	}); err != nil {
		return nil, err
	}

	return s.GetWarehouseZoneDetail(warehouseID, zoneID)
}

func (s *ErpWarehouseService) DeleteWarehouseZone(warehouseID, zoneID string, req models.DeleteErpWarehouseZoneRequest, operatorID string) error {
	if err := validateErpWarehouseUUID(warehouseID, "仓库ID"); err != nil {
		return err
	}
	if err := validateErpWarehouseUUID(zoneID, "库区ID"); err != nil {
		return err
	}
	if req.ExpectedRowVersion <= 0 {
		return fmt.Errorf("%w: 缺少数据版本号", ErrErpWarehouseInvalidInput)
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		current, err := s.getWarehouseZone(tx, warehouseID, zoneID, true)
		if err != nil {
			return err
		}
		if current.RowVersion != req.ExpectedRowVersion {
			return fmt.Errorf("%w: 库区已被其他人修改，请刷新后重试", ErrErpWarehouseConflict)
		}
		now := time.Now()
		return tx.Model(&models.ErpWarehouseZone{}).Where("zone_id = ?", zoneID).Updates(map[string]interface{}{
			"del_flag":    1,
			"row_version": current.RowVersion + 1,
			"updater_id":  optionalErpWarehouseOperatorID(operatorID),
			"update_date": now,
		}).Error
	})
}

func (s *ErpWarehouseService) GetWarehouseLocationList(warehouseID, zoneID string, req models.ErpWarehouseLocationListRequest) (*utils.PaginationResponse, error) {
	if err := validateErpWarehouseUUID(warehouseID, "仓库ID"); err != nil {
		return nil, err
	}
	if err := validateErpWarehouseUUID(zoneID, "库区ID"); err != nil {
		return nil, err
	}
	if _, err := s.getWarehouseZone(database.DB, warehouseID, zoneID, false); err != nil {
		return nil, err
	}

	page, pageSize := normalizeErpWarehousePage(req.Page, req.PageSize, 20, 100)
	query := database.DB.Model(&models.ErpWarehouseLocation{}).Where("warehouse_id = ? AND zone_id = ? AND del_flag = 0", warehouseID, zoneID)
	if strings.TrimSpace(req.Keyword) != "" {
		keyword := "%" + strings.ToLower(strings.TrimSpace(req.Keyword)) + "%"
		query = query.Where("LOWER(location_code) LIKE ? OR LOWER(location_name) LIKE ?", keyword, keyword)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	order := utils.BuildOrderBy(req.Sorts, map[string]string{
		"locationCode": "location_code",
		"locationName": "location_name",
		"createDate":   "create_date",
		"updateDate":   "update_date",
	})
	if order == "" {
		order = "location_code asc"
	}

	var locations []models.ErpWarehouseLocation
	if err := query.Order(order).Offset((page - 1) * pageSize).Limit(pageSize).Find(&locations).Error; err != nil {
		return nil, err
	}

	return &utils.PaginationResponse{Items: erpWarehouseLocationsToResponses(locations), Total: total}, nil
}

func (s *ErpWarehouseService) GetWarehouseLocationDetail(warehouseID, zoneID, locationID string) (*models.ErpWarehouseLocationResponse, error) {
	if err := validateErpWarehouseUUID(warehouseID, "仓库ID"); err != nil {
		return nil, err
	}
	if err := validateErpWarehouseUUID(zoneID, "库区ID"); err != nil {
		return nil, err
	}
	if err := validateErpWarehouseUUID(locationID, "货位ID"); err != nil {
		return nil, err
	}
	location, err := s.getWarehouseLocation(database.DB, warehouseID, zoneID, locationID, false)
	if err != nil {
		return nil, err
	}
	return erpWarehouseLocationToResponse(*location), nil
}

func (s *ErpWarehouseService) GetWarehouseLocationOptions(warehouseID, zoneID string, req models.ErpWarehouseLocationOptionsRequest) ([]models.ErpWarehouseLocationOptionResponse, error) {
	if err := validateErpWarehouseUUID(warehouseID, "仓库ID"); err != nil {
		return nil, err
	}
	if err := validateErpWarehouseUUID(zoneID, "库区ID"); err != nil {
		return nil, err
	}
	if _, err := s.getWarehouseZone(database.DB, warehouseID, zoneID, false); err != nil {
		return nil, err
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	query := database.DB.Model(&models.ErpWarehouseLocation{}).Where("warehouse_id = ? AND zone_id = ? AND del_flag = 0", warehouseID, zoneID)
	if strings.TrimSpace(req.Keyword) != "" {
		keyword := "%" + strings.ToLower(strings.TrimSpace(req.Keyword)) + "%"
		query = query.Where("LOWER(location_code) LIKE ? OR LOWER(location_name) LIKE ?", keyword, keyword)
	}
	var locations []models.ErpWarehouseLocation
	if err := query.Order("location_code asc").Limit(pageSize).Find(&locations).Error; err != nil {
		return nil, err
	}

	options := make([]models.ErpWarehouseLocationOptionResponse, 0, len(locations))
	for _, location := range locations {
		options = append(options, models.ErpWarehouseLocationOptionResponse{
			LocationID:   location.LocationID,
			WarehouseID:  location.WarehouseID,
			ZoneID:       location.ZoneID,
			LocationCode: location.LocationCode,
			LocationName: location.LocationName,
		})
	}
	return options, nil
}

func (s *ErpWarehouseService) CreateWarehouseLocation(warehouseID, zoneID string, req models.SaveErpWarehouseLocationRequest, operatorID string) (*models.ErpWarehouseLocationResponse, error) {
	if err := validateErpWarehouseUUID(warehouseID, "仓库ID"); err != nil {
		return nil, err
	}
	if err := validateErpWarehouseUUID(zoneID, "库区ID"); err != nil {
		return nil, err
	}
	normalized, err := s.normalizeLocationSaveRequest(req, false)
	if err != nil {
		return nil, err
	}

	var createdID string
	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		if _, err := s.getWarehouseZone(tx, warehouseID, zoneID, false); err != nil {
			return err
		}
		if err := s.ensureWarehouseLocationNameUnique(tx, zoneID, normalized.LocationNameNormalized, ""); err != nil {
			return err
		}
		code, err := s.nextWarehouseLocationCode(tx)
		if err != nil {
			return err
		}

		now := time.Now()
		createdID = utils.GenerateUUID()
		location := models.ErpWarehouseLocation{
			LocationID:             createdID,
			WarehouseID:            warehouseID,
			ZoneID:                 zoneID,
			LocationCode:           code,
			LocationName:           normalized.LocationName,
			LocationNameNormalized: normalized.LocationNameNormalized,
			Remark:                 normalized.Remark,
			RowVersion:             1,
			CreatorID:              optionalErpWarehouseOperatorID(operatorID),
			UpdaterID:              optionalErpWarehouseOperatorID(operatorID),
			CreateDate:             &now,
			UpdateDate:             &now,
			DelFlag:                0,
		}
		return tx.Create(&location).Error
	}); err != nil {
		return nil, err
	}

	return s.GetWarehouseLocationDetail(warehouseID, zoneID, createdID)
}

func (s *ErpWarehouseService) UpdateWarehouseLocation(warehouseID, zoneID, locationID string, req models.SaveErpWarehouseLocationRequest, operatorID string) (*models.ErpWarehouseLocationResponse, error) {
	if err := validateErpWarehouseUUID(warehouseID, "仓库ID"); err != nil {
		return nil, err
	}
	if err := validateErpWarehouseUUID(zoneID, "库区ID"); err != nil {
		return nil, err
	}
	if err := validateErpWarehouseUUID(locationID, "货位ID"); err != nil {
		return nil, err
	}
	normalized, err := s.normalizeLocationSaveRequest(req, true)
	if err != nil {
		return nil, err
	}

	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		current, err := s.getWarehouseLocation(tx, warehouseID, zoneID, locationID, true)
		if err != nil {
			return err
		}
		if current.RowVersion != normalized.ExpectedRowVersion {
			return fmt.Errorf("%w: 货位已被其他人修改，请刷新后重试", ErrErpWarehouseConflict)
		}
		if err := s.ensureWarehouseLocationNameUnique(tx, zoneID, normalized.LocationNameNormalized, locationID); err != nil {
			return err
		}

		now := time.Now()
		return tx.Model(&models.ErpWarehouseLocation{}).Where("location_id = ?", locationID).Updates(map[string]interface{}{
			"location_name":            normalized.LocationName,
			"location_name_normalized": normalized.LocationNameNormalized,
			"remark":                   normalized.Remark,
			"row_version":              current.RowVersion + 1,
			"updater_id":               optionalErpWarehouseOperatorID(operatorID),
			"update_date":              now,
		}).Error
	}); err != nil {
		return nil, err
	}

	return s.GetWarehouseLocationDetail(warehouseID, zoneID, locationID)
}

func (s *ErpWarehouseService) DeleteWarehouseLocation(warehouseID, zoneID, locationID string, req models.DeleteErpWarehouseLocationRequest, operatorID string) error {
	if err := validateErpWarehouseUUID(warehouseID, "仓库ID"); err != nil {
		return err
	}
	if err := validateErpWarehouseUUID(zoneID, "库区ID"); err != nil {
		return err
	}
	if err := validateErpWarehouseUUID(locationID, "货位ID"); err != nil {
		return err
	}
	if req.ExpectedRowVersion <= 0 {
		return fmt.Errorf("%w: 缺少数据版本号", ErrErpWarehouseInvalidInput)
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		current, err := s.getWarehouseLocation(tx, warehouseID, zoneID, locationID, true)
		if err != nil {
			return err
		}
		if current.RowVersion != req.ExpectedRowVersion {
			return fmt.Errorf("%w: 货位已被其他人修改，请刷新后重试", ErrErpWarehouseConflict)
		}
		now := time.Now()
		return tx.Model(&models.ErpWarehouseLocation{}).Where("location_id = ?", locationID).Updates(map[string]interface{}{
			"del_flag":    1,
			"row_version": current.RowVersion + 1,
			"updater_id":  optionalErpWarehouseOperatorID(operatorID),
			"update_date": now,
		}).Error
	})
}

func (s *ErpWarehouseService) ensureWarehouseExists(tx *gorm.DB, warehouseID string) error {
	var count int64
	if err := tx.Model(&models.ErpWarehouse{}).Where("warehouse_id = ? AND del_flag = 0", warehouseID).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("%w: 仓库不存在", ErrErpWarehouseNotFound)
	}
	return nil
}

func (s *ErpWarehouseService) getWarehouseZone(tx *gorm.DB, warehouseID, zoneID string, forUpdate bool) (*models.ErpWarehouseZone, error) {
	query := tx.Where("zone_id = ? AND warehouse_id = ? AND del_flag = 0", zoneID, warehouseID)
	if forUpdate {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	var zone models.ErpWarehouseZone
	if err := query.First(&zone).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: 库区不存在", ErrErpWarehouseNotFound)
		}
		return nil, err
	}
	return &zone, nil
}

func (s *ErpWarehouseService) getWarehouseLocation(tx *gorm.DB, warehouseID, zoneID, locationID string, forUpdate bool) (*models.ErpWarehouseLocation, error) {
	query := tx.Where("location_id = ? AND warehouse_id = ? AND zone_id = ? AND del_flag = 0", locationID, warehouseID, zoneID)
	if forUpdate {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	var location models.ErpWarehouseLocation
	if err := query.First(&location).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: 货位不存在", ErrErpWarehouseNotFound)
		}
		return nil, err
	}
	return &location, nil
}

func (s *ErpWarehouseService) normalizeZoneSaveRequest(req models.SaveErpWarehouseZoneRequest, requireVersion bool) (*normalizedErpWarehouseZoneSave, error) {
	zoneName := strings.TrimSpace(req.ZoneName)
	if zoneName == "" {
		return nil, fmt.Errorf("%w: 库区名称不能为空", ErrErpWarehouseInvalidInput)
	}
	if len([]rune(zoneName)) > 128 {
		return nil, fmt.Errorf("%w: 库区名称不能超过128个字符", ErrErpWarehouseInvalidInput)
	}
	if requireVersion && req.ExpectedRowVersion <= 0 {
		return nil, fmt.Errorf("%w: 缺少数据版本号", ErrErpWarehouseInvalidInput)
	}
	zoneType, err := normalizeErpWarehouseZoneType(req.ZoneType)
	if err != nil {
		return nil, err
	}
	return &normalizedErpWarehouseZoneSave{
		ZoneName:           zoneName,
		ZoneNameNormalized: normalizeErpWarehouseText(zoneName),
		ZoneType:           zoneType,
		Remark:             normalizeErpWarehouseOptionalString(req.Remark),
		ExpectedRowVersion: req.ExpectedRowVersion,
	}, nil
}

func (s *ErpWarehouseService) normalizeLocationSaveRequest(req models.SaveErpWarehouseLocationRequest, requireVersion bool) (*normalizedErpWarehouseLocationSave, error) {
	locationName := strings.TrimSpace(req.LocationName)
	if locationName == "" {
		return nil, fmt.Errorf("%w: 货位名称不能为空", ErrErpWarehouseInvalidInput)
	}
	if len([]rune(locationName)) > 128 {
		return nil, fmt.Errorf("%w: 货位名称不能超过128个字符", ErrErpWarehouseInvalidInput)
	}
	if requireVersion && req.ExpectedRowVersion <= 0 {
		return nil, fmt.Errorf("%w: 缺少数据版本号", ErrErpWarehouseInvalidInput)
	}
	return &normalizedErpWarehouseLocationSave{
		LocationName:           locationName,
		LocationNameNormalized: normalizeErpWarehouseText(locationName),
		Remark:                 normalizeErpWarehouseOptionalString(req.Remark),
		ExpectedRowVersion:     req.ExpectedRowVersion,
	}, nil
}

func (s *ErpWarehouseService) applyWarehouseZoneFilters(query *gorm.DB, keyword, zoneType string) (*gorm.DB, error) {
	if strings.TrimSpace(keyword) != "" {
		kw := "%" + strings.ToLower(strings.TrimSpace(keyword)) + "%"
		query = query.Where("LOWER(zone_code) LIKE ? OR LOWER(zone_name) LIKE ?", kw, kw)
	}
	if strings.TrimSpace(zoneType) != "" {
		normalized, err := normalizeErpWarehouseZoneType(zoneType)
		if err != nil {
			return nil, err
		}
		query = query.Where("zone_type = ?", normalized)
	}
	return query, nil
}

func (s *ErpWarehouseService) ensureWarehouseZoneNameUnique(tx *gorm.DB, warehouseID, nameNormalized, excludeID string) error {
	query := tx.Model(&models.ErpWarehouseZone{}).Where("warehouse_id = ? AND del_flag = 0 AND zone_name_normalized = ?", warehouseID, nameNormalized)
	if excludeID != "" {
		query = query.Where("zone_id != ?", excludeID)
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("%w: 当前仓库下库区名称已存在", ErrErpWarehouseConflict)
	}
	return nil
}

func (s *ErpWarehouseService) ensureWarehouseLocationNameUnique(tx *gorm.DB, zoneID, nameNormalized, excludeID string) error {
	query := tx.Model(&models.ErpWarehouseLocation{}).Where("zone_id = ? AND del_flag = 0 AND location_name_normalized = ?", zoneID, nameNormalized)
	if excludeID != "" {
		query = query.Where("location_id != ?", excludeID)
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("%w: 当前库区下货位名称已存在", ErrErpWarehouseConflict)
	}
	return nil
}

func (s *ErpWarehouseService) nextWarehouseZoneCode(tx *gorm.DB) (string, error) {
	return NewBaseCodeSequenceService().NextBusinessCode(tx, "ERP_WAREHOUSE_ZONE", "WZ", 6)
}

func (s *ErpWarehouseService) nextWarehouseLocationCode(tx *gorm.DB) (string, error) {
	return NewBaseCodeSequenceService().NextBusinessCode(tx, "ERP_WAREHOUSE_LOCATION", "WL", 6)
}

func (s *ErpWarehouseService) fillWarehouseZoneCounts(tx *gorm.DB, warehouses []models.ErpWarehouseResponse) error {
	if len(warehouses) == 0 {
		return nil
	}
	ids := make([]string, 0, len(warehouses))
	for _, warehouse := range warehouses {
		ids = append(ids, warehouse.WarehouseID)
	}
	var rows []struct {
		WarehouseID string
		Count       int
	}
	if err := tx.Model(&models.ErpWarehouseZone{}).
		Select("warehouse_id, COUNT(*) AS count").
		Where("warehouse_id IN ? AND del_flag = 0", ids).
		Group("warehouse_id").
		Scan(&rows).Error; err != nil {
		return err
	}
	countMap := make(map[string]int, len(rows))
	for _, row := range rows {
		countMap[row.WarehouseID] = row.Count
	}
	for i := range warehouses {
		warehouses[i].ZoneCount = countMap[warehouses[i].WarehouseID]
	}
	return nil
}

func (s *ErpWarehouseService) fillWarehouseZoneLocationCounts(tx *gorm.DB, zones []models.ErpWarehouseZoneResponse) error {
	if len(zones) == 0 {
		return nil
	}
	ids := make([]string, 0, len(zones))
	for _, zone := range zones {
		ids = append(ids, zone.ZoneID)
	}
	var rows []struct {
		ZoneID string
		Count  int
	}
	if err := tx.Model(&models.ErpWarehouseLocation{}).
		Select("zone_id, COUNT(*) AS count").
		Where("zone_id IN ? AND del_flag = 0", ids).
		Group("zone_id").
		Scan(&rows).Error; err != nil {
		return err
	}
	countMap := make(map[string]int, len(rows))
	for _, row := range rows {
		countMap[row.ZoneID] = row.Count
	}
	for i := range zones {
		zones[i].LocationCount = countMap[zones[i].ZoneID]
	}
	return nil
}

func erpWarehouseZonesToResponses(zones []models.ErpWarehouseZone) []models.ErpWarehouseZoneResponse {
	responses := make([]models.ErpWarehouseZoneResponse, 0, len(zones))
	for _, zone := range zones {
		responses = append(responses, *erpWarehouseZoneToResponse(zone))
	}
	return responses
}

func erpWarehouseZoneToResponse(zone models.ErpWarehouseZone) *models.ErpWarehouseZoneResponse {
	return &models.ErpWarehouseZoneResponse{
		ZoneID:      zone.ZoneID,
		WarehouseID: zone.WarehouseID,
		ZoneCode:    zone.ZoneCode,
		ZoneName:    zone.ZoneName,
		ZoneType:    zone.ZoneType,
		Remark:      zone.Remark,
		RowVersion:  zone.RowVersion,
		CreateDate:  models.TimeToStringPtr(zone.CreateDate),
		UpdateDate:  models.TimeToStringPtr(zone.UpdateDate),
	}
}

func erpWarehouseLocationsToResponses(locations []models.ErpWarehouseLocation) []models.ErpWarehouseLocationResponse {
	responses := make([]models.ErpWarehouseLocationResponse, 0, len(locations))
	for _, location := range locations {
		responses = append(responses, *erpWarehouseLocationToResponse(location))
	}
	return responses
}

func erpWarehouseLocationToResponse(location models.ErpWarehouseLocation) *models.ErpWarehouseLocationResponse {
	return &models.ErpWarehouseLocationResponse{
		LocationID:   location.LocationID,
		WarehouseID:  location.WarehouseID,
		ZoneID:       location.ZoneID,
		LocationCode: location.LocationCode,
		LocationName: location.LocationName,
		Remark:       location.Remark,
		RowVersion:   location.RowVersion,
		CreateDate:   models.TimeToStringPtr(location.CreateDate),
		UpdateDate:   models.TimeToStringPtr(location.UpdateDate),
	}
}

func normalizeErpWarehouseZoneType(value string) (string, error) {
	normalized := strings.ToUpper(strings.TrimSpace(value))
	switch normalized {
	case models.WarehouseZoneTypeNormal,
		models.WarehouseZoneTypePendingInspection,
		models.WarehouseZoneTypeQualified,
		models.WarehouseZoneTypeUnqualified,
		models.WarehouseZoneTypeReturned:
		return normalized, nil
	default:
		return "", fmt.Errorf("%w: 库区类型不支持", ErrErpWarehouseInvalidInput)
	}
}
