package usecase

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/go-pdf/fpdf"
	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"

	"github.com/asepzainudin14/mcbt/internal/model"
	"github.com/asepzainudin14/mcbt/internal/pkg/apperror"
	"github.com/asepzainudin14/mcbt/internal/repository"
)

type ExportUsecase struct {
	results  *ResultUsecase
	exams    *repository.ExamRepository
	students *repository.StudentRepository
	teachers *repository.TeacherRepository
}

func NewExportUsecase(
	results *ResultUsecase,
	exams *repository.ExamRepository,
	students *repository.StudentRepository,
	teachers *repository.TeacherRepository,
) *ExportUsecase {
	return &ExportUsecase{results: results, exams: exams, students: students, teachers: teachers}
}

type ExportFile struct {
	Filename    string
	ContentType string
	Data        []byte
}

var exportFormats = map[string]string{
	"xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	"pdf":  "application/pdf",
}

func normalizeFormat(raw string) (string, error) {
	f := strings.ToLower(strings.TrimSpace(raw))
	if f == "" || f == "excel" {
		f = "xlsx"
	}
	if _, ok := exportFormats[f]; !ok {
		return "", apperror.BadRequest("format harus xlsx atau pdf", nil)
	}
	return f, nil
}

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == ' ', r == '-', r == '_', r == '.':
			b.WriteRune('-')
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		out = "data"
	}
	return out
}

// ExamResults: rekap nilai + ranking ujian → xlsx/pdf.
func (u *ExportUsecase) ExamResults(ctx context.Context, examID uuid.UUID, format string) (*ExportFile, error) {
	format, err := normalizeFormat(format)
	if err != nil {
		return nil, err
	}
	exam, err := u.exams.FindByID(ctx, examID)
	if err != nil {
		return nil, apperror.NotFound("Ujian tidak ditemukan", err)
	}
	ranked, err := u.results.ExamResults(ctx, examID, nil)
	if err != nil {
		return nil, err
	}

	headers := []string{"Peringkat", "NIS", "Nama", "Kelas", "Skor", "KKM", "Status", "Percobaan", "Waktu Submit"}
	rows := make([][]string, 0, len(ranked))
	passedCount, totalScore := 0, 0.0
	for _, r := range ranked {
		score := "-"
		if r.Score != nil {
			score = fmt.Sprintf("%.1f", *r.Score)
			totalScore += *r.Score
			if r.Passed {
				passedCount++
			}
		}
		submittedAt := "-"
		if r.SubmittedAt != nil {
			submittedAt = r.SubmittedAt.Local().Format("02/01/2006 15:04")
		}
		className := "-"
		if r.ClassName != nil && *r.ClassName != "" {
			className = *r.ClassName
		}
		statusLulus := "TIDAK LULUS"
		if r.Passed {
			statusLulus = "LULUS"
		}
		rows = append(rows, []string{
			fmt.Sprintf("%d", r.Rank), r.Nis, r.StudentName, className,
			score, fmt.Sprintf("%.0f", r.PassingGrade), statusLulus,
			fmt.Sprintf("%d", r.AttemptsUsed), submittedAt,
		})
	}

	base := fmt.Sprintf("hasil-ujian-%s", slugify(exam.Title))
	switch format {
	case "xlsx":
		data, err := buildXLSX(fmt.Sprintf("Rekap Hasil - %s", exam.Title), headers, rows)
		if err != nil {
			return nil, apperror.Internal(err)
		}
		return &ExportFile{Filename: base + ".xlsx", ContentType: exportFormats["xlsx"], Data: data}, nil
	default:
		avgScore := 0.0
		if len(ranked) > 0 {
			avgScore = totalScore / float64(len(ranked))
		}
		summary := fmt.Sprintf(
			"Peserta lulus: %d/%d (%.0f%%) | Rata-rata skor: %.1f",
			passedCount, len(ranked), percent(passedCount, len(ranked)), avgScore,
		)
		widths := []float64{18, 26, 60, 24, 16, 12, 24, 18, 30}
		data, err := buildPDF(fmt.Sprintf("Rekap Hasil Ujian - %s", exam.Title), summary, headers, rows, widths)
		if err != nil {
			return nil, apperror.Internal(err)
		}
		return &ExportFile{Filename: base + ".pdf", ContentType: exportFormats["pdf"], Data: data}, nil
	}
}

// Students: daftar seluruh siswa → xlsx/pdf.
func (u *ExportUsecase) Students(ctx context.Context, format string) (*ExportFile, error) {
	format, err := normalizeFormat(format)
	if err != nil {
		return nil, err
	}
	page, err := u.students.List(ctx, "", nil, 1, 100000)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	headers := []string{"NIS", "Nama", "Email", "Kelas"}
	rows := make([][]string, 0, len(page.Items))
	for _, s := range page.Items {
		name, email := "", ""
		className := ""
		if s.User != nil {
			name = s.User.Name
			email = s.User.Email
		}
		if s.Class != nil {
			className = s.Class.Name
		}
		rows = append(rows, []string{s.Nis, name, email, className})
	}

	switch format {
	case "xlsx":
		data, err := buildXLSX("Daftar Siswa", headers, rows)
		if err != nil {
			return nil, apperror.Internal(err)
		}
		return &ExportFile{Filename: "daftar-siswa.xlsx", ContentType: exportFormats["xlsx"], Data: data}, nil
	default:
		data, err := buildPDF("Daftar Siswa", fmt.Sprintf("Total: %d siswa", page.TotalItems),
			headers, rows, []float64{28, 62, 78, 42})
		if err != nil {
			return nil, apperror.Internal(err)
		}
		return &ExportFile{Filename: "daftar-siswa.pdf", ContentType: exportFormats["pdf"], Data: data}, nil
	}
}

