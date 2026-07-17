package services

import (
	"testing"
	"time"

	"hive-admin-go/models"
)

func TestNormalizeRegistrationFeeAmount(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		want      string
		wantError bool
	}{
		{name: "integer", value: "30", want: "30.00"},
		{name: "one decimal", value: "30.5", want: "30.50"},
		{name: "two decimals", value: "0.01", want: "0.01"},
		{name: "zero", value: "0", wantError: true},
		{name: "negative", value: "-1", wantError: true},
		{name: "too many decimals", value: "10.001", wantError: true},
		{name: "too large", value: "100000000.00", wantError: true},
		{name: "leading zero", value: "030.00", wantError: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := normalizeRegistrationFeeAmount(test.value)
			if (err != nil) != test.wantError {
				t.Fatalf("normalizeRegistrationFeeAmount() error = %v, wantError = %v", err, test.wantError)
			}
			if got != test.want {
				t.Fatalf("normalizeRegistrationFeeAmount() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestRegistrationFeePeriodsOverlap(t *testing.T) {
	date := func(value string) time.Time {
		parsed, err := time.Parse("2006-01-02", value)
		if err != nil {
			t.Fatal(err)
		}
		return parsed
	}
	datePtr := func(value string) *time.Time {
		parsed := date(value)
		return &parsed
	}

	tests := []struct {
		name       string
		leftStart  time.Time
		leftEnd    *time.Time
		rightStart time.Time
		rightEnd   *time.Time
		want       bool
	}{
		{name: "same day is overlap", leftStart: date("2026-01-01"), leftEnd: datePtr("2026-01-31"), rightStart: date("2026-01-31"), rightEnd: datePtr("2026-02-28"), want: true},
		{name: "adjacent periods do not overlap", leftStart: date("2026-01-01"), leftEnd: datePtr("2026-01-31"), rightStart: date("2026-02-01"), rightEnd: datePtr("2026-02-28")},
		{name: "open period overlaps future", leftStart: date("2026-01-01"), rightStart: date("2027-01-01"), rightEnd: datePtr("2027-12-31"), want: true},
		{name: "bounded period before open period", leftStart: date("2026-01-01"), leftEnd: datePtr("2026-12-31"), rightStart: date("2027-01-01")},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := registrationFeePeriodsOverlap(test.leftStart, test.leftEnd, test.rightStart, test.rightEnd); got != test.want {
				t.Fatalf("registrationFeePeriodsOverlap() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestValidateRegistrationFeePeriod(t *testing.T) {
	expiryBeforeStart := "2026-01-31"
	expiryOnStart := "2026-02-01"
	tests := []struct {
		name          string
		effectiveDate string
		expiryDate    *string
		wantError     bool
	}{
		{name: "open ended", effectiveDate: "2026-02-01"},
		{name: "same day period", effectiveDate: "2026-02-01", expiryDate: &expiryOnStart},
		{name: "expiry before start", effectiveDate: "2026-02-01", expiryDate: &expiryBeforeStart, wantError: true},
		{name: "invalid date", effectiveDate: "2026/02/01", wantError: true},
		{name: "missing effective date", effectiveDate: "", wantError: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, _, err := validateRegistrationFeePeriod(test.effectiveDate, test.expiryDate)
			if (err != nil) != test.wantError {
				t.Fatalf("validateRegistrationFeePeriod() error = %v, wantError = %v", err, test.wantError)
			}
		})
	}
}

func TestNextRegistrationFeeVersion(t *testing.T) {
	rules := []models.MedRegistrationFeeRule{{Version: 1}, {Version: 3}, {Version: 2}}
	if got := nextRegistrationFeeVersion(rules); got != 4 {
		t.Fatalf("nextRegistrationFeeVersion() = %d, want 4", got)
	}
}

func TestMedicalDateComparisonIgnoresStorageTimezone(t *testing.T) {
	utcDate := time.Date(2026, 7, 17, 0, 0, 0, 0, time.UTC)
	businessDate := time.Date(2026, 7, 17, 0, 0, 0, 0, medicalBusinessLocation)
	if medicalDateBefore(utcDate, businessDate) || medicalDateAfter(utcDate, businessDate) {
		t.Fatal("the same business date must compare equal across storage timezones")
	}
}
