package usecase

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/kurochkinivan/Meet/internal/apperr"
	"github.com/kurochkinivan/Meet/internal/entity"
	"golang.org/x/sync/errgroup"
)

type PhotoUseCase struct {
	PhotoStorage
	PhotoCloud
	PhotoCache
	photoLimit int
}

func NewPhotoUseCase(storage PhotoStorage, cloud PhotoCloud, cache PhotoCache, photoLimit int) *PhotoUseCase {
	return &PhotoUseCase{
		PhotoStorage: storage,
		PhotoCloud:   cloud,
		PhotoCache:   cache,
		photoLimit:   photoLimit,
	}
}

type PhotoStorage interface {
	CreatePhoto(ctx context.Context, userID string, url string, objectKey string) error
	GetPhotos(ctx context.Context, userID string) ([]*entity.Photo, error)
	GetPhoto(ctx context.Context, photoID string) (*entity.Photo, error)
	DeletePhoto(ctx context.Context, userID string, photoID string) error
}

type PhotoCloud interface {
	UploadPhoto(ctx context.Context, userID string, file io.Reader) (url string, objectKey string, err error)
	DeletePhoto(ctx context.Context, objectKey string) error
}

type PhotoCache interface {
	Delete(ctx context.Context, userID string) error
}

func (u *PhotoUseCase) UploadPhotos(ctx context.Context, userID string, files []*multipart.FileHeader) error {
	photos, err := u.PhotoStorage.GetPhotos(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get all photos, err: %w", err)
	}

	if len(photos)+len(files) > u.photoLimit {
		return apperr.WithHTTPStatus(errors.New("photo limit exceeded"), http.StatusBadRequest)
	}

	erg, ctx := errgroup.WithContext(ctx)
	erg.SetLimit(10)

	for _, file := range files {
		erg.Go(func() error {
			f, err := file.Open()
			if err != nil {
				return apperr.WithHTTPStatus(fmt.Errorf("failed to open file, err: %w", err), http.StatusInternalServerError)
			}
			defer f.Close()

			url, objectKey, err := u.PhotoCloud.UploadPhoto(ctx, userID, f)
			if err != nil {
				return fmt.Errorf("failed to upload photo, err: %w", err)
			}

			err = u.PhotoStorage.CreatePhoto(ctx, userID, url, objectKey)
			if err != nil {
				errDelete := u.PhotoCloud.DeletePhoto(ctx, objectKey)
				if errDelete != nil {
					return fmt.Errorf("failed to create photo: %w; rollback failed: %v", err, errDelete)
				}
				return fmt.Errorf("failed to create photo, rollback cloud upload, err: %w", err)
			}

			return nil
		})
	}

	err = u.PhotoCache.Delete(ctx, userID)
	if err != nil {
		return err
	}

	return erg.Wait()
}

func (u *PhotoUseCase) GetPhotos(ctx context.Context, userID string) ([]*entity.Photo, error) {
	photos, err := u.PhotoStorage.GetPhotos(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get all photos, err: %w", err)
	}

	return photos, nil
}

func (u *PhotoUseCase) DeletePhoto(ctx context.Context, userID string, photoID string) error {
	photo, err := u.PhotoStorage.GetPhoto(ctx, photoID)
	if err != nil {
		if errors.Is(err, apperr.ErrNoRows) {
			return nil
		}
		return fmt.Errorf("failed to get photo, err: %w", err)
	}

	err = u.PhotoStorage.DeletePhoto(ctx, userID, photoID)
	if err != nil {
		return fmt.Errorf("failed to delete photo, err: %w", err)
	}

	err = u.PhotoCloud.DeletePhoto(ctx, photo.ObjectKey)
	if err != nil {
		errCreate := u.PhotoStorage.CreatePhoto(ctx, photo.UserID.String(), photo.URL, photo.ObjectKey)
		if errCreate != nil {
			return fmt.Errorf("failed to delete photo from cloud: %w, rollback failed: %w", err, errCreate)
		}
		return fmt.Errorf("failed to delete photo from cloud, rollback db delete, err: %w", err)
	}

	err = u.PhotoCache.Delete(ctx, userID)
	if err != nil {
		return err
	}

	return nil
}