// Teachers: daftar seluruh guru → xlsx/pdf.
func (u *ExportUsecase) Teachers(ctx context.Context, format string) (*ExportFile, error) {
	format, err := normalizeFormat(format)
	if err != nil {
		return nil, err
	}
	page, err := u.teachers.List(ctx, "", 1, 100000)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	headers := []string{"NIP", "Nama", "Email"}
	rows := make([][]string, 0, len(page.Items))
	for _, t := range page.Items {
		name, email := "", ""
		nip := ""
		if t.User != nil {
			name = t.User.Name
			email = t.User.Email
		}
		if t.Nip != nil {
			nip = *t.Nip
		}
		rows = append(rows, []string{nip, name, email})
	}

	switch format {
	case "xlsx":
		data, err := buildXLSX("Daftar Guru", headers, rows)
		if err != nil {
			return nil, apperror.Internal(err)
		}
		return &ExportFile{Filename: "daftar-guru.xlsx", ContentType: exportFormats["xlsx"], Data: data}, nil
	default:
		data, err := buildPDF("Daftar Guru", fmt.Sprintf("Total: %d guru", page.TotalItems),
			headers, rows, []float64{34, 70, 90})
		if err != nil {
			return nil, apperror.Internal(err)
		}
		return &ExportFile{Filename: "daftar-guru.pdf", ContentType: exportFormats["pdf"], Data: data}, nil
	}
}

func percent(part, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(part) / float64(total) * 100
}

var _ = model.AttemptStatusSubmitted

// ---------- XLSX ----------

func buildXLSX(title string, headers []string, rows [][]string) ([]byte, error) {
	const sheet = "Data"
	f := excelize.NewFile()
	if err := f.SetSheetName("Sheet1", sheet); err != nil {
		return nil, err
	}

	styleTitle, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true, Size: 13}})
	styleHead, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "#FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#2563EB"}},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})

	f.SetCellValue(sheet, "A1", title)
	endCell, _ := excelize.CoordinatesToCellName(len(headers), 1)
	_ = f.MergeCell(sheet, "A1", endCell)
	_ = f.SetCellStyle(sheet, "A1", endCell, styleTitle)

	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 3)
		_ = f.SetCellValue(sheet, cell, h)
	}
	headEnd, _ := excelize.CoordinatesToCellName(len(headers), 3)
	firstHead, _ := excelize.CoordinatesToCellName(1, 3)
	_ = f.SetCellStyle(sheet, firstHead, headEnd, styleHead)

	for ri, row := range rows {
		rn := 4 + ri
		for ci, v := range row {
			cell, _ := excelize.CoordinatesToCellName(ci+1, rn)
			_ = f.SetCellValue(sheet, cell, v)
		}
	}

	for i := range headers {
		col, _ := excelize.ColumnNumberToName(i + 1)
		_ = f.SetColWidth(sheet, col, col, 16)
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ---------- PDF ----------

// buildPDF membuat PDF landscape A4 satu-baris-per-record dengan teks yang dipotong agar muat kolom.
func buildPDF(title, summary string, headers []string, rows [][]string, widths []float64) ([]byte, error) {
	pdf := fpdf.New("L", "mm", "A4", "")
	pdf.SetMargins(10, 12, 10)
	pdf.SetAutoPageBreak(false, 0)
	pdf.AddPage()

	pdf.SetFont("Helvetica", "B", 15)
	pdf.CellFormat(277, 9, fitText(pdf, title, 277), "", 1, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 10)
	pdf.CellFormat(277, 6, fitText(pdf, summary, 277), "", 1, "L", false, 0, "")
	pdf.Ln(2)

	drawHeader := func() {
		pdf.SetFillColor(37, 99, 235)
		pdf.SetTextColor(255, 255, 255)
		pdf.SetFont("Helvetica", "B", 9)
		pdf.SetX(10)
		for i, h := range headers {
			pdf.CellFormat(widths[i], 8, fitText(pdf, h, widths[i]-2), "1", 0, "C", true, 0, "")
		}
		pdf.Ln(-1)
		pdf.SetTextColor(33, 33, 33)
		pdf.SetFont("Helvetica", "", 9)
	}
	drawHeader()

	fill := false
	rowH := 6.5
	maxY := 195.0
	for _, row := range rows {
		if pdf.GetY()+rowH > maxY { // footer aman sebelum auto-break dimatikan
			pdf.AddPage()
			drawHeader()
		}
		pdf.SetX(10)
		for i, v := range row {
			align := "C"
			if i == 2 || i == 1 && len(headers) > 4 { // kolom nama/teks panjang rata kiri
				align = "L"
			}
			pdf.CellFormat(widths[i], rowH, fitText(pdf, v, widths[i]-2), "1", 0, align, fill, 0, "")
		}
		pdf.Ln(rowH)
		fill = !fill
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// fitText memotong teks agar muat pada lebar tertentu (mm).
func fitText(pdf *fpdf.Fpdf, s string, maxW float64) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if pdf.GetStringWidth(s) <= maxW {
		return s
	}
	runes := []rune(s)
	for len(runes) > 1 && pdf.GetStringWidth(string(runes)+"...") > maxW {
		runes = runes[:len(runes)-1]
	}
	return string(runes) + "..."
}
