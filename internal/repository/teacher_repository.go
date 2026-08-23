package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/asepzainudin14/mcbt/internal/model"
)

type TeacherRepository struct {
	db *gorm.DB
}

func NewTeacherRepository(db *gorm.DB) *TeacherRepository {
	return &TeacherRepository{db: db}
}

type TeacherUpsert struct {
	UserID       uuid.UUID
	Username     string
	Name         string
	Email        string
	PasswordHash string
	Nip          *string
	Phone        *string
	Address      *string
}

func (r *TeacherRepository) List(ctx context.Context, search string, page, limit int) (*PageResult[model.Teacher], error) {
	var (
		items []model.Teacher
		total int64
	)

	q := r.db.WithContext(ctx).Model(&model.Teacher{}).
		Joins("JOIN users ON users.id = teachers.user_id")
	if search != "" {
		pattern := "%" + search + "%"
		q = q.Where(
			"users.name ILIKE ? OR users.email ILIKE ? OR users.username ILIKE ? OR teachers.nip ILIKE ?",
			pattern, pattern, pattern, pattern,
		)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, err
	}

	err := q.Preload("User").
		Order("users.name ASC").
		Limit(limit).Offset((page - 1) * limit).
		Find(&items).Error
	if err != nil {
		return nil, err
	}
	return &PageResult[model.Teacher]{Items: items, TotalItems: total, Page: page, Limit: limit}, nil
}

func (r *TeacherRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Teacher, error) {
	var t model.Teacher
	err := r.db.WithContext(ctx).Preload("User").First(&t, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TeacherRepository) userExists(ctx context.Context, username, email string, excludeUserID *uuid.UUID) (bool, error) {
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

func (r *TeacherRepository) nipExists(ctx context.Context, nip string, excludeUserID *uuid.UUID) (bool, error) {
	q := r.db.WithContext(ctx).Model(&model.Teacher{}).Where("nip = ?", nip)
	if excludeUserID != nil {
		q = q.Where("user_id <> ?", *excludeUserID)
	}
	var count int64
	if err := q.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// ExistsDuplicate returns which identifier collides ("username/email" or "nip").
func (r *TeacherRepository) FindByUserID(ctx context.Context, userID uuid.UUID) (*model.Teacher, error) {
	var t model.Teacher
	err := r.db.WithContext(ctx).Preload("User").First(&t, "user_id = ?", userID).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TeacherRepository) ExistsDuplicate(ctx context.Context, username, email, nip string, excludeUserID *uuid.UUID) (string, bool, error) {
	dup, err := r.userExists(ctx, username, email, excludeUserID)
	if err != nil || dup {
		return "username/email", dup, err
	}
	if nip != "" {
		dup, err := r.nipExists(ctx, nip, excludeUserID)
		if err != nil || dup {
			return "nip", dup, err
		}
	}
	return "", false, nil
}

func (r *TeacherRepository) CreateWithUser(ctx context.Context, up TeacherUpsert, roleID uuid.UUID) (*model.Teacher, error) {
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
		teacher := model.Teacher{UserID: user.ID, Nip: up.Nip, Phone: up.Phone, Address: up.Address}
		if err := tx.Create(&teacher).Error; err != nil {
			return err
		}
		up.UserID = user.ID
		return nil
	}), "Username, email, atau NIP sudah digunakan")
	if err != nil {
		return nil, err
	}
	return r.FindByUserID(ctx, up.UserID)
}

func (r *TeacherRepository) CreateManyWithUsers(ctx context.Context, ups []TeacherUpsert, roleID uuid.UUID) error {
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
			if err := tx.Create(&model.Teacher{UserID: user.ID, Nip: up.Nip, Phone: up.Phone, Address: up.Address}).Error; err != nil {
				return err
			}
		}
		return nil
	}), "Ada data yang duplikat saat import")
}

type TeacherUpdate struct {
	Name    string
	Email   string
	Nip     *string
	Phone   *string
	Address *string
}

func (r *TeacherRepository) Update(ctx context.Context, teacher *model.Teacher, up TeacherUpdate) error {
	return TranslateDBError(r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		updates := map[string]any{"name": up.Name, "email": up.Email, "updated_at": gorm.Expr("now()")}
		if err := tx.Model(&model.User{}).Where("id = ?", teacher.UserID).Updates(updates).Error; err != nil {
			return err
		}
		teacher.Nip = up.Nip
		teacher.Phone = up.Phone
		teacher.Address = up.Address
		return tx.Model(teacher).Select("nip", "phone", "address", "updated_at").Updates(teacher).Error
	}), "Username, email, atau NIP sudah digunakan")
}

func (r *TeacherRepository) DeleteWithUser(ctx context.Context, teacher *model.Teacher) error {
	return TranslateDBError(r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Delete(&model.User{}, "id = ?", teacher.UserID)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	}), "")
}
