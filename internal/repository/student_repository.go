package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/asepzainudin14/mcbt/internal/model"
)

type StudentRepository struct {
	db *gorm.DB
}

func NewStudentRepository(db *gorm.DB) *StudentRepository {
	return &StudentRepository{db: db}
}

type StudentUpsert struct {
	UserID       uuid.UUID
	Username     string
	Name         string
	Email        string
	PasswordHash string
	Nis          string
	ClassID      *uuid.UUID
	Phone        *string
	Address      *string
}

func (r *StudentRepository) List(ctx context.Context, search string, classID *uuid.UUID, page, limit int) (*PageResult[model.Student], error) {
	var (
		items []model.Student
		total int64
	)

	q := r.db.WithContext(ctx).Model(&model.Student{}).
		Joins("JOIN users ON users.id = students.user_id")
	if classID != nil {
		q = q.Where("students.class_id = ?", *classID)
	}
	if search != "" {
		pattern := "%" + search + "%"
		q = q.Where(
			"users.name ILIKE ? OR users.email ILIKE ? OR users.username ILIKE ? OR students.nis ILIKE ?",
			pattern, pattern, pattern, pattern,
		)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, err
	}

	err := q.Preload("User").Preload("Class").
		Order("users.name ASC").
		Limit(limit).Offset((page - 1) * limit).
		Find(&items).Error
	if err != nil {
		return nil, err
	}
	return &PageResult[model.Student]{Items: items, TotalItems: total, Page: page, Limit: limit}, nil
}

func (r *StudentRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Student, error) {
	var s model.Student
	err := r.db.WithContext(ctx).Preload("User").Preload("Class").First(&s, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *StudentRepository) userExists(ctx context.Context, username, email string, excludeUserID *uuid.UUID) (bool, error) {
	q := r.db.WithContext(ctx).Model(&model.User{}).Where("username = ? OR email = ?", username, email)
	if excludeUserID != nil {
		q = q.Where("id <> ?", *excludeUserID)
	}
	var count int64
	if err := q.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *StudentRepository) nisExists(ctx context.Context, nis string, excludeUserID *uuid.UUID) (bool, error) {
	q := r.db.WithContext(ctx).Model(&model.Student{}).Where("nis = ?", nis)
	if excludeUserID != nil {
		q = q.Where("user_id <> ?", *excludeUserID)
	}
	var count int64
	if err := q.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *StudentRepository) FindByUserID(ctx context.Context, userID uuid.UUID) (*model.Student, error) {
	var s model.Student
	err := r.db.WithContext(ctx).Preload("User").Preload("Class").First(&s, "user_id = ?", userID).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *StudentRepository) ExistsDuplicate(ctx context.Context, username, email, nis string, excludeUserID *uuid.UUID) (string, bool, error) {
	dup, err := r.userExists(ctx, username, email, excludeUserID)
	if err != nil || dup {
		return "username/email", dup, err
	}
	dup, err = r.nisExists(ctx, nis, excludeUserID)
	if err != nil || dup {
		return "nis", dup, err
	}
	return "", false, nil
}

func (r *StudentRepository) CreateWithUser(ctx context.Context, up StudentUpsert, roleID uuid.UUID) (*model.Student, error) {
	err := TranslateDBError(r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		user := model.User{
			Username:     up.Username,
			Name:         up.Name,
			Email:        up.Email,
			PasswordHash: up.PasswordHash,
		}
		if err := tx.Omit("Roles").Create(&user).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.UserRole{UserID: user.ID, RoleID: roleID}).Error; err != nil {
			return err
		}
		student := model.Student{UserID: user.ID, Nis: up.Nis, ClassID: up.ClassID, Phone: up.Phone, Address: up.Address}
		if err := tx.Create(&student).Error; err != nil {
			return err
		}
		up.UserID = user.ID
		return nil
	}), "Username, email, atau NIS sudah digunakan")
	if err != nil {
		return nil, err
	}
	return r.FindByUserID(ctx, up.UserID)
}

func (r *StudentRepository) CreateManyWithUsers(ctx context.Context, ups []StudentUpsert, roleID uuid.UUID) error {
	return TranslateDBError(r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, up := range ups {
			user := model.User{
				Username:     up.Username,
				Name:         up.Name,
				Email:        up.Email,
				PasswordHash: up.PasswordHash,
			}
			if err := tx.Omit("Roles").Create(&user).Error; err != nil {
				return err
			}
			if err := tx.Create(&model.UserRole{UserID: user.ID, RoleID: roleID}).Error; err != nil {
				return err
			}
			student := model.Student{UserID: user.ID, Nis: up.Nis, ClassID: up.ClassID, Phone: up.Phone, Address: up.Address}
			if err := tx.Create(&student).Error; err != nil {
				return err
			}
		}
		return nil
	}), "Ada data yang duplikat saat import")
}

type StudentUpdate struct {
	Name    string
	Email   string
	Nis     string
	ClassID *uuid.UUID
	Phone   *string
	Address *string
}

func (r *StudentRepository) Update(ctx context.Context, student *model.Student, up StudentUpdate) error {
	return TranslateDBError(r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		updates := map[string]any{"name": up.Name, "email": up.Email, "updated_at": gorm.Expr("now()")}
		if err := tx.Model(&model.User{}).Where("id = ?", student.UserID).Updates(updates).Error; err != nil {
			return err
		}
		student.Nis = up.Nis
		student.ClassID = up.ClassID
		student.Phone = up.Phone
		student.Address = up.Address
		return tx.Model(student).Select("nis", "class_id", "phone", "address", "updated_at").Updates(student).Error
	}), "Username, email, atau NIS sudah digunakan")
}

func (r *StudentRepository) ChangeClass(ctx context.Context, studentID, targetClassID uuid.UUID) (*model.Student, error) {
	err := TranslateDBError(r.db.WithContext(ctx).
		Model(&model.Student{}).
		Where("id = ?", studentID).
		Update("class_id", targetClassID).Error, "")
	if err != nil {
		return nil, err
	}
	return r.FindByID(ctx, studentID)
}

func (r *StudentRepository) UpdatePasswordByUser(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	return r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("id = ?", userID).
		Updates(map[string]any{
			"password_hash": passwordHash,
			"token_version": gorm.Expr("token_version + 1"),
		}).Error
}

func (r *StudentRepository) DeleteWithUser(ctx context.Context, student *model.Student) error {
	return TranslateDBError(r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Delete(&model.User{}, "id = ?", student.UserID)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	}), "")
}
