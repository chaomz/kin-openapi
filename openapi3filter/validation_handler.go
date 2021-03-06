package openapi3filter

import (
	"context"
	"net/http"
)

type AuthenticationFunc func(context.Context, *AuthenticationInput) error

func NoopAuthenticationFunc(context.Context, *AuthenticationInput) error { return nil }

var _ AuthenticationFunc = NoopAuthenticationFunc

type ValidationHandler struct {
	Handler            http.Handler
	AuthenticationFunc AuthenticationFunc
	SwaggerFile        string
	ErrorEncoder       ErrorEncoder
	router             *Router
}

func (h *ValidationHandler) Load() error {
	h.router = NewRouter()

	err := h.router.AddSwaggerFromFile(h.SwaggerFile)
	if err != nil {
		return err
	}

	// set defaults
	if h.Handler == nil {
		h.Handler = http.DefaultServeMux
	}
	if h.AuthenticationFunc == nil {
		h.AuthenticationFunc = NoopAuthenticationFunc
	}
	if h.ErrorEncoder == nil {
		h.ErrorEncoder = DefaultErrorEncoder
	}

	return nil
}

func (h *ValidationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h.validateRequest(r)
	if err != nil {
		h.ErrorEncoder(r.Context(), err, w)
		return
	}
	// TODO: validateResponse
	h.Handler.ServeHTTP(w, r)
}

func (h *ValidationHandler) validateRequest(r *http.Request) error {
	// Find route
	route, pathParams, err := h.router.FindRoute(r.Method, r.URL)
	if err != nil {
		return err
	}

	options := &Options{
		AuthenticationFunc: h.AuthenticationFunc,
	}

	// Validate request
	requestValidationInput := &RequestValidationInput{
		Request:    r,
		PathParams: pathParams,
		Route:      route,
		Options:    options,
	}
	err = ValidateRequest(r.Context(), requestValidationInput)
	if err != nil {
		return err
	}

	return nil
}
