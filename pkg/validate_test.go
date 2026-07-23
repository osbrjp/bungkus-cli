package pkg

import "testing"

func TestValidateProjectName(t *testing.T) {
	cases := []struct {
		name string
		in   string
		ok   bool
	}{
		{"simple", "my-app", true},
		{"digits and dots", "app2.0_beta", true},
		{"single char", "a", true},
		{"empty", "", false},
		{"uppercase", "MyApp", false},
		{"space", "bad name", false},
		{"traversal", "../../foo", false},
		{"absolute", "/etc/foo", false},
		{"nested path", "a/b", false},
		{"leading dot", ".hidden", false},
		{"leading dash", "-app", false},
		{"shell metachars", "app;rm -rf", false},
		{"too long", string(make([]byte, 215)), false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := ValidateProjectName(c.in)
			if c.ok && err != nil {
				t.Errorf("ValidateProjectName(%q) = %v, want nil", c.in, err)
			}
			if !c.ok && err == nil {
				t.Errorf("ValidateProjectName(%q) = nil, want error", c.in)
			}
		})
	}
}

func TestValidateDest(t *testing.T) {
	cases := []struct {
		in string
		ok bool
	}{
		{".", true},
		{"my-app", true},
		{"nested/ok", true},
		{"..", false},
		{"../escape", false},
		{"/abs/path", false},
		{"", false},
	}
	for _, c := range cases {
		err := ValidateDest(c.in)
		if c.ok && err != nil {
			t.Errorf("ValidateDest(%q) = %v, want nil", c.in, err)
		}
		if !c.ok && err == nil {
			t.Errorf("ValidateDest(%q) = nil, want error", c.in)
		}
	}
}
