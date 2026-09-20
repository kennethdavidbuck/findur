package buildinfo

import "testing"

func TestValidateProduction(t *testing.T) {
	original := SHA
	t.Cleanup(func() { SHA = original })

	for _, test := range []struct {
		sha   string
		valid bool
	}{
		{sha: "0123456789abcdef0123456789abcdef01234567", valid: true},
		{sha: "development", valid: false},
		{sha: "0123456789ABCDEF0123456789ABCDEF01234567", valid: false},
		{sha: "0123456", valid: false},
	} {
		SHA = test.sha
		if err := ValidateProduction(); (err == nil) != test.valid {
			t.Fatalf("ValidateProduction() error = %v for %q, valid = %v", err, test.sha, test.valid)
		}
	}
}
