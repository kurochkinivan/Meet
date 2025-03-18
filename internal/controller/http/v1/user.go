package v1

import (
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/kurochkinivan/Meet/internal/apperr"
	"github.com/kurochkinivan/Meet/internal/entity"
)

type PhotoUseCase interface {
	UploadPhotos(ctx context.Context, userID string, files []*multipart.FileHeader) error
	DeletePhoto(ctx context.Context, userID string, photoID string) error
	GetPhotos(ctx context.Context, userID string) ([]*entity.Photo, error)
}

type UserUseCase interface {
	GetUserByID(ctx context.Context, userID string) (*entity.User, error)
}

type UserHandler struct {
	PhotoUseCase
	UserUseCase
	bytesLimit int64
	maxMemory  int64
}

func NewUserHandler(bytesLimit int64, maxMemory int64, userUseCase UserUseCase, photoUseCase PhotoUseCase) Handler {
	return &UserHandler{
		UserUseCase:  userUseCase,
		PhotoUseCase: photoUseCase,
		bytesLimit:   bytesLimit,
		maxMemory:    maxMemory,
	}
}

func (h *UserHandler) Register(r *httprouter.Router) {
	r.GET("/v1/users/:id", errorHandler(h.getUser))
	r.GET("/v1/users/:id/photos", errorHandler(h.getPhotos))
	r.POST("/v1/users/:id/photos", errorHandler(h.uploadPhotos))
	r.DELETE("/v1/users/:id/photo/:photo_id", errorHandler(h.deletePhoto))
}

func (h *UserHandler) uploadPhotos(w http.ResponseWriter, r *http.Request, p httprouter.Params) error {
	err := r.ParseMultipartForm(h.maxMemory)
	if err != nil {
		return apperr.WithHTTPStatus(err, http.StatusBadRequest)
	}

	userID := p.ByName("id")
	files := r.MultipartForm.File["photo"]

	err = h.PhotoUseCase.UploadPhotos(r.Context(), userID, files)
	if err != nil {
		return err
	}

	return nil
}

type (
	getUserResponse struct {
		UUID      uuid.UUID          `json:"uuid"`
		Name      string             `json:"name"`
		Email     string             `json:"email"`
		Location  entity.Coordiantes `json:"location"`
		CreatedAt time.Time          `json:"created_at"`
		Photos    []photoResponse    `json:"photos"`
	}

	photoResponse struct {
		ID  int64  `json:"id"`
		URL string `json:"url"`
	}
)

func (h *UserHandler) getUser(w http.ResponseWriter, r *http.Request, p httprouter.Params) error {
	userID := p.ByName("id")

	user, err := h.UserUseCase.GetUserByID(r.Context(), userID)
	if err != nil {
		return err
	}

	resp := getUserResponse{
		UUID:      user.UUID,
		Name:      user.Name,
		Email:     user.Email,
		Location:  user.Location,
		CreatedAt: user.CreatedAt,
	}

	for _, photo := range user.Photos {
		resp.Photos = append(resp.Photos, photoResponse{
			ID:  photo.ID,
			URL: photo.URL,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		return apperr.WithHTTPStatus(err, http.StatusInternalServerError)
	}

	return nil
}

type (
	getPhotosRepsponse struct {
		Photos []photoResponse `json:"photos"`
		Total  int             `json:"total"`
	}
)

func (h *UserHandler) getPhotos(w http.ResponseWriter, r *http.Request, p httprouter.Params) error {
	userID := p.ByName("id")

	photos, err := h.PhotoUseCase.GetPhotos(r.Context(), userID)
	if err != nil {
		return err
	}

	resp := &getPhotosRepsponse{
		Total: len(photos),
	}
	for _, photo := range photos {
		resp.Photos = append(resp.Photos, photoResponse{
			ID:  photo.ID,
			URL: photo.URL,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		return apperr.WithHTTPStatus(err, http.StatusInternalServerError)
	}

	return nil
}

func (h *UserHandler) deletePhoto(w http.ResponseWriter, r *http.Request, p httprouter.Params) error {
	userID := p.ByName("id")
	photoID := p.ByName("photo_id")

	err := h.PhotoUseCase.DeletePhoto(r.Context(), userID, photoID)
	if err != nil {
		return err
	}

	return nil
}
