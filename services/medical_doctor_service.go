package services

import (
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"gorm.io/gorm"

	"hive-admin-go/database"
	"hive-admin-go/models"
	"hive-admin-go/utils"
)

const (
	doctorGenderDictType     = "MED_DOCTOR_GENDER"
	doctorTitleDictType      = "MED_DOCTOR_TITLE"
	doctorEmploymentDictType = "MED_EMPLOYMENT_TYPE"
)

type MedicalDoctorService struct{}

func NewMedicalDoctorService() *MedicalDoctorService {
	return &MedicalDoctorService{}
}

func (s *MedicalDoctorService) GetDoctorList(req models.DoctorListRequest) (*utils.PageResult, error) {
	query := database.DB.Model(&models.MedDoctor{}).Where("med_doctor.del_flag = 0")
	keyword := strings.TrimSpace(req.Keyword)
	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("med_doctor.name LIKE ? OR med_doctor.doctor_no LIKE ? OR med_doctor.name_pinyin LIKE ?", like, like, like)
	}
	if req.DepartmentID != "" {
		if err := validateMedicalUUID(req.DepartmentID, "临床科室ID"); err != nil {
			return nil, err
		}
		query = query.Where("EXISTS (SELECT 1 FROM med_doctor_department mdd WHERE mdd.doctor_id = med_doctor.doctor_id AND mdd.department_id = ? AND mdd.status = 1 AND mdd.del_flag = 0)", req.DepartmentID)
	}
	if req.ProfessionalTitle != "" {
		query = query.Where("professional_title = ?", req.ProfessionalTitle)
	}
	if req.EmploymentType != "" {
		query = query.Where("employment_type = ?", req.EmploymentType)
	}
	if req.Status != nil {
		if err := validateMedicalStatus(*req.Status); err != nil {
			return nil, err
		}
		query = query.Where("status = ?", *req.Status)
	}

	order := utils.BuildOrderBy(req.Sorts, map[string]string{
		"doctorNo":   "doctor_no",
		"name":       "name",
		"sort":       "sort",
		"status":     "status",
		"createDate": "create_date",
	})
	if order == "" {
		order = "sort asc, create_date desc"
	}
	query = query.Order(order)

	pageSize := req.PageSize
	if pageSize > 100 {
		pageSize = 100
	}
	var doctors []models.MedDoctor
	pageResult, err := utils.Paginate(query, req.Page, pageSize, &doctors)
	if err != nil {
		return nil, err
	}
	responses, err := s.buildDoctorResponses(database.DB, doctors)
	if err != nil {
		return nil, err
	}
	pageResult.Items = responses
	return pageResult, nil
}

func (s *MedicalDoctorService) GetAllDoctors() ([]*models.DoctorOptionResponse, error) {
	var doctors []models.MedDoctor
	if err := database.DB.Where("del_flag = 0 AND status = 1").Order("sort asc, name asc").Find(&doctors).Error; err != nil {
		return nil, err
	}
	responses, err := s.buildDoctorResponses(database.DB, doctors)
	if err != nil {
		return nil, err
	}
	options := make([]*models.DoctorOptionResponse, 0, len(responses))
	for _, doctor := range responses {
		options = append(options, &models.DoctorOptionResponse{
			DoctorID:              doctor.DoctorID,
			DoctorNo:              doctor.DoctorNo,
			Name:                  doctor.Name,
			ProfessionalTitle:     doctor.ProfessionalTitle,
			PrimaryDepartmentID:   doctor.PrimaryDepartmentID,
			PrimaryDepartmentName: doctor.PrimaryDepartmentName,
		})
	}
	return options, nil
}

