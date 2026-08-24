package usecase

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"

	"github.com/asepzainudin14/mcbt/internal/model"
	"github.com/asepzainudin14/mcbt/internal/pkg/apperror"
	"github.com/asepzainudin14/mcbt/internal/repository"
)

const importTokenTTL = 15 * time.Minute

var questionImportColumns = []string{
	"type", "content", "score_weight", "explanation",
	"option_a", "option_b", "option_c", "option_d", "option_e", "answer",
}

type QuestionDraft struct {
	Type        string
	Content     string
	ScoreWeight float64
	Explanation *string
	AnswerKeys  *string
	Options     []model.Option
}

type importJob struct {
	BankID   uuid.UUID
	Drafts   []QuestionDraft
	ExpireAt time.Time
}

type ImportTokenStore struct {
	mu   sync.Mutex
	jobs map[string]importJob
	now  func() time.Time
}

func NewImportTokenStore() *ImportTokenStore {
	return &ImportTokenStore{jobs: make(map[string]importJob), now: time.Now}
}

func (s *ImportTokenStore) put(bankID uuid.UUID, drafts []QuestionDraft) (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	token := hex.EncodeToString(b)
	s.mu.Lock()
	defer s.mu.Unlock()
	for k, v := range s.jobs {
		if s.now().After(v.ExpireAt) {
			delete(s.jobs, k)
		}
	}
	s.jobs[token] = importJob{BankID: bankID, Drafts: drafts, ExpireAt: s.now().Add(importTokenTTL)}
	return token, nil
}

func (s *ImportTokenStore) pop(token string) (importJob, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	job, ok := s.jobs[token]
	if !ok || s.now().After(job.ExpireAt) {
		delete(s.jobs, token)
		return importJob{}, errors.New("token tidak ditemukan atau kedaluwarsa")
	}
	delete(s.jobs, token)
	return job, nil
}

type ImportRowError struct {
	Row    int    `json:"row"`
	Field  string `json:"field"`
	Reason string `json:"reason"`
}

type ImportResult struct {
	ImportedCount int              `json:"imported_count"`
	Skipped       []ImportRowError `json:"skipped"`
}

type QuestionImportUsecase struct {
	store *ImportTokenStore
	repo  *repository.QuestionRepository
}

func NewQuestionImportUsecase(store *ImportTokenStore, repo *repository.QuestionRepository) *QuestionImportUsecase {
	return &QuestionImportUsecase{store: store, repo: repo}
}

func (u *QuestionImportUsecase) TemplateXLSX() ([]byte, error) {
	sample := [][]any{
		{"multiple_choice", "Ibu kota Indonesia adalah ...", "1", "Soal pengetahuan umum", "Jakarta", "Bandung", "Surabaya", "Medan", "", "A"},
		{"multiple_answer", "Pilih bilangan prima berikut", "2", "", "2", "3", "4", "9", "", "A,B"},
		{"true_false", "Matahari terbit dari barat.", "1", "", "BENAR", "SALAH", "", "", "", "B"},
		{"essay", "Jelaskan proses fotosintesis!", "5", "Jawaban mencakup cahaya, klorofil, air, CO2", "", "", "", "", "", ""},
		{"short_answer", "Warna bendera Indonesia (merah dan ...)", "2", "", "", "", "", "", "", "putih"},
	}
	return buildTemplate("soal", questionImportColumns, sample)
}

