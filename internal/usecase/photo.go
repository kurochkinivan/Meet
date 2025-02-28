package usecase

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/kurochkinivan/Meet/internal/apperr"
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
	CreatePhoto(ctx context.Context, userID string, url string) error
}

type PhotoCloudRepository interface {
	UploadPhoto(ctx context.Context, userID string, file io.Reader) (string, error)
}

func (u *PhotoUseCase) UploadUserPhotos(ctx context.Context, userID string, files []*multipart.FileHeader) error {
	erg, ctx := errgroup.WithContext(ctx)
	erg.SetLimit(10)

	for _, file := range files {
		erg.Go(func() error {
			f, err := file.Open()
			if err != nil {
				return apperr.WithHTTPStatus(fmt.Errorf("failed to open file, err: %w", err), http.StatusInternalServerError)
			}
			defer f.Close()

			url, err := u.PhotoCloudRepository.UploadPhoto(ctx, userID, f)
			if err != nil {
				return fmt.Errorf("failed to upload photo, err: %w", err)
			}

			err = u.PhotoStorageRepository.CreatePhoto(ctx, userID, url)
			if err != nil {
				return fmt.Errorf("failed to create photo, err: %w", err)
			}

			return nil
		})
	}

	return erg.Wait()
}
