package passwordutil

import "testing"

func TestGenerate_Length(t *testing.T) {
	for _, want := range []int{8, 10, 16, 32} {
		got, err := Generate(want)
		if err != nil {
			t.Fatalf("Generate(%d) error: %v", want, err)
		}
		if len(got) != want {
			t.Errorf("Generate(%d) len = %d, want %d", want, len(got), want)
		}
	}
}

func TestGenerate_MinimumLength(t *testing.T) {
	got, err := Generate(3)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(got) != 8 {
		t.Errorf("len = %d, want minimum 8", len(got))
	}
}

func TestGenerate_OnlyAllowedCharset(t *testing.T) {
	const allowed = "abcdefghjkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	got, err := Generate(64)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	for _, c := range got {
		if !containsRune(allowed, c) {
			t.Errorf("karakter tidak diizinkan: %q", c)
		}
	}
}

func containsRune(s string, r rune) bool {
	for _, c := range s {
		if c == r {
			return true
		}
	}
	return false
}

func TestDefaultPassword(t *testing.T) {
	if DefaultPassword != "McBT@1234" {
		t.Errorf("DefaultPassword berubah: %q", DefaultPassword)
	}
}
