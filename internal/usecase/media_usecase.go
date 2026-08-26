package usecase

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/asepzainudin14/mcbt/internal/model"
	"github.com/asepzainudin14/mcbt/internal/pkg/apperror"
	"github.com/asepzainudin14/mcbt/internal/pkg/storage"
)

type MediaUsecase struct {
	repo     MediaRepo
	storage  *storage.Client
	maxBytes int64
}

func NewMediaUsecase(repo MediaRepo, st *storage.Client, maxBytes int64) *MediaUsecase {
	return &MediaUsecase{repo: repo, storage: st, maxBytes: maxBytes}
}

var validMediaTypes = map[string]string{
	"QUESTION_IMAGE": "question",
	"OPTION_IMAGE":   "option",
}

func (u *MediaUsecase) Upload(ctx context.Context, fileHeader *multipart.FileHeader, mediaType string, uploaderID uuid.UUID) (*model.Media, error) {
	folder, ok := validMediaTypes[mediaType]
	if !ok {
		return nil, apperror.New(422, "type harus QUESTION_IMAGE atau OPTION_IMAGE", nil)
	}

	mime := fileHeader.Header.Get("Content-Type")
	if !storage.IsAllowedImageType(mime) {
		return nil, apperror.New(422, "File harus gambar PNG, JPG, GIF, atau WebP", fmt.Errorf("rejected mime %s", mime))
	}
	if fileHeader.Size > u.maxBytes {
		return nil, apperror.New(422, fmt.Sprintf("Ukuran file maksimal %d MB", u.maxBytes>>20), nil)
	}

	f, err := fileHeader.Open()
	if err != nil {
		return nil, apperror.BadRequest("Gagal membaca file", err)
	}
	defer f.Close()

	key, size, err := u.storage.PutImage(ctx, f, mime, folder)
	if err != nil {
		return nil, apperror.Internal(err)
	}

	media := &model.Media{
		UploadedBy: &uploaderID,
		FileName:   fileHeader.Filename,
		FilePath:   key,
		MimeType:   mime,
		FileSize:   size,
	}
	if err := u.repo.Create(ctx, media); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("Media tidak ditemukan", err)
		}
		return nil, err
	}
	return media, nil
}

func (u *MediaUsecase) GetByID(ctx context.Context, id uuid.UUID) (*model.Media, error) {
	media, err := u.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("Media tidak ditemukan", err)
		}
		return nil, apperror.Internal(err)
	}
	return media, nil
}

func (u *MediaUsecase) GetFile(ctx context.Context, key string) (io.ReadCloser, string, int64, error) {
	return u.storage.GetImage(ctx, key)
}
