package v1

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/kurochkinivan/Meet/internal/usecase"
)

type Handler interface {
	Register(r *httprouter.Router)
}

func NewHandler(usecases *usecase.UseCases, bytesLimit, maxMemory int64) http.Handler {
	r := httprouter.New()

	authHandler := NewAuthHandler(bytesLimit, usecases.AuthUseCase)
	authHandler.Register(r)

	userHandler := NewUserHandler(bytesLimit, maxMemory, usecases.UserUseCase, usecases.PhotoUseCase)
	userHandler.Register(r)

	return r
}
