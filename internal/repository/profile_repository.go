package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProfileRepository struct {
	db *gorm.DB
}

func NewProfileRepository(db *gorm.DB) *ProfileRepository {
	return &ProfileRepository{db: db}
}

type ProfileRow struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Nis       *string   `json:"nis"`
	ClassName *string   `json:"class_name"`
	Nip       *string   `json:"nip"`
	Phone     *string   `json:"phone"`
}

func (r *ProfileRepository) FindByUserID(ctx context.Context, userID uuid.UUID) (*ProfileRow, error) {
	var row ProfileRow
	err := r.db.WithContext(ctx).
		Raw(`SELECT
			u.id       AS id,
			u.username AS username,
			u.name     AS name,
			u.email    AS email,
			s.nis      AS nis,
			c.name     AS class_name,
			t.nip      AS nip,
			COALESCE(s.phone, t.phone) AS phone
		FROM users u
		LEFT JOIN students s ON s.user_id = u.id
		LEFT JOIN classes c ON c.id = s.class_id
		LEFT JOIN teachers t ON t.user_id = u.id
		WHERE u.id = ?`, userID).
		Scan(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *ProfileRepository) UpdateName(ctx context.Context, userID uuid.UUID, name string) error {
	return r.db.WithContext(ctx).Exec(
		"UPDATE users SET name = ?, updated_at = ? WHERE id = ?",
		name, time.Now(), userID,
	).Error
}

func (r *ProfileRepository) UpdateStudentPhone(ctx context.Context, userID uuid.UUID, phone *string) error {
	return r.db.WithContext(ctx).Exec(
		"UPDATE students SET phone = ?, updated_at = ? WHERE user_id = ?",
		phone, time.Now(), userID,
	).Error
}

func (r *ProfileRepository) UpdateTeacherPhone(ctx context.Context, userID uuid.UUID, phone *string) error {
	return r.db.WithContext(ctx).Exec(
		"UPDATE teachers SET phone = ?, updated_at = ? WHERE user_id = ?",
		phone, time.Now(), userID,
	).Error
}
