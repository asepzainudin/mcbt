package usecase

import (
	"bytes"
	"testing"

	"github.com/xuri/excelize/v2"

	"github.com/asepzainudin14/mcbt/internal/pkg/apperror"
)

func TestNormalizeFormat(t *testing.T) {
	cases := []struct{ in, want string }{
		{"", "xlsx"},
		{"xlsx", "xlsx"},
		{"XLSX", "xlsx"},
		{"excel", "xlsx"},
		{" pdf ", "pdf"},
	}
	for _, c := range cases {
		got, err := normalizeFormat(c.in)
		if err != nil {
			t.Errorf("normalizeFormat(%q) error: %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("normalizeFormat(%q) = %q, want %q", c.in, got, c.want)
		}
	}

	if _, err := normalizeFormat("docx"); err == nil {
		t.Error("format tidak didukung seharusnya error")
	} else if ae, ok := err.(*apperror.AppError); !ok || ae.Code != apperror.CodeBadRequest {
		t.Errorf("harus bad request, got %v", err)
	}
}

func TestSlugify(t *testing.T) {
	cases := []struct{ in, want string }{
		{"Ujian Akhir Semester", "ujian-akhir-semester"},
		{"  Matematika -- Kelas 7A  ", "matematika-kelas-7a"},
		{"", "data"},
		{"!!!???", "data"},
	}
	for _, c := range cases {
		if got := slugify(c.in); got != c.want {
			t.Errorf("slugify(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestPercent(t *testing.T) {
	if percent(1, 4) != 25.0 {
		t.Errorf("percent(1,4) = %v, want 25", percent(1, 4))
	}
	if percent(0, 0) != 0 {
		t.Errorf("percent(0,0) harus 0 (tanpa panic)")
	}
}

func TestBuildXLSX(t *testing.T) {
	data, err := buildXLSX(
		"Judul Dokumen",
		[]string{"Kolom A", "Kolom B"},
		[][]string{{"1", "dua"}, {"3", "empat"}},
	)
	if err != nil {
		t.Fatalf("buildXLSX error: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("hasil kosong")
	}

	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("buka hasil xlsx: %v", err)
	}
	defer f.Close()

	rows, err := f.GetRows("Data")
	if err != nil {
		t.Fatalf("ambil rows: %v", err)
	}
	// GetRows menyertakan baris kosong sbg entry kosong:
	// [judul, [], header, data1, data2]
	if len(rows) < 5 {
		t.Fatalf("jumlah baris = %d, want >= 5", len(rows))
	}
	header, d1, d2 := rows[2], rows[3], rows[4]
	if len(header) != 2 || header[0] != "Kolom A" || header[1] != "Kolom B" {
		t.Errorf("header tidak sesuai: %v", header)
	}
	if len(d1) != 2 || d1[0] != "1" || d1[1] != "dua" {
		t.Errorf("data1 tidak sesuai: %v", d1)
	}
	if len(d2) != 2 || d2[0] != "3" || d2[1] != "empat" {
		t.Errorf("data2 tidak sesuai: %v", d2)
	}
}
