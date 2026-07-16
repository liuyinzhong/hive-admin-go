package services

import "testing"

func TestValidateDoctorDepartmentSelection(t *testing.T) {
	tests := []struct {
		name          string
		departmentIDs []string
		primaryID     string
		wantError     bool
	}{
		{name: "valid", departmentIDs: []string{"a", "b"}, primaryID: "a"},
		{name: "empty", departmentIDs: nil, primaryID: "", wantError: true},
		{name: "duplicate", departmentIDs: []string{"a", "a"}, primaryID: "a", wantError: true},
		{name: "primary missing", departmentIDs: []string{"a", "b"}, primaryID: "c", wantError: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateDoctorDepartmentSelection(test.departmentIDs, test.primaryID)
			if (err != nil) != test.wantError {
				t.Fatalf("validateDoctorDepartmentSelection() error = %v, wantError = %v", err, test.wantError)
			}
		})
	}
}
