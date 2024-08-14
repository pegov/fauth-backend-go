package model

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestValidateEmail(t *testing.T) {
	validUserPartEmail := strings.Repeat("a", maxUserPartLen) + "@test.com"
	invalidUserPartEmail := strings.Repeat("a", maxUserPartLen+1) + "@test.com"

	validDomainPartEmail := "aaa@" + strings.Repeat("a", maxDomainPart-4) + ".com"
	invalidDomainPartEmail := "aaa@" + strings.Repeat("a", maxDomainPart+1-4) + ".com"

	validEmail := "test@test.com"

	tests := []struct {
		name       string
		input      string
		wantOutput string
		wantErr    error
	}{
		{"Email must not be empty", "", "", ErrEmailEmpty},
		{"Email must not be just spaces", "   ", "", ErrEmailEmpty},
		{"Email must must contain @", "first_part", "", ErrEmailMissingSeparator},
		{
			fmt.Sprintf("Email user part must be <= %d", maxUserPartLen),
			validUserPartEmail,
			validUserPartEmail,
			nil,
		},
		{
			fmt.Sprintf("Email user part must be <= %d", maxUserPartLen),
			invalidUserPartEmail,
			"",
			ErrEmailLength,
		},
		{
			fmt.Sprintf("Email domain part must be <= %d", maxDomainPart),
			validDomainPartEmail,
			validDomainPartEmail,
			nil,
		},
		{
			fmt.Sprintf("Email domain part must be <= %d", maxDomainPart),
			invalidDomainPartEmail,
			"",
			ErrEmailLength,
		},
		{
			"Email must match regex",
			"userpart@nodot",
			"",
			ErrEmailWrong,
		},
		{
			"Email must match regex",
			validEmail,
			validEmail,
			nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := validateEmail(tt.input)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("%s: got err = %v, want = %v", tt.name, err, tt.wantErr)
			}

			if output != tt.wantOutput {
				t.Fatalf("%s: got output = %s, want = %s", tt.name, output, tt.wantOutput)
			}
		})
	}
}
