package usecase

import (
	"context"
	"mime/multipart"
)

type PhotoUseCase struct {
	PhotoRepository
}

func NewPhotoUseCase(photoRepository PhotoRepository) *PhotoUseCase {
	return &PhotoUseCase{
		PhotoRepository: photoRepository,
	}
}

type PhotoRepository interface {
	CreatePhotos(ctx context.Context, files []*multipart.FileHeader)
}

func (u *PhotoUseCase) UploadUserPhotos(ctx context.Context, userID string, files []*multipart.FileHeader) error {
	

	return nil
}