func (u *QuestionImportUsecase) Validate(ctx context.Context, fileBytes []byte, bankID uuid.UUID) (string, int, []ImportRowError, error) {
	rows, err := parseSheet(fileBytes, questionImportColumns)
	if err != nil {
		return "", 0, nil, apperror.BadRequest(err.Error(), nil)
	}
	if len(rows) == 0 {
		return "", 0, nil, apperror.New(422, "File tidak berisi data soal", nil)
	}

	var drafts []QuestionDraft
	var skipped []ImportRowError

	for i, row := range rows {
		rowNo := i + 2

		qType := lower(row[0])
		content := row[1]
		weightRaw := row[2]
		explanation := row[3]
		optionCells := []string{row[4], row[5], row[6], row[7], row[8]}
		answerRaw := strings.TrimSpace(row[9])
		answer := strings.ToUpper(answerRaw)

		fieldErr := func(field, reason string) {
			skipped = append(skipped, ImportRowError{Row: rowNo, Field: field, Reason: reason})
		}

		switch qType {
		case model.QuestionTypeMultipleChoice, model.QuestionTypeTrueFalse,
			model.QuestionTypeMultipleAnswer, model.QuestionTypeEssay,
			model.QuestionTypeShortAnswer:
		default:
			fieldErr("type", fmt.Sprintf("harus %s/%s/%s/%s/%s",
				model.QuestionTypeMultipleChoice, model.QuestionTypeTrueFalse,
				model.QuestionTypeMultipleAnswer, model.QuestionTypeEssay, model.QuestionTypeShortAnswer))
			continue
		}

		if content == "" {
			fieldErr("content", "wajib diisi")
			continue
		}

		scoreWeight := 1.0
		if weightRaw != "" {
			parsed, convErr := strconv.ParseFloat(weightRaw, 64)
			if convErr != nil || parsed <= 0 {
				fieldErr("score_weight", "harus angka lebih dari 0")
				continue
			}
			scoreWeight = parsed
		}

		draft := QuestionDraft{
			Type:        qType,
			Content:     content,
			ScoreWeight: scoreWeight,
			Options:     []model.Option{},
		}
		if explanation != "" {
			exp := explanation
			draft.Explanation = &exp
		}

		switch qType {
		case model.QuestionTypeTrueFalse:
			opts := []model.Option{
				{Label: "A", Content: "BENAR", Position: 0},
				{Label: "B", Content: "SALAH", Position: 1},
			}
			if answer != "A" && answer != "B" {
				fieldErr("answer", "untuk true_false harus A atau B")
				continue
			}
			opts[answer[0]-'A'].IsCorrect = true
			draft.Options = opts

		case model.QuestionTypeEssay:
			// esai tidak memiliki opsi

		case model.QuestionTypeShortAnswer:
			if answerRaw == "" {
				fieldErr("answer", "isian singkat wajib memiliki jawaban yang diterima")
				continue
			}
			keys := answerRaw
			draft.AnswerKeys = &keys

		default: // multiple_choice & multiple_answer
			filledCount := 0
			for idx, cell := range optionCells {
				if cell != "" {
					filledCount++
					draft.Options = append(draft.Options, model.Option{
						Label:    string(rune('A' + idx)),
						Content:  cell,
						Position: int16(idx),
					})
				}
			}
			if filledCount < 2 {
				fieldErr("option_a", "minimal memiliki opsi A dan B")
				continue
			}

			letters := strings.Split(strings.ToUpper(strings.ReplaceAll(answer, " ", "")), ",")
			valid := len(letters) > 0
			for _, L := range letters {
				if len(L) != 1 || L[0] < 'A' || int(L[0]-'A') >= len(draft.Options) {
					valid = false
					break
				}
			}
			if !valid {
				fieldErr("answer", "jawaban harus huruf opsi yang terisi (contoh: A atau A,C)")
				continue
			}
			if qType == model.QuestionTypeMultipleChoice && len(letters) > 1 {
				fieldErr("answer", "multiple_choice hanya boleh satu jawaban benar")
				continue
			}
			for _, L := range letters {
				draft.Options[L[0]-'A'].IsCorrect = true
			}
		}

		drafts = append(drafts, draft)
	}

	if len(skipped) > 0 {
		return "", len(drafts), skipped, apperror.New(422, "Terdapat baris tidak valid — perbaiki lalu unggah ulang", nil)
	}

	token, err := u.store.put(bankID, drafts)
	if err != nil {
		return "", 0, nil, apperror.Internal(err)
	}
	return token, len(drafts), nil, nil
}

func (u *QuestionImportUsecase) Process(ctx context.Context, token string) (*ImportResult, error) {
	job, err := u.store.pop(token)
	if err != nil {
		return nil, apperror.New(422, err.Error(), nil)
	}

	questions := make([]model.Question, 0, len(job.Drafts))
	for _, d := range job.Drafts {
		questions = append(questions, model.Question{
			QuestionBankID: job.BankID,
			QuestionType:   d.Type,
			Content:        d.Content,
			ScoreWeight:    d.ScoreWeight,
			Explanation:    d.Explanation,
			AnswerKeys:     d.AnswerKeys,
			Options:        d.Options,
		})
	}

	if err := u.repo.ImportQuestions(ctx, questions); err != nil {
		return nil, err
	}
	return &ImportResult{ImportedCount: len(questions)}, nil
}

func parseSheet(fileBytes []byte, columns []string) ([][]string, error) {
	f, err := excelize.OpenReader(bytes.NewReader(fileBytes))
	if err != nil {
		return nil, fmt.Errorf("file Excel tidak dapat dibaca")
	}
	defer f.Close()

	sheet := f.GetSheetList()[0]
	tableRows, err := f.GetRows(sheet)
	if err != nil {
		return nil, fmt.Errorf("sheet kosong atau tidak terbaca")
	}
	if len(tableRows) < 1 {
		return nil, fmt.Errorf("file harus memiliki baris header")
	}

	header := normalizeHeader(tableRows[0])
	want := columns
	if len(header) < len(want) {
		return nil, fmt.Errorf("header harus berisi kolom: %s", strings.Join(want, ", "))
	}
	for i, w := range want {
		if header[i] != w {
			return nil, fmt.Errorf("header harus berisi kolom: %s", strings.Join(want, ", "))
		}
	}

	var data [][]string
	for _, r := range tableRows[1:] {
		row := make([]string, len(columns))
		for i := range columns {
			if i < len(r) {
				row[i] = strings.TrimSpace(r[i])
			}
		}
		data = append(data, row)
	}
	return data, nil
}

func normalizeHeader(cells []string) []string {
	out := make([]string, len(cells))
	for i, c := range cells {
		out[i] = strings.ToLower(strings.TrimSpace(c))
	}
	return out
}

func buildTemplate(sheetName string, columns []string, sample [][]any) ([]byte, error) {
	f := excelize.NewFile()
	sheet := f.GetSheetName(0)
	if sheetName != "Sheet1" {
		f.SetSheetName(sheet, sheetName)
		sheet = sheetName
	}
	for i, col := range columns {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(sheet, cell, col)
	}
	for rIdx, row := range sample {
		for cIdx, v := range row {
			cell, _ := excelize.CoordinatesToCellName(cIdx+1, rIdx+2)
			_ = f.SetCellValue(sheet, cell, v)
		}
	}
	widths := map[int]float64{1: 18, 2: 36, 3: 12, 4: 30, 5: 18, 6: 18, 7: 18, 8: 18, 9: 18, 10: 12}
	for idx, w := range widths {
		col, _ := excelize.ColumnNumberToName(idx + 1)
		_ = f.SetColWidth(sheet, col, col, w)
	}
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

var _ = gorm.ErrRecordNotFound

func trimSpace(s string) string { return strings.TrimSpace(s) }

func lower(s string) string { return strings.ToLower(s) }
