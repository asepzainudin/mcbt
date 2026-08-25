package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/asepzainudin14/mcbt/internal/model"
)

type ExamParticipantRepository struct {
	db *gorm.DB
}

func NewExamParticipantRepository(db *gorm.DB) *ExamParticipantRepository {
	return &ExamParticipantRepository{db: db}
}

func (r *ExamParticipantRepository) ListByExam(ctx context.Context, examID uuid.UUID) ([]model.ExamParticipant, error) {
	var participants []model.ExamParticipant
	err := r.db.WithContext(ctx).
		Preload("Student").
		Preload("Student.User").
		Preload("Student.Class").
		Where("exam_id = ?", examID).
		Order("created_at ASC").
		Find(&participants).Error
	return participants, err
}

// StudentIDsByClasses returns student ids for the given class ids.
func (r *ExamParticipantRepository) StudentIDsByClasses(ctx context.Context, classIDs []uuid.UUID) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := r.db.WithContext(ctx).Model(&model.Student{}).
		Where("class_id IN ?", classIDs).
		Pluck("id", &ids).Error
	return ids, err
}

// AssignedStudentIDsByExam returns the set of student ids already assigned.
func (r *ExamParticipantRepository) AssignedStudentIDsByExam(ctx context.Context, examID uuid.UUID) (map[uuid.UUID]bool, error) {
	var ids []uuid.UUID
	err := r.db.WithContext(ctx).Model(&model.ExamParticipant{}).
		Where("exam_id = ?", examID).
		Pluck("student_id", &ids).Error
	if err != nil {
		return nil, err
	}
	set := make(map[uuid.UUID]bool, len(ids))
	for _, id := range ids {
		set[id] = true
	}
	return set, nil
}

// Assign inserts participants skipping existing (exam_id, student_id) pairs.
func (r *ExamParticipantRepository) Assign(ctx context.Context, examID uuid.UUID, studentIDs []uuid.UUID, via string) (int, error) {
	var assigned int
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, sid := range studentIDs {
			res := tx.Where(model.ExamParticipant{ExamID: examID, StudentID: sid}).
				FirstOrCreate(&model.ExamParticipant{ExamID: examID, StudentID: sid, AssignedVia: via})
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected > 0 {
				assigned++
			}
		}
		return nil
	})
	if err != nil {
		return 0, TranslateDBError(err, "")
	}
	return assigned, nil
}

func (r *ExamParticipantRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.ExamParticipant, error) {
	var p model.ExamParticipant
	err := r.db.WithContext(ctx).First(&p, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *ExamParticipantRepository) Delete(ctx context.Context, id uuid.UUID) error {
	res := r.db.WithContext(ctx).Delete(&model.ExamParticipant{}, "id = ?", id)
	if res.Error != nil {
		return TranslateDBError(res.Error, "")
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// StudentsExist validates all student ids exist.
func (r *ExamParticipantRepository) StudentsExist(ctx context.Context, ids []uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Student{}).
		Where("id IN ?", ids).
		Count(&count).Error
	return int(count) == len(ids), err
}

// RemoveWithCleanup deletes the participant together with all their attempts
// (answers cascade via exam_attempts FK) so they can retake the exam fresh.
func (r *ExamParticipantRepository) RemoveWithCleanup(ctx context.Context, examID, participantID, studentID uuid.UUID) error {
	return TranslateDBError(r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("exam_id = ? AND student_id = ?", examID, studentID).
			Delete(&model.ExamAttempt{}).Error; err != nil {
			return err
		}
		res := tx.Delete(&model.ExamParticipant{}, "id = ?", participantID)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	}), "")
}
