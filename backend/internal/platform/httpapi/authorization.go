package httpapi

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/kennethdavidbuck/findur/backend/internal/auth"
	"github.com/kennethdavidbuck/findur/backend/internal/generated"
	requestvalidator "github.com/oapi-codegen/nethttp-middleware"
)

type authorizationInitiator interface {
	Begin(context.Context, string) (auth.BeginResult, error)
}

const authorizationRequestLimit = 4 << 10

type authorizationAPI struct {
	logger    *slog.Logger
	initiator authorizationInitiator
}

func registerAuthorizationAPI(mux *http.ServeMux, logger *slog.Logger, initiator authorizationInitiator) {
	api := &authorizationAPI{logger: logger, initiator: initiator}
	strict := generated.NewStrictHandlerWithOptions(api, nil, generated.StrictHTTPServerOptions{
		RequestErrorHandlerFunc: func(w http.ResponseWriter, _ *http.Request, _ error) {
			writeGeneratedError(w, http.StatusBadRequest, generated.InvalidRequest)
		},
		ResponseErrorHandlerFunc: func(w http.ResponseWriter, _ *http.Request, _ error) {
			writeGeneratedError(w, http.StatusServiceUnavailable, generated.InitializationFailed)
		},
	})
	spec, err := generated.GetSwagger()
	if err != nil {
		panic("generated OpenAPI specification is invalid")
	}
	validate := requestvalidator.OapiRequestValidatorWithOptions(spec, &requestvalidator.Options{
		ErrorHandlerWithOpts: func(_ context.Context, _ error, w http.ResponseWriter, _ *http.Request, _ requestvalidator.ErrorHandlerOpts) {
			writeGeneratedError(w, http.StatusBadRequest, generated.InvalidRequest)
		},
	})
	generated.HandlerWithOptions(strict, generated.StdHTTPServerOptions{BaseRouter: mux, Middlewares: []generated.MiddlewareFunc{
		validate,
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Cache-Control", "no-store")
				next.ServeHTTP(w, r)
			})
		},
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				r.Body = http.MaxBytesReader(w, r.Body, authorizationRequestLimit)
				next.ServeHTTP(w, r)
			})
		},
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				contentType := strings.TrimSpace(strings.Split(r.Header.Get("Content-Type"), ";")[0])
				if contentType != "" && contentType != "application/json" && contentType != "application/x-www-form-urlencoded" {
					writeGeneratedError(w, http.StatusUnsupportedMediaType, generated.InvalidRequest)
					return
				}
				next.ServeHTTP(w, r)
			})
		},
	}})
	mux.HandleFunc("/api/auth/snaptrade/authorize", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Allow", http.MethodPost)
		writeGeneratedError(w, http.StatusMethodNotAllowed, generated.InvalidRequest)
	})
}

func (a *authorizationAPI) BeginSnapTradeAuthorization(ctx context.Context, request generated.BeginSnapTradeAuthorizationRequestObject) (generated.BeginSnapTradeAuthorizationResponseObject, error) {
	if a.initiator == nil {
		return unavailableResponse(generated.AuthorizationUnavailable), nil
	}
	returnTo := ""
	if request.JSONBody != nil && request.JSONBody.ReturnTo != nil {
		returnTo = string(*request.JSONBody.ReturnTo)
	}
	if request.FormdataBody != nil && request.FormdataBody.ReturnTo != nil {
		returnTo = string(*request.FormdataBody.ReturnTo)
	}
	result, err := a.initiator.Begin(ctx, returnTo)
	if err != nil {
		code := generated.InitializationFailed
		category := string(code)
		if errors.Is(err, auth.ErrUnavailable) {
			code = generated.AuthorizationUnavailable
			category = string(code)
		} else if errors.Is(err, context.Canceled) {
			category = "request_canceled"
		} else if errors.Is(err, context.DeadlineExceeded) {
			category = "deadline_exceeded"
		}
		a.logger.WarnContext(ctx, "authorization initiation refused",
			"request_id", requestIDFromContext(ctx),
			"category", category,
			"stage", auth.InitializationStageOf(err),
		)
		return unavailableResponse(code), nil
	}
	maxAge := int(time.Until(result.ExpiresAt) / time.Second)
	if maxAge < 1 || maxAge > int(auth.AttemptLifetime/time.Second) {
		return unavailableResponse(generated.InitializationFailed), nil
	}
	cookie := (&http.Cookie{Name: "findur_oauth_attempt", Value: result.BrowserBinding, Path: "/api/auth/snaptrade/callback", MaxAge: maxAge, Secure: true, HttpOnly: true, SameSite: http.SameSiteLaxMode}).String()
	return generated.BeginSnapTradeAuthorization303Response{Headers: generated.BeginSnapTradeAuthorization303ResponseHeaders{Location: result.AuthorizationURL, SetCookie: cookie}}, nil
}

func unavailableResponse(code generated.ErrorCode) generated.BeginSnapTradeAuthorization503JSONResponse {
	return generated.BeginSnapTradeAuthorization503JSONResponse{Body: generated.Error{Code: code}, Headers: generated.BeginSnapTradeAuthorization503ResponseHeaders{CacheControl: "no-store"}}
}

func writeGeneratedError(w http.ResponseWriter, status int, code generated.ErrorCode) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(`{"code":"` + string(code) + `"}` + "\n"))
}
