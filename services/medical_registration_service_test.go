package services

import (
	"errors"
	"testing"

	"hive-admin-go/models"
)

func TestValidateRegistrationTransitionAllowsOnlyApprovedLifecycle(t *testing.T) {
	allowed := [][2]int{
		{models.MedRegistrationStatusPendingPayment, models.MedRegistrationStatusPaid},
		{models.MedRegistrationStatusPendingPayment, models.MedRegistrationStatusCanceled},
		{models.MedRegistrationStatusPaid, models.MedRegistrationStatusCheckedIn},
		{models.MedRegistrationStatusPaid, models.MedRegistrationStatusNoShow},
		{models.MedRegistrationStatusPaid, models.MedRegistrationStatusRefundStarted},
		{models.MedRegistrationStatusCheckedIn, models.MedRegistrationStatusCompleted},
		{models.MedRegistrationStatusRefundStarted, models.MedRegistrationStatusRefunding},
		{models.MedRegistrationStatusRefunding, models.MedRegistrationStatusRefunded},
	}
	for _, transition := range allowed {
		if err := ValidateRegistrationTransition(transition[0], transition[1]); err != nil {
			t.Fatalf("transition %d -> %d rejected: %v", transition[0], transition[1], err)
		}
	}

	rejected := [][2]int{
		{models.MedRegistrationStatusPendingPayment, models.MedRegistrationStatusCheckedIn},
		{models.MedRegistrationStatusPaid, models.MedRegistrationStatusCompleted},
		{models.MedRegistrationStatusCompleted, models.MedRegistrationStatusCanceled},
		{models.MedRegistrationStatusRefundStarted, models.MedRegistrationStatusRefunded},
	}
	for _, transition := range rejected {
		if err := ValidateRegistrationTransition(transition[0], transition[1]); !errors.Is(err, ErrMedicalConflict) {
			t.Fatalf("transition %d -> %d error = %v, want ErrMedicalConflict", transition[0], transition[1], err)
		}
	}
}

func TestRegistrationTransitionReleasesQuotaOnlyAtApprovedEndpoints(t *testing.T) {
	if !RegistrationTransitionReleasesQuota(models.MedRegistrationStatusPendingPayment, models.MedRegistrationStatusCanceled) {
		t.Fatal("pending payment -> canceled should release quota")
	}
	if !RegistrationTransitionReleasesQuota(models.MedRegistrationStatusRefunding, models.MedRegistrationStatusRefunded) {
		t.Fatal("refunding -> refunded should release quota")
	}
	if RegistrationTransitionReleasesQuota(models.MedRegistrationStatusPaid, models.MedRegistrationStatusRefundStarted) {
		t.Fatal("refund start must keep quota occupied")
	}
	if RegistrationTransitionReleasesQuota(models.MedRegistrationStatusRefundStarted, models.MedRegistrationStatusRefunding) {
		t.Fatal("refund processing must keep quota occupied")
	}
}
