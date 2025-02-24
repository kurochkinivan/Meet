package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
	httperrors "github.com/kurochkinivan/Meet/internal/controller/httpErrors"
	"github.com/kurochkinivan/Meet/internal/entity"
)

type AuthUseCase interface {
	Register(ctx context.Context, user *entity.User) (*entity.User, error)
}

type AuthHandler struct {
	AuthUseCase
	bytesLimit int64
}

func NewAuthHandler(bytesLimit int64, authUseCase AuthUseCase) Handler {
	return &AuthHandler{
		AuthUseCase: authUseCase,
		bytesLimit:  bytesLimit,
	}
}

func (h *AuthHandler) Register(r *httprouter.Router) {
	r.POST("/v1/auth/register", h.register)
}

type (
	registerReq struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
		Location struct {
			Type        string             `json:"type"`
			Coordinates entity.Coordiantes `json:"coordinates"`
		} `json:"location"`
	}

	registerResp struct {
		UUID     uuid.UUID          `json:"uuid"`
		Name     string             `json:"name"`
		Email    string             `json:"email"`
		Location entity.Coordiantes `json:"location"`
	}
)

func (h *AuthHandler) register(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var req registerReq
	err := json.NewDecoder(io.LimitReader(r.Body, h.bytesLimit)).Decode(&req)
	if err != nil {
		http.Error(w, fmt.Sprintln(httperrors.ErrSerializeData, err.Error()), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	user, err := h.AuthUseCase.Register(r.Context(), &entity.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
		Location: req.Location.Coordinates,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(registerResp{
		UUID:     user.UUID,
		Name:     user.Name,
		Email:    user.Email,
		Location: user.Location,
	})
	if err != nil {
		http.Error(w, httperrors.ErrSerializeData, http.StatusInternalServerError)
		return
	}
}
