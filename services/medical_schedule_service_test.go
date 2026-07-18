package services

import (
	"errors"
	"testing"
	"time"

	"hive-admin-go/models"
)

const scheduleTestUUID = "00000000-0000-0000-0000-000000000001"

func TestScheduleTimeOverlapRules(t *testing.T) {
	tests := []struct {
		name       string
		leftStart  string
		leftEnd    string
		rightStart string
		rightEnd   string
		want       bool
	}{
		{
			name:       "adjacent periods are allowed",
			leftStart:  "08:00:00",
			leftEnd:    "09:00:00",
			rightStart: "09:00:00",
			rightEnd:   "10:00:00",
		},
		{
			name:       "intersecting periods conflict",
			leftStart:  "08:00:00",
			leftEnd:    "09:30:00",
			rightStart: "09:00:00",
			rightEnd:   "10:00:00",
			want:       true,
		},
		{
			name:       "same period conflicts",
			leftStart:  "08:00:00",
			leftEnd:    "09:00:00",
			rightStart: "08:00:00",
			rightEnd:   "09:00:00",
			want:       true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := scheduleTimesOverlap(test.leftStart, test.leftEnd, test.rightStart, test.rightEnd); got != test.want {
				t.Fatalf("scheduleTimesOverlap() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestScheduleTemplateConflictRequiresWeekdayDateAndTimeOverlap(t *testing.T) {
	effectiveDate := time.Date(2026, 7, 1, 0, 0, 0, 0, medicalBusinessLocation)
	expiryDate := time.Date(2026, 7, 31, 0, 0, 0, 0, medicalBusinessLocation)
	prepared := &preparedScheduleTemplate{
		weekdays:      []int{1, 3, 5},
		startTime:     "08:00:00",
		endTime:       "12:00:00",
		effectiveDate: effectiveDate,
		expiryDate:    &expiryDate,
	}

	tests := []struct {
		name      string
		candidate models.MedScheduleTemplate
		want      bool
	}{
		{
			name: "all three dimensions overlap",
			candidate: models.MedScheduleTemplate{
				Weekdays:      []int{3, 6},
				StartTime:     "11:30:00",
				EndTime:       "13:00:00",
				EffectiveDate: effectiveDate,
				ExpiryDate:    &expiryDate,
			},
			want: true,
		},
		{
			name: "disjoint weekdays are allowed",
			candidate: models.MedScheduleTemplate{
				Weekdays:      []int{2, 4},
				StartTime:     "11:30:00",
				EndTime:       "13:00:00",
				EffectiveDate: effectiveDate,
				ExpiryDate:    &expiryDate,
			},
		},
		{
			name: "adjacent time is allowed",
			candidate: models.MedScheduleTemplate{
				Weekdays:      []int{1},
				StartTime:     "12:00:00",
				EndTime:       "13:00:00",
				EffectiveDate: effectiveDate,
				ExpiryDate:    &expiryDate,
			},
		},
		{
			name: "disjoint effective periods are allowed",
			candidate: models.MedScheduleTemplate{
				Weekdays:      []int{1},
				StartTime:     "11:30:00",
				EndTime:       "13:00:00",
				EffectiveDate: expiryDate.AddDate(0, 0, 1),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := scheduleTemplateConflicts(test.candidate, prepared); got != test.want {
				t.Fatalf("scheduleTemplateConflicts() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestBuildGeneratedSchedulesUsesEverySelectedWeekday(t *testing.T) {
	startDate := time.Date(2026, 7, 20, 0, 0, 0, 0, medicalBusinessLocation)
	template := models.MedScheduleTemplate{
		TemplateID:       scheduleTestUUID,
		DoctorID:         scheduleTestUUID,
		DepartmentID:     scheduleTestUUID,
		RegistrationType: "1",
		Weekdays:         []int{1, 3, 5},
		StartTime:        "08:00:00",
		EndTime:          "09:00:00",
		DefaultSlotQuota: 1,
		EffectiveDate:    startDate,
		Status:           1,
	}
	relations := map[string]models.MedDoctorDepartment{
		scheduleDimensionKey(template.DoctorID, template.DepartmentID, template.RegistrationType): {},
	}

	schedules, _, skipped, err := buildGeneratedSchedules(
		[]models.MedScheduleTemplate{template},
		relations,
		nil,
		"batch-id",
		startDate,
		startDate.AddDate(0, 0, 6),
		"",
	)
	if err != nil {
		t.Fatal(err)
	}
	if skipped != 0 {
		t.Fatalf("skipped count = %d, want 0", skipped)
	}
	if len(schedules) != 3 {
		t.Fatalf("schedule count = %d, want 3", len(schedules))
	}
	wantDates := []string{"2026-07-20", "2026-07-22", "2026-07-24"}
	for index, want := range wantDates {
		if got := schedules[index].ScheduleDate.Format("2006-01-02"); got != want {
			t.Fatalf("schedule %d date = %s, want %s", index, got, want)
		}
	}
}

func TestBuildGeneratedSchedulesSkipsScheduleFromMergedTemplate(t *testing.T) {
	scheduleDate := time.Date(2026, 7, 20, 0, 0, 0, 0, medicalBusinessLocation)
	template := models.MedScheduleTemplate{
		TemplateID:       scheduleTestUUID,
		DoctorID:         scheduleTestUUID,
		DepartmentID:     scheduleTestUUID,
		RegistrationType: "1",
		Weekdays:         []int{1},
		StartTime:        "08:00:00",
		EndTime:          "09:00:00",
		DefaultSlotQuota: 1,
		EffectiveDate:    scheduleDate,
		Status:           1,
	}
	legacyTemplateID := "00000000-0000-0000-0000-000000000002"
	existingSchedules := []models.MedSchedule{{
		TemplateID:       &legacyTemplateID,
		DoctorID:         template.DoctorID,
		DepartmentID:     template.DepartmentID,
		RegistrationType: template.RegistrationType,
		ScheduleDate:     scheduleDate,
		StartTime:        template.StartTime,
		EndTime:          template.EndTime,
		Status:           models.MedScheduleStatusDraft,
	}}
	relations := map[string]models.MedDoctorDepartment{
		scheduleDimensionKey(template.DoctorID, template.DepartmentID, template.RegistrationType): {},
	}

	schedules, _, skipped, err := buildGeneratedSchedules(
		[]models.MedScheduleTemplate{template},
		relations,
		existingSchedules,
		"batch-id",
		scheduleDate,
		scheduleDate,
		"",
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(schedules) != 0 || skipped != 1 {
		t.Fatalf("generated = %d, skipped = %d, want generated 0 and skipped 1", len(schedules), skipped)
	}
}

func TestBuildScheduleSlotDraftsUsesHalfHourStepsAndOverrides(t *testing.T) {
	slots, totalQuota, _, err := buildScheduleSlotDrafts(
		"schedule-id",
		"08:00:00",
		"10:00:00",
		3,
		[]models.ScheduleSlotQuotaRequest{{StartTime: "09:00", Quota: 0}},
		"",
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(slots) != 4 {
		t.Fatalf("slot count = %d, want 4", len(slots))
	}
	wantQuotas := []int{3, 3, 0, 3}
	for index, want := range wantQuotas {
		if slots[index].Quota != want {
			t.Fatalf("slot %d quota = %d, want %d", index, slots[index].Quota, want)
		}
	}
	if totalQuota != 9 {
		t.Fatalf("total quota = %d, want 9", totalQuota)
	}
}

func TestBuildScheduleSlotDraftsRejectsNonHalfHourBoundary(t *testing.T) {
	_, _, _, err := buildScheduleSlotDrafts("schedule-id", "08:15:00", "10:00:00", 1, nil, "")
	if !errors.Is(err, ErrMedicalInvalidInput) {
		t.Fatalf("buildScheduleSlotDrafts() error = %v, want ErrMedicalInvalidInput", err)
	}
}

func TestBuildScheduleSlotDraftsRejectsUnknownOverride(t *testing.T) {
	_, _, _, err := buildScheduleSlotDrafts(
		"schedule-id",
		"08:00:00",
		"09:00:00",
		1,
		[]models.ScheduleSlotQuotaRequest{{StartTime: "09:00", Quota: 2}},
		"",
	)
	if !errors.Is(err, ErrMedicalInvalidInput) {
		t.Fatalf("buildScheduleSlotDrafts() error = %v, want ErrMedicalInvalidInput", err)
	}
}

func TestScheduleAutomationWeekBoundaries(t *testing.T) {
	executedAt := time.Date(2026, 7, 19, 20, 0, 0, 0, medicalBusinessLocation)
	if got := scheduleWeekMonday(executedAt).Format("2006-01-02"); got != "2026-07-13" {
		t.Fatalf("scheduleWeekMonday() = %s, want 2026-07-13", got)
	}
	if got := nextScheduleAutomationRun(time.Date(2026, 7, 15, 9, 0, 0, 0, medicalBusinessLocation)); got.Format("2006-01-02 15:04") != "2026-07-19 20:00" {
		t.Fatalf("nextScheduleAutomationRun() = %s", got.Format("2006-01-02 15:04"))
	}
	if got := nextScheduleAutomationRun(time.Date(2026, 7, 19, 20, 1, 0, 0, medicalBusinessLocation)); got.Format("2006-01-02 15:04") != "2026-07-26 20:00" {
		t.Fatalf("missed run should not be compensated, got %s", got.Format("2006-01-02 15:04"))
	}
}

func TestMedicalScheduleServiceRejectsInvalidTimeRange(t *testing.T) {
	service := NewMedicalScheduleService()
	err := service.CreateScheduleTemplate(models.SaveScheduleTemplateRequest{
		ScheduleTemplateBaseRequest: models.ScheduleTemplateBaseRequest{
			TemplateName:     "上午门诊",
			DoctorID:         scheduleTestUUID,
			DepartmentID:     scheduleTestUUID,
			RegistrationType: "1",
			StartTime:        "09:00",
			EndTime:          "09:00",
			DefaultSlotQuota: 1,
			EffectiveDate:    medicalToday().Format("2006-01-02"),
			Status:           1,
		},
		Weekdays: []int{1},
	}, "")
	if !errors.Is(err, ErrMedicalInvalidInput) {
		t.Fatalf("CreateScheduleTemplate() error = %v, want ErrMedicalInvalidInput", err)
	}
}

func TestMedicalScheduleServiceLimitsGenerationTo90Days(t *testing.T) {
	service := NewMedicalScheduleService()
	startDate := medicalToday()
	_, err := service.GenerateSchedules(models.GenerateSchedulesRequest{
		IdempotencyKey: "schedule-test-91-days",
		TemplateIDs:    []string{scheduleTestUUID},
		StartDate:      startDate.Format("2006-01-02"),
		EndDate:        startDate.AddDate(0, 0, 90).Format("2006-01-02"),
	}, "")
	if !errors.Is(err, ErrMedicalInvalidInput) {
		t.Fatalf("GenerateSchedules() error = %v, want ErrMedicalInvalidInput", err)
	}
}

func TestMedicalScheduleServiceRejectsGenerationBeyondNext90Days(t *testing.T) {
	service := NewMedicalScheduleService()
	startDate := medicalToday().AddDate(0, 0, 90)
	_, err := service.GenerateSchedules(models.GenerateSchedulesRequest{
		IdempotencyKey: "schedule-test-too-far-in-future",
		TemplateIDs:    []string{scheduleTestUUID},
		StartDate:      startDate.Format("2006-01-02"),
		EndDate:        startDate.Format("2006-01-02"),
	}, "")
	if !errors.Is(err, ErrMedicalInvalidInput) {
		t.Fatalf("GenerateSchedules() error = %v, want ErrMedicalInvalidInput", err)
	}
}

func TestMedicalScheduleServiceRejectsDuplicateTemplateIDs(t *testing.T) {
	service := NewMedicalScheduleService()
	startDate := medicalToday()
	_, err := service.GenerateSchedules(models.GenerateSchedulesRequest{
		IdempotencyKey: "schedule-test-duplicate-template",
		TemplateIDs:    []string{scheduleTestUUID, scheduleTestUUID},
		StartDate:      startDate.Format("2006-01-02"),
		EndDate:        startDate.Format("2006-01-02"),
	}, "")
	if !errors.Is(err, ErrMedicalInvalidInput) {
		t.Fatalf("GenerateSchedules() error = %v, want ErrMedicalInvalidInput", err)
	}
}

func TestScheduleGenerationHashIsStableForSortedTemplateIDs(t *testing.T) {
	date := time.Date(2026, 7, 17, 0, 0, 0, 0, medicalBusinessLocation)
	leftIDs, err := normalizeScheduleUUIDs([]string{
		"00000000-0000-0000-0000-000000000002",
		"00000000-0000-0000-0000-000000000001",
	}, "排班模板ID", 100)
	if err != nil {
		t.Fatal(err)
	}
	rightIDs, err := normalizeScheduleUUIDs([]string{
		"00000000-0000-0000-0000-000000000001",
		"00000000-0000-0000-0000-000000000002",
	}, "排班模板ID", 100)
	if err != nil {
		t.Fatal(err)
	}
	leftHash, _, err := scheduleGenerationRequestHash(leftIDs, date, date)
	if err != nil {
		t.Fatal(err)
	}
	rightHash, _, err := scheduleGenerationRequestHash(rightIDs, date, date)
	if err != nil {
		t.Fatal(err)
	}
	if leftHash != rightHash {
		t.Fatalf("equivalent requests produced different hashes: %q != %q", leftHash, rightHash)
	}
}

func TestSelectGeneratedScheduleFeeRule(t *testing.T) {
	scheduleDate := time.Date(2026, 7, 17, 0, 0, 0, 0, medicalBusinessLocation)
	openStart := time.Date(2026, 7, 1, 0, 0, 0, 0, medicalBusinessLocation)
	secondStart := time.Date(2026, 7, 10, 0, 0, 0, 0, medicalBusinessLocation)

	t.Run("selects the only active rule", func(t *testing.T) {
		rule, err := selectGeneratedScheduleFeeRule([]models.MedRegistrationFeeRule{{
			FeeRuleID:     scheduleTestUUID,
			EffectiveDate: openStart,
		}}, scheduleDate)
		if err != nil {
			t.Fatal(err)
		}
		if rule.FeeRuleID != scheduleTestUUID {
			t.Fatalf("FeeRuleID = %q, want %q", rule.FeeRuleID, scheduleTestUUID)
		}
	})

	t.Run("rejects overlapping rules", func(t *testing.T) {
		_, err := selectGeneratedScheduleFeeRule([]models.MedRegistrationFeeRule{
			{FeeRuleID: scheduleTestUUID, EffectiveDate: openStart},
			{FeeRuleID: "00000000-0000-0000-0000-000000000002", EffectiveDate: secondStart},
		}, scheduleDate)
		if !errors.Is(err, ErrMedicalConflict) {
			t.Fatalf("selectGeneratedScheduleFeeRule() error = %v, want ErrMedicalConflict", err)
		}
	})

	t.Run("rejects missing rule", func(t *testing.T) {
		_, err := selectGeneratedScheduleFeeRule(nil, scheduleDate)
		if !errors.Is(err, ErrMedicalConflict) {
			t.Fatalf("selectGeneratedScheduleFeeRule() error = %v, want ErrMedicalConflict", err)
		}
	})
}
