package v1

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/kurochkinivan/Meet/internal/apperr"
)

type appHandler func(http.ResponseWriter, *http.Request, httprouter.Params) error

func errorHandler(f appHandler) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		err := f(w, r, p)
		if err != nil {
			http.Error(w, err.Error(), apperr.HTTPStatus(err))
		}
	}
}
