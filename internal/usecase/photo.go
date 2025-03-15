package usecase

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/kurochkinivan/Meet/internal/apperr"
	"github.com/kurochkinivan/Meet/internal/entity"
	"golang.org/x/sync/errgroup"
)

type PhotoUseCase struct {
	PhotoStorageRepository
	PhotoCloudRepository
}

func NewPhotoUseCase(storage PhotoStorageRepository, cloud PhotoCloudRepository) *PhotoUseCase {
	return &PhotoUseCase{
		PhotoStorageRepository: storage,
		PhotoCloudRepository:   cloud,
	}
}

type PhotoStorageRepository interface {
	CreatePhoto(ctx context.Context, userID string, url string, objectKey string) error
	GetPhotos(ctx context.Context, userID string) ([]*entity.Photo, error)
	DeletePhoto(ctx context.Context, userID string, photoID string) error
}

type PhotoCloudRepository interface {
	UploadPhoto(ctx context.Context, userID string, file io.Reader) (url string, objectKey string, err error)
}

func (u *PhotoUseCase) UploadPhotos(ctx context.Context, userID string, files []*multipart.FileHeader) error {
	erg, ctx := errgroup.WithContext(ctx)
	erg.SetLimit(10)

	for _, file := range files {
		erg.Go(func() error {
			f, err := file.Open()
			if err != nil {
				return apperr.WithHTTPStatus(fmt.Errorf("failed to open file, err: %w", err), http.StatusInternalServerError)
			}
			defer f.Close()

			url, objectKey, err := u.PhotoCloudRepository.UploadPhoto(ctx, userID, f)
			if err != nil {
				return fmt.Errorf("failed to upload photo, err: %w", err)
			}

			err = u.PhotoStorageRepository.CreatePhoto(ctx, userID, url, objectKey)
			if err != nil {
				return fmt.Errorf("failed to create photo, err: %w", err)
			}

			return nil
		})
	}

	return erg.Wait()
}

func (u *PhotoUseCase) GetPhotos(ctx context.Context, userID string) ([]*entity.Photo, error) {
	photos, err := u.PhotoStorageRepository.GetPhotos(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get all photos, err: %w", err)
	}

	return photos, nil
}

func (u *PhotoUseCase) DeletePhoto(ctx context.Context, userID string, photoID string) error {
	err := u.PhotoStorageRepository.DeletePhoto(ctx, userID, photoID)
	if err != nil {
		return fmt.Errorf("failed to delete photo, err: %w", err)
	}

	return nil 
}