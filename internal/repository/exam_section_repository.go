package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/asepzainudin14/mcbt/internal/model"
)

type ExamSectionRepository struct {
	db *gorm.DB
}

func NewExamSectionRepository(db *gorm.DB) *ExamSectionRepository {
	return &ExamSectionRepository{db: db}
}

type SectionWithCount struct {
	model.ExamSection
	QuestionCount int64 `json:"question_count"`
}

func (r *ExamSectionRepository) ListByExam(ctx context.Context, examID uuid.UUID) ([]SectionWithCount, error) {
	var sections []model.ExamSection
	err := r.db.WithContext(ctx).
		Where("exam_id = ?", examID).
		Order("sequence ASC").
		Find(&sections).Error
	if err != nil {
		return nil, err
	}

	out := make([]SectionWithCount, 0, len(sections))
	for _, s := range sections {
		var count int64
		if err := r.db.WithContext(ctx).Model(&model.ExamSectionQuestion{}).
			Where("section_id = ?", s.ID).
			Count(&count).Error; err != nil {
			return nil, err
		}
		out = append(out, SectionWithCount{ExamSection: s, QuestionCount: count})
	}
	return out, nil
}

func (r *ExamSectionRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.ExamSection, error) {
	var s model.ExamSection
	err := r.db.WithContext(ctx).First(&s, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *ExamSectionRepository) Create(ctx context.Context, s *model.ExamSection) error {
	return TranslateDBError(r.db.WithContext(ctx).Create(s).Error,
		"Nama atau urutan section sudah digunakan di ujian ini")
}

func (r *ExamSectionRepository) Update(ctx context.Context, s *model.ExamSection) error {
	return TranslateDBError(
		r.db.WithContext(ctx).
			Model(s).
			Select("name", "sequence", "updated_at").
			Updates(s).Error,
		"Nama atau urutan section sudah digunakan di ujian ini")
}

func (r *ExamSectionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	res := r.db.WithContext(ctx).Delete(&model.ExamSection{}, "id = ?", id)
	if res.Error != nil {
		return TranslateDBError(res.Error, "")
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// QuestionIDsByBanks returns all question ids belonging to the given banks.
func (r *ExamSectionRepository) QuestionIDsByBanks(ctx context.Context, bankIDs []uuid.UUID) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := r.db.WithContext(ctx).Model(&model.Question{}).
		Where("question_bank_id IN ?", bankIDs).
		Pluck("id", &ids).Error
	return ids, err
}

// MappedQuestionIDsByExam returns question ids already mapped in ANY section of the exam.
func (r *ExamSectionRepository) MappedQuestionIDsByExam(ctx context.Context, examID uuid.UUID) (map[uuid.UUID]bool, error) {
	var ids []uuid.UUID
	err := r.db.WithContext(ctx).
		Model(&model.ExamSectionQuestion{}).
		Joins("JOIN exam_sections es ON es.id = exam_section_questions.section_id").
		Where("es.exam_id = ?", examID).
		Pluck("exam_section_questions.question_id", &ids).Error
	if err != nil {
		return nil, err
	}
	set := make(map[uuid.UUID]bool, len(ids))
	for _, id := range ids {
		set[id] = true
	}
	return set, nil
}

// InsertMappings skips rows whose (section,question) already exists.
func (r *ExamSectionRepository) InsertMappings(ctx context.Context, sectionID uuid.UUID, questionIDs []uuid.UUID) (int, error) {
	var inserted int
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, qid := range questionIDs {
			res := tx.Where("section_id = ? AND question_id = ?", sectionID, qid).
				FirstOrCreate(&model.ExamSectionQuestion{SectionID: sectionID, QuestionID: qid})
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected > 0 {
				inserted++
			}
		}
		return nil
	})
	if err != nil {
		return 0, TranslateDBError(err, "")
	}
	return inserted, nil
}

func (r *ExamSectionRepository) RemoveQuestion(ctx context.Context, sectionID, questionID uuid.UUID) error {
	res := r.db.WithContext(ctx).
		Where("section_id = ? AND question_id = ?", sectionID, questionID).
		Delete(&model.ExamSectionQuestion{})
	if res.Error != nil {
		return TranslateDBError(res.Error, "")
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// ListQuestions returns questions mapped to the section, ordered by mapping time.
func (r *ExamSectionRepository) ListQuestions(ctx context.Context, sectionID uuid.UUID) ([]model.Question, error) {
	var questions []model.Question
	err := r.db.WithContext(ctx).
		Joins("JOIN exam_section_questions esq ON esq.question_id = questions.id").
		Where("esq.section_id = ?", sectionID).
		Preload("QuestionBank").
		Order("esq.created_at ASC").
		Find(&questions).Error
	return questions, err
}

// ListQuestions preload sudah di-usecase; versi dengan bank:

type ExamQuestionGroup struct {
	Section   model.ExamSection
	Questions []model.Question
}

// ListExamQuestions groups mapped questions per section (sequence order).
// Falls back to bank questions grouped in one pseudo-section when no mapping exists.
func (r *ExamSectionRepository) ListExamQuestions(ctx context.Context, exam *model.Exam) ([]ExamQuestionGroup, error) {
	var sections []model.ExamSection
	if err := r.db.WithContext(ctx).
		Where("exam_id = ?", exam.ID).
		Order("sequence ASC").
		Find(&sections).Error; err != nil {
		return nil, err
	}

	groups := make([]ExamQuestionGroup, 0, len(sections)+1)
	hasAny := false
	for _, s := range sections {
		var qs []model.Question
		if err := r.db.WithContext(ctx).
			Joins("JOIN exam_section_questions esq ON esq.question_id = questions.id").
			Where("esq.section_id = ?", s.ID).
			Preload("Options", func(db *gorm.DB) *gorm.DB { return db.Order("position ASC") }).
			Preload("Options.Media").
			Preload("Media").
			Order("questions.created_at ASC").
			Find(&qs).Error; err != nil {
			return nil, err
		}
		if len(qs) > 0 {
			hasAny = true
		}
		groups = append(groups, ExamQuestionGroup{Section: s, Questions: qs})
	}

	if !hasAny && exam.QuestionBankID != nil {
		var qs []model.Question
		if err := r.db.WithContext(ctx).
			Where("question_bank_id = ?", exam.QuestionBankID).
			Preload("Options", func(db *gorm.DB) *gorm.DB { return db.Order("position ASC") }).
			Preload("Options.Media").
			Preload("Media").
			Order("created_at ASC").
			Find(&qs).Error; err != nil {
			return nil, err
		}
		groups = append(groups, ExamQuestionGroup{
			Section:   model.ExamSection{Name: "Soal", Sequence: 1},
			Questions: qs,
		})
	}
	return groups, nil
}

// QuestionInExam checks whether a question is mapped in any section of the exam.
func (r *ExamSectionRepository) QuestionInExam(ctx context.Context, examID, questionID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.ExamSectionQuestion{}).
		Joins("JOIN exam_sections es ON es.id = exam_section_questions.section_id").
		Where("es.exam_id = ? AND exam_section_questions.question_id = ?", examID, questionID).
		Count(&count).Error
	return count > 0, err
}
