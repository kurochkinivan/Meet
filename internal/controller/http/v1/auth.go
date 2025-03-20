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
	AuthenticateEmail(ctx context.Context, email, password string) (*entity.User, error)
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
		UUID      uuid.UUID          `json:"uuid"`
		Name      string             `json:"name"`
		Email     string             `json:"email"`
		Location  entity.Coordiantes `json:"location"`
		CreatedAt time.Time          `json:"created_at"`
	}
)

func (h *AuthHandler) register(w http.ResponseWriter, r *http.Request, p httprouter.Params) error {
	var req registerReq
	err := json.NewDecoder(io.LimitReader(r.Body, h.bytesLimit)).Decode(&req)
	if err != nil {
		return apperr.WithHTTPStatus(err, http.StatusBadRequest)
	}
	defer r.Body.Close()

	user, err := h.AuthUseCase.Register(r.Context(), &entity.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
		Location: req.Location.Coordinates,
	})
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(registerResp{
		UUID:      user.UUID,
		Name:      user.Name,
		Email:     user.Email,
		Location:  user.Location,
		CreatedAt: user.CreatedAt,
	})
	if err != nil {
		return apperr.WithHTTPStatus(err, http.StatusInternalServerError)
	}

	return nil
}

type (
	loginReq struct {
		OAuthToken string `json:"oauth_token"`
		Email      string `json:"email"`
		Password   string `json:"password"`
	}

	loginResp struct {
		UUID      uuid.UUID          `json:"uuid"`
		Name      string             `json:"name"`
		Email     string             `json:"email"`
		Location  entity.Coordiantes `json:"location"`
		CreatedAt time.Time          `json:"created_at"`
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

	hasCreds := req.Email != "" || req.Password != ""
	hasToken := req.OAuthToken != ""
	if hasCreds && hasToken {
		return errors.New("either OAuthToken or Email/Password should be provided, not both")
	} else if !hasToken && (req.Email == "" || req.Password == "") {
		return errors.New("either OAuthToken or Email/Password must be provided")
	}

	var user *entity.User
	if hasToken {
		user, err = h.AuthUseCase.AuthenticateOAuth(r.Context(), req.OAuthToken)
		if err != nil {
			return err
		}
	} else {
		user, err = h.AuthUseCase.AuthenticateEmail(r.Context(), req.Email, req.Password)
		if err != nil {
			return err
		}
	}

	err = json.NewEncoder(w).Encode(&loginResp{
		UUID:      user.UUID,
		Name:      user.Name,
		Email:     user.Email,
		Location:  user.Location,
		CreatedAt: user.CreatedAt,
	})
	if err != nil {
		return apperr.WithHTTPStatus(err, http.StatusInternalServerError)
	}

	return nil
}
