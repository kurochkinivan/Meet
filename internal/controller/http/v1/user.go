package v1

import (
	"context"
	"mime/multipart"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/kurochkinivan/Meet/internal/apperr"
)

type PhotoUseCase interface {
	UploadUserPhotos(ctx context.Context, userID string, files []*multipart.FileHeader) error
}

type UserHandler struct {
	PhotoUseCase
	bytesLimit int64
	maxMemory  int64
}

func NewUserHandler(bytesLimit int64, maxMemory int64, photoUseCase PhotoUseCase) Handler {
	return &UserHandler{
		PhotoUseCase: photoUseCase,
		bytesLimit:   bytesLimit,
		maxMemory:    maxMemory,
	}
}

func (h *UserHandler) Register(r *httprouter.Router) {
	r.POST("/v1/users/:id/photos", errorHandler(h.uploadPhotos))
}

func (h *UserHandler) uploadPhotos(w http.ResponseWriter, r *http.Request, p httprouter.Params) error {
	err := r.ParseMultipartForm(h.maxMemory)
	if err != nil {
		return apperr.WithHTTPStatus(err, http.StatusBadRequest)
	}

	userID := p.ByName("id")
	files := r.MultipartForm.File["photo"]

	err = h.PhotoUseCase.UploadUserPhotos(r.Context(), userID, files)
	if err != nil {
		return err
	}

	return nil
}