func (s *MedicalDoctorService) CreateDoctor(req models.SaveDoctorRequest, operatorID string) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		prepared, err := s.prepareDoctorRequest(tx, req, "")
		if err != nil {
			return err
		}
		now := time.Now()
		doctorID := utils.GenerateUUID()
		doctor := models.MedDoctor{
			DoctorID:               doctorID,
			DoctorNo:               strings.TrimSpace(req.DoctorNo),
			UserID:                 prepared.userID,
			Name:                   strings.TrimSpace(req.Name),
			NamePinyin:             normalizeMedicalOptionalString(req.NamePinyin),
			Gender:                 normalizeMedicalOptionalString(req.Gender),
			BirthDate:              prepared.birthDate,
			Phone:                  normalizeMedicalOptionalString(req.Phone),
			Email:                  normalizeMedicalOptionalString(req.Email),
			Avatar:                 normalizeMedicalOptionalString(req.Avatar),
			ProfessionalTitle:      req.ProfessionalTitle,
			AdministrativePosition: normalizeMedicalOptionalString(req.AdministrativePosition),
			EmploymentType:         req.EmploymentType,
			PracticeStartDate:      prepared.practiceStartDate,
			EmploymentDate:         prepared.employmentDate,
			DepartureDate:          prepared.departureDate,
			Expertise:              normalizeMedicalOptionalString(req.Expertise),
			Introduction:           normalizeMedicalOptionalString(req.Introduction),
			DefaultVisitMinutes:    prepared.defaultVisitMinutes,
			OnlineConsultation:     req.OnlineConsultation,
			AppointmentEnabled:     req.AppointmentEnabled,
			ProfileVisible:         req.ProfileVisible,
			Sort:                   req.Sort,
			Status:                 req.Status,
			Remark:                 normalizeMedicalOptionalString(req.Remark),
			CreatorID:              optionalOperatorID(operatorID),
			UpdaterID:              optionalOperatorID(operatorID),
			CreateDate:             &now,
			UpdateDate:             &now,
			DelFlag:                0,
		}
		if err := tx.Create(&doctor).Error; err != nil {
			return err
		}
		return s.replaceDoctorDepartments(tx, doctorID, req.DepartmentIDs, req.PrimaryDepartmentID, operatorID)
	})
}

func (s *MedicalDoctorService) GetDoctorDetail(doctorID string) (*models.DoctorResponse, error) {
	if err := validateMedicalUUID(doctorID, "医生ID"); err != nil {
		return nil, err
	}
	var doctor models.MedDoctor
	if err := database.DB.Where("doctor_id = ? AND del_flag = 0", doctorID).First(&doctor).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: 医生不存在", ErrMedicalNotFound)
		}
		return nil, err
	}
	responses, err := s.buildDoctorResponses(database.DB, []models.MedDoctor{doctor})
	if err != nil {
		return nil, err
	}
	return responses[0], nil
}

func (s *MedicalDoctorService) UpdateDoctor(doctorID string, req models.SaveDoctorRequest, operatorID string) error {
	if err := validateMedicalUUID(doctorID, "医生ID"); err != nil {
		return err
	}
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var doctor models.MedDoctor
		if err := tx.Where("doctor_id = ? AND del_flag = 0", doctorID).First(&doctor).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("%w: 医生不存在", ErrMedicalNotFound)
			}
			return err
		}
		prepared, err := s.prepareDoctorRequest(tx, req, doctorID)
		if err != nil {
			return err
		}
		updates := map[string]interface{}{
			"doctor_no":               strings.TrimSpace(req.DoctorNo),
			"user_id":                 prepared.userID,
			"name":                    strings.TrimSpace(req.Name),
			"name_pinyin":             normalizeMedicalOptionalString(req.NamePinyin),
			"gender":                  normalizeMedicalOptionalString(req.Gender),
			"birth_date":              prepared.birthDate,
			"phone":                   normalizeMedicalOptionalString(req.Phone),
			"email":                   normalizeMedicalOptionalString(req.Email),
			"avatar":                  normalizeMedicalOptionalString(req.Avatar),
			"professional_title":      req.ProfessionalTitle,
			"administrative_position": normalizeMedicalOptionalString(req.AdministrativePosition),
			"employment_type":         req.EmploymentType,
			"practice_start_date":     prepared.practiceStartDate,
			"employment_date":         prepared.employmentDate,
			"departure_date":          prepared.departureDate,
			"expertise":               normalizeMedicalOptionalString(req.Expertise),
			"introduction":            normalizeMedicalOptionalString(req.Introduction),
			"default_visit_minutes":   prepared.defaultVisitMinutes,
			"online_consultation":     req.OnlineConsultation,
			"appointment_enabled":     req.AppointmentEnabled,
			"profile_visible":         req.ProfileVisible,
			"sort":                    req.Sort,
			"status":                  req.Status,
			"remark":                  normalizeMedicalOptionalString(req.Remark),
			"updater_id":              optionalOperatorID(operatorID),
			"update_date":             time.Now(),
		}
		if err := tx.Model(&models.MedDoctor{}).Where("doctor_id = ?", doctorID).Updates(updates).Error; err != nil {
			return err
		}
		return s.replaceDoctorDepartments(tx, doctorID, req.DepartmentIDs, req.PrimaryDepartmentID, operatorID)
	})
}

