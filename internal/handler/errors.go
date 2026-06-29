package handler

import (
	"errors"
	"net/http"
)

type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

func (h HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h(w, r)
	if err != nil {
		if httpError, ok := errors.AsType[HTTPError](err); ok {
			w.WriteHeader(httpError.Code)
		}
	}
}

type HTTPError struct {
	Code int
	Msg  string
	Err  error
}

func (e HTTPError) Error() string {
	return e.Msg
}
