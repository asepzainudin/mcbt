package middleware

import "testing"

func TestToSnakeCase(t *testing.T) {
	cases := map[string]string{
		"ClassID":       "class_id",
		"UserID":        "user_id",
		"Nip":           "nip",
		"Username":      "username",
		"AcademicYearID": "academic_year_id",
		"TargetClassID":  "target_class_id",
	}
	for in, want := range cases {
		if got := toSnakeCase(in); got != want {
			t.Errorf("toSnakeCase(%q) = %q, want %q", in, got, want)
		}
	}
}