func (s *MedicalDoctorService) UpdateDoctorStatus(doctorID string, status int, operatorID string) error {
	if err := validateMedicalUUID(doctorID, "医生ID"); err != nil {
		return err
	}
	if err := validateMedicalStatus(status); err != nil {
		return err
	}
	result := database.DB.Model(&models.MedDoctor{}).Where("doctor_id = ? AND del_flag = 0", doctorID).Updates(map[string]interface{}{
		"status":      status,
		"updater_id":  optionalOperatorID(operatorID),
		"update_date": time.Now(),
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("%w: 医生不存在", ErrMedicalNotFound)
	}
	return nil
}

func (s *MedicalDoctorService) DeleteDoctors(doctorIDs []string, operatorID string) error {
	if len(doctorIDs) == 0 {
		return fmt.Errorf("%w: 请选择要删除的医生", ErrMedicalInvalidInput)
	}
	return database.DB.Transaction(func(tx *gorm.DB) error {
		for _, doctorID := range doctorIDs {
			if err := validateMedicalUUID(doctorID, "医生ID"); err != nil {
				return err
			}
			var count int64
			if err := tx.Model(&models.MedDoctor{}).Where("doctor_id = ? AND del_flag = 0", doctorID).Count(&count).Error; err != nil {
				return err
			}
			if count == 0 {
				return fmt.Errorf("%w: 医生不存在", ErrMedicalNotFound)
			}
			now := time.Now()
			if err := tx.Model(&models.MedDoctorDepartment{}).Where("doctor_id = ? AND del_flag = 0", doctorID).Updates(map[string]interface{}{
				"status":      0,
				"del_flag":    1,
				"updater_id":  optionalOperatorID(operatorID),
				"update_date": now,
			}).Error; err != nil {
				return err
			}
			if err := tx.Model(&models.MedDoctor{}).Where("doctor_id = ?", doctorID).Updates(map[string]interface{}{
				"status":      0,
				"del_flag":    1,
				"updater_id":  optionalOperatorID(operatorID),
				"update_date": now,
			}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

type preparedDoctorRequest struct {
	userID              *string
	birthDate           *time.Time
	practiceStartDate   *time.Time
	employmentDate      *time.Time
	departureDate       *time.Time
	defaultVisitMinutes int
}

func (s *MedicalDoctorService) prepareDoctorRequest(tx *gorm.DB, req models.SaveDoctorRequest, excludeDoctorID string) (*preparedDoctorRequest, error) {
	if strings.TrimSpace(req.DoctorNo) == "" || strings.TrimSpace(req.Name) == "" {
		return nil, fmt.Errorf("%w: 医生编号和姓名不能为空", ErrMedicalInvalidInput)
	}
	if err := validateMedicalStatus(req.Status); err != nil {
		return nil, err
	}
	for label, value := range map[string]int{
		"是否支持线上问诊": req.OnlineConsultation,
		"是否允许预约":   req.AppointmentEnabled,
		"是否公开展示":   req.ProfileVisible,
	} {
		if value != 0 && value != 1 {
			return nil, fmt.Errorf("%w: %s只能是0或1", ErrMedicalInvalidInput, label)
		}
	}
	defaultVisitMinutes := req.DefaultVisitMinutes
	if defaultVisitMinutes == 0 {
		defaultVisitMinutes = 15
	}
	if defaultVisitMinutes < 5 || defaultVisitMinutes > 240 {
		return nil, fmt.Errorf("%w: 默认接诊时长必须在5到240分钟之间", ErrMedicalInvalidInput)
	}
	if req.Phone != nil && len(strings.TrimSpace(*req.Phone)) > 20 {
		return nil, fmt.Errorf("%w: 联系电话不能超过20个字符", ErrMedicalInvalidInput)
	}
	if email := normalizeMedicalOptionalString(req.Email); email != nil {
		if _, err := mail.ParseAddress(*email); err != nil || len(*email) > 128 {
			return nil, fmt.Errorf("%w: 工作邮箱格式错误", ErrMedicalInvalidInput)
		}
	}

	query := tx.Model(&models.MedDoctor{}).Where("doctor_no = ? AND del_flag = 0", strings.TrimSpace(req.DoctorNo))
	if excludeDoctorID != "" {
		query = query.Where("doctor_id != ?", excludeDoctorID)
	}
	var duplicateCount int64
	if err := query.Count(&duplicateCount).Error; err != nil {
		return nil, err
	}
	if duplicateCount > 0 {
		return nil, fmt.Errorf("%w: 医生编号已存在", ErrMedicalConflict)
	}

	userID := normalizeMedicalOptionalString(req.UserID)
	if userID != nil {
		if err := validateMedicalUUID(*userID, "系统用户ID"); err != nil {
			return nil, err
		}
		var userCount int64
		if err := tx.Model(&models.SysUser{}).Where("user_id = ? AND status = 1 AND del_flag = 0", *userID).Count(&userCount).Error; err != nil {
			return nil, err
		}
		if userCount == 0 {
			return nil, fmt.Errorf("%w: 绑定用户不存在或已停用", ErrMedicalNotFound)
		}
		boundQuery := tx.Model(&models.MedDoctor{}).Where("user_id = ? AND del_flag = 0", *userID)
		if excludeDoctorID != "" {
			boundQuery = boundQuery.Where("doctor_id != ?", excludeDoctorID)
		}
		if err := boundQuery.Count(&duplicateCount).Error; err != nil {
			return nil, err
		}
		if duplicateCount > 0 {
			return nil, fmt.Errorf("%w: 该系统用户已绑定其他医生", ErrMedicalConflict)
		}
	}

	if err := validateDoctorDepartmentSelection(req.DepartmentIDs, req.PrimaryDepartmentID); err != nil {
		return nil, err
	}
	for _, departmentID := range req.DepartmentIDs {
		if err := validateMedicalUUID(departmentID, "临床科室ID"); err != nil {
			return nil, err
		}
	}
	var departmentCount int64
	if err := tx.Model(&models.MedDepartment{}).
		Where("department_id IN ? AND status = 1 AND del_flag = 0", req.DepartmentIDs).
		Count(&departmentCount).Error; err != nil {
		return nil, err
	}
	if departmentCount != int64(len(req.DepartmentIDs)) {
		return nil, fmt.Errorf("%w: 出诊科室不存在或已停用", ErrMedicalNotFound)
	}

	if err := s.validateDictValue(tx, doctorTitleDictType, req.ProfessionalTitle, true); err != nil {
		return nil, err
	}
	if err := s.validateDictValue(tx, doctorEmploymentDictType, req.EmploymentType, true); err != nil {
		return nil, err
	}
	if req.Gender != nil {
		if err := s.validateDictValue(tx, doctorGenderDictType, *req.Gender, false); err != nil {
			return nil, err
		}
	}

	birthDate, err := parseMedicalDate(req.BirthDate, "出生日期")
	if err != nil {
		return nil, err
	}
	practiceStartDate, err := parseMedicalDate(req.PracticeStartDate, "开始从业日期")
	if err != nil {
		return nil, err
	}
	employmentDate, err := parseMedicalDate(req.EmploymentDate, "入职日期")
	if err != nil {
		return nil, err
	}
	departureDate, err := parseMedicalDate(req.DepartureDate, "离职日期")
	if err != nil {
		return nil, err
	}
	if employmentDate != nil && departureDate != nil && departureDate.Before(*employmentDate) {
		return nil, fmt.Errorf("%w: 离职日期不能早于入职日期", ErrMedicalInvalidInput)
	}

	return &preparedDoctorRequest{
		userID:              userID,
		birthDate:           birthDate,
		practiceStartDate:   practiceStartDate,
		employmentDate:      employmentDate,
		departureDate:       departureDate,
		defaultVisitMinutes: defaultVisitMinutes,
	}, nil
}

func (s *MedicalDoctorService) validateDictValue(tx *gorm.DB, dictType, value string, required bool) error {
	value = strings.TrimSpace(value)
	if value == "" {
		if required {
			return fmt.Errorf("%w: 字典值不能为空", ErrMedicalInvalidInput)
		}
		return nil
	}
	var count int64
	if err := tx.Model(&models.SysDict{}).Where("type = ? AND value = ? AND status = 1 AND del_flag = 0", dictType, value).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("%w: 字典%s中不存在值%s", ErrMedicalInvalidInput, dictType, value)
	}
	return nil
}

func (s *MedicalDoctorService) replaceDoctorDepartments(tx *gorm.DB, doctorID string, departmentIDs []string, primaryDepartmentID, operatorID string) error {
	var existing []models.MedDoctorDepartment
	if err := tx.Where("doctor_id = ?", doctorID).Find(&existing).Error; err != nil {
		return err
	}
	existingByDepartment := make(map[string]models.MedDoctorDepartment, len(existing))
	for _, relation := range existing {
		existingByDepartment[relation.DepartmentID] = relation
	}
	selected := make(map[string]struct{}, len(departmentIDs))
	now := time.Now()
	for index, departmentID := range departmentIDs {
		selected[departmentID] = struct{}{}
		isPrimary := 0
		if departmentID == primaryDepartmentID {
			isPrimary = 1
		}
		if relation, exists := existingByDepartment[departmentID]; exists {
			if err := tx.Model(&models.MedDoctorDepartment{}).Where("doctor_department_id = ?", relation.DoctorDepartmentID).Updates(map[string]interface{}{
				"is_primary":          isPrimary,
				"appointment_enabled": 1,
				"sort":                index,
				"status":              1,
				"del_flag":            0,
				"updater_id":          optionalOperatorID(operatorID),
				"update_date":         now,
			}).Error; err != nil {
				return err
			}
			continue
		}
		relation := models.MedDoctorDepartment{
			DoctorDepartmentID: utils.GenerateUUID(),
			DoctorID:           doctorID,
			DepartmentID:       departmentID,
			IsPrimary:          isPrimary,
			AppointmentEnabled: 1,
			Sort:               index,
			Status:             1,
			CreatorID:          optionalOperatorID(operatorID),
			UpdaterID:          optionalOperatorID(operatorID),
			CreateDate:         &now,
			UpdateDate:         &now,
			DelFlag:            0,
		}
		if err := tx.Create(&relation).Error; err != nil {
			return err
		}
	}
	for _, relation := range existing {
		if _, exists := selected[relation.DepartmentID]; exists {
			continue
		}
		if err := tx.Model(&models.MedDoctorDepartment{}).Where("doctor_department_id = ?", relation.DoctorDepartmentID).Updates(map[string]interface{}{
			"is_primary":  0,
			"status":      0,
			"updater_id":  optionalOperatorID(operatorID),
			"update_date": now,
		}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *MedicalDoctorService) buildDoctorResponses(tx *gorm.DB, doctors []models.MedDoctor) ([]*models.DoctorResponse, error) {
	responses := make([]*models.DoctorResponse, 0, len(doctors))
	if len(doctors) == 0 {
		return responses, nil
	}
	doctorIDs := make([]string, 0, len(doctors))
	userIDs := make([]string, 0)
	for _, doctor := range doctors {
		doctorIDs = append(doctorIDs, doctor.DoctorID)
		if doctor.UserID != nil {
			userIDs = append(userIDs, *doctor.UserID)
		}
	}

	var relations []models.MedDoctorDepartment
	if err := tx.Where("doctor_id IN ? AND status = 1 AND del_flag = 0", doctorIDs).
		Order("is_primary desc, sort asc, create_date asc").Find(&relations).Error; err != nil {
		return nil, err
	}
	departmentIDs := make([]string, 0, len(relations))
	for _, relation := range relations {
		departmentIDs = append(departmentIDs, relation.DepartmentID)
	}
	var departments []models.MedDepartment
	if len(departmentIDs) > 0 {
		if err := tx.Where("department_id IN ?", departmentIDs).Find(&departments).Error; err != nil {
			return nil, err
		}
	}
	departmentByID := make(map[string]models.MedDepartment, len(departments))
	for _, department := range departments {
		departmentByID[department.DepartmentID] = department
	}
	relationsByDoctor := make(map[string][]models.MedDoctorDepartment)
	for _, relation := range relations {
		relationsByDoctor[relation.DoctorID] = append(relationsByDoctor[relation.DoctorID], relation)
	}

	userNameByID := make(map[string]*string)
	if len(userIDs) > 0 {
		var users []models.SysUser
		if err := tx.Select("user_id", "real_name").Where("user_id IN ? AND del_flag = 0", userIDs).Find(&users).Error; err != nil {
			return nil, err
		}
		for _, user := range users {
			userNameByID[user.UserID] = user.RealName
		}
	}

	for _, doctor := range doctors {
		response := doctorToResponse(doctor)
		if doctor.UserID != nil {
			response.UserName = userNameByID[*doctor.UserID]
		}
		for _, relation := range relationsByDoctor[doctor.DoctorID] {
			department, exists := departmentByID[relation.DepartmentID]
			if !exists {
				continue
			}
			response.DepartmentIDs = append(response.DepartmentIDs, department.DepartmentID)
			response.DepartmentNames = append(response.DepartmentNames, department.DepartmentName)
			response.Departments = append(response.Departments, models.DoctorDepartmentResponse{
				DoctorDepartmentID: relation.DoctorDepartmentID,
				DepartmentID:       department.DepartmentID,
				DepartmentCode:     department.DepartmentCode,
				DepartmentName:     department.DepartmentName,
				IsPrimary:          relation.IsPrimary,
				DepartmentPosition: relation.DepartmentPosition,
				AppointmentEnabled: relation.AppointmentEnabled,
				Status:             relation.Status,
			})
			if relation.IsPrimary == 1 {
				id := department.DepartmentID
				name := department.DepartmentName
				response.PrimaryDepartmentID = &id
				response.PrimaryDepartmentName = &name
			}
		}
		responses = append(responses, response)
	}
	return responses, nil
}

func doctorToResponse(doctor models.MedDoctor) *models.DoctorResponse {
	return &models.DoctorResponse{
		DoctorID:               doctor.DoctorID,
		DoctorNo:               doctor.DoctorNo,
		UserID:                 doctor.UserID,
		Name:                   doctor.Name,
		NamePinyin:             doctor.NamePinyin,
		Gender:                 doctor.Gender,
		BirthDate:              timeToDateStringPtr(doctor.BirthDate),
		Phone:                  doctor.Phone,
		Email:                  doctor.Email,
		Avatar:                 doctor.Avatar,
		ProfessionalTitle:      doctor.ProfessionalTitle,
		AdministrativePosition: doctor.AdministrativePosition,
		EmploymentType:         doctor.EmploymentType,
		PracticeStartDate:      timeToDateStringPtr(doctor.PracticeStartDate),
		EmploymentDate:         timeToDateStringPtr(doctor.EmploymentDate),
		DepartureDate:          timeToDateStringPtr(doctor.DepartureDate),
		Expertise:              doctor.Expertise,
		Introduction:           doctor.Introduction,
		DefaultVisitMinutes:    doctor.DefaultVisitMinutes,
		OnlineConsultation:     doctor.OnlineConsultation,
		AppointmentEnabled:     doctor.AppointmentEnabled,
		ProfileVisible:         doctor.ProfileVisible,
		Sort:                   doctor.Sort,
		Status:                 doctor.Status,
		Remark:                 doctor.Remark,
		DepartmentIDs:          make([]string, 0),
		DepartmentNames:        make([]string, 0),
		Departments:            make([]models.DoctorDepartmentResponse, 0),
		CreateDate:             models.TimeToStringPtr(doctor.CreateDate),
		UpdateDate:             models.TimeToStringPtr(doctor.UpdateDate),
	}
}

func validateDoctorDepartmentSelection(departmentIDs []string, primaryDepartmentID string) error {
	if len(departmentIDs) == 0 {
		return fmt.Errorf("%w: 医生至少需要一个出诊科室", ErrMedicalInvalidInput)
	}
	seen := make(map[string]struct{}, len(departmentIDs))
	primaryFound := false
	for _, departmentID := range departmentIDs {
		if _, exists := seen[departmentID]; exists {
			return fmt.Errorf("%w: 出诊科室不能重复", ErrMedicalInvalidInput)
		}
		seen[departmentID] = struct{}{}
		if departmentID == primaryDepartmentID {
			primaryFound = true
		}
	}
	if !primaryFound {
		return fmt.Errorf("%w: 主科室必须包含在出诊科室中", ErrMedicalInvalidInput)
	}
	return nil
}

func parseMedicalDate(value *string, label string) (*time.Time, error) {
	value = normalizeMedicalOptionalString(value)
	if value == nil {
		return nil, nil
	}
	parsed, err := time.ParseInLocation("2006-01-02", *value, time.Local)
	if err != nil {
		return nil, fmt.Errorf("%w: %s格式必须为YYYY-MM-DD", ErrMedicalInvalidInput, label)
	}
	return &parsed, nil
}

func timeToDateStringPtr(value *time.Time) *string {
	if value == nil {
		return nil
	}
	formatted := value.Format("2006-01-02")
	return &formatted
}
