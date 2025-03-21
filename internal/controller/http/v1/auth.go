package v1

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/kurochkinivan/Meet/internal/apperr"
	"github.com/kurochkinivan/Meet/internal/entity"
)

type AuthUseCase interface {
	Register(ctx context.Context, user *entity.User) (*entity.User, error)
	AuthenticatePhone(ctx context.Context, phone, password string) (*entity.User, error)
	AuthenticateOAuth(ctx context.Context, OAuth string) (*entity.User, error)
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

// TODO: make auth middleware and jwt
func (h *AuthHandler) Register(r *httprouter.Router) {
	r.POST("/v1/auth/register", errorHandler(h.register))
	r.POST("/v1/auth/login", errorHandler(h.login))
}

type (
	userCoordinates struct {
		Longitude float64 `json:"longitude"`
		Latitude  float64 `json:"latitude"`
	}
)

type (
	registerReq struct {
		Name     string `json:"name"`
		Birthday string `json:"birthday"`
		Sex      string `json:"sex"`
		Phone    string `json:"phone"`
		Password string `json:"password"`
		Location struct {
			Type        string             `json:"type"`
			Coordinates entity.Coordiantes `json:"coordinates"`
		} `json:"location"`
	}

	registerResp struct {
		UUID      uuid.UUID       `json:"uuid"`
		Name      string          `json:"name"`
		Birthday  string          `json:"birthday"`
		Sex       string          `json:"sex"`
		Phone     string          `json:"phone"`
		Location  userCoordinates `json:"location"`
		CreatedAt string          `json:"created_at"`
	}
)

func (h *AuthHandler) register(w http.ResponseWriter, r *http.Request, p httprouter.Params) error {
	var req registerReq
	err := json.NewDecoder(io.LimitReader(r.Body, h.bytesLimit)).Decode(&req)
	if err != nil {
		return apperr.WithHTTPStatus(err, http.StatusBadRequest)
	}
	defer r.Body.Close()

	birthday, err := time.Parse(time.DateOnly, req.Birthday)
	if err != nil {
		return apperr.WithHTTPStatus(err, http.StatusBadRequest)
	}

	user, err := h.AuthUseCase.Register(r.Context(), &entity.User{
		Name:     req.Name,
		Phone:    req.Phone,
		BirthDay: birthday,
		Sex:      req.Sex,
		Password: req.Password,
		Location: req.Location.Coordinates,
	})
	if err != nil {
		if errors.Is(err, apperr.ErrUserExists) {
			return apperr.WithHTTPStatus(err, http.StatusUnauthorized)
		}
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(registerResp{
		UUID:     user.UUID,
		Name:     user.Name,
		Birthday: user.BirthDay.Format(time.DateOnly),
		Sex:      user.Sex,
		Phone:    user.Phone,
		Location: userCoordinates{
			Longitude: user.Location.Longitude,
			Latitude:  user.Location.Latitude,
		},
		CreatedAt: user.CreatedAt.Format(time.DateTime),
	})
	if err != nil {
		return apperr.WithHTTPStatus(err, http.StatusInternalServerError)
	}

	return nil
}

type (
	loginReq struct {
		OAuthToken string `json:"oauth_token"`
		Phone      string `json:"phone"`
		Password   string `json:"password"`
	}

	loginResp struct {
		UUID      uuid.UUID       `json:"uuid"`
		Name      string          `json:"name"`
		Birthday  string          `json:"birthday"`
		Sex       string          `json:"sex"`
		Phone     string          `json:"phone"`
		Location  userCoordinates `json:"location"`
		CreatedAt string          `json:"created_at"`
	}
)

func (h *AuthHandler) login(w http.ResponseWriter, r *http.Request, p httprouter.Params) error {
	var req loginReq
	err := json.NewDecoder(io.LimitReader(r.Body, h.bytesLimit)).Decode(&req)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return apperr.WithHTTPStatus(apperr.ErrEmptyBody, http.StatusBadRequest)
		}
		return apperr.WithHTTPStatus(err, http.StatusBadRequest)
	}
	defer r.Body.Close()

	hasCreds := req.Phone != "" || req.Password != ""
	hasToken := req.OAuthToken != ""
	if hasCreds && hasToken {
		return errors.New("either OAuthToken or Phone/Password should be provided, not both")
	} else if !hasToken && (req.Phone == "" || req.Password == "") {
		return errors.New("either OAuthToken or Phone/Password must be provided")
	}

	var user *entity.User
	if hasToken {
		user, err = h.AuthUseCase.AuthenticateOAuth(r.Context(), req.OAuthToken)
		if err != nil {
			return err
		}
	} else {
		user, err = h.AuthUseCase.AuthenticatePhone(r.Context(), req.Phone, req.Password)
		if err != nil {
			return err
		}
	}

	err = json.NewEncoder(w).Encode(&loginResp{
		UUID:     user.UUID,
		Name:     user.Name,
		Birthday: user.BirthDay.Format(time.DateOnly),
		Sex:      user.Sex,
		Phone:    user.Phone,
		Location: userCoordinates{
			Longitude: user.Location.Longitude,
			Latitude:  user.Location.Latitude,
		},
		CreatedAt: user.CreatedAt.Format(time.DateTime),
	})
	if err != nil {
		return apperr.WithHTTPStatus(err, http.StatusInternalServerError)
	}

	return nil
}
