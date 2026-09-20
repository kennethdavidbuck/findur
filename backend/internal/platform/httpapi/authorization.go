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

type authorizationCompleter interface {
	Complete(context.Context, auth.CallbackInput) (auth.CallbackResult, error)
}

type authorizationStatusProvider interface {
	Status(context.Context, string) (auth.AuthorizationStatus, error)
}

type callbackCookies struct{ attempt, session string }
type callbackCookieKey struct{}

const authorizationRequestLimit = 4 << 10

type authorizationAPI struct {
	logger    *slog.Logger
	initiator authorizationInitiator
	completer authorizationCompleter
	status    authorizationStatusProvider
}

func registerAuthorizationAPI(mux *http.ServeMux, logger *slog.Logger, initiator authorizationInitiator, completer authorizationCompleter) {
	api := &authorizationAPI{logger: logger, initiator: initiator, completer: completer}
	if provider, ok := completer.(authorizationStatusProvider); ok {
		api.status = provider
	}
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
				if r.URL.Path == "/api/auth/snaptrade/callback" || r.URL.Path == "/api/auth/status" {
					cookies := callbackCookies{}
					if cookie, err := r.Cookie("findur_oauth_attempt"); err == nil {
						cookies.attempt = cookie.Value
					}
					if cookie, err := r.Cookie("findur_session"); err == nil {
						cookies.session = cookie.Value
					}
					r = r.WithContext(context.WithValue(r.Context(), callbackCookieKey{}, cookies))
				}
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

func (a *authorizationAPI) GetAuthorizationStatus(ctx context.Context, _ generated.GetAuthorizationStatusRequestObject) (generated.GetAuthorizationStatusResponseObject, error) {
	cookies, _ := ctx.Value(callbackCookieKey{}).(callbackCookies)
	status := auth.AuthorizationStatus{}
	if a.status != nil {
		var err error
		status, err = a.status.Status(ctx, cookies.session)
		if err != nil {
			a.logger.WarnContext(ctx, "authorization status unavailable", "request_id", requestIDFromContext(ctx), "category", "session_check_failed")
		}
	}
	return generated.GetAuthorizationStatus200JSONResponse{Body: generated.AuthorizationStatus{AuthorizationAvailable: status.AuthorizationAvailable, Authenticated: status.Authenticated}, Headers: generated.GetAuthorizationStatus200ResponseHeaders{CacheControl: "no-store"}}, nil
}

type callbackRedirect struct {
	location string
	cookies  []*http.Cookie
}

func (r callbackRedirect) VisitCompleteSnapTradeAuthorizationResponse(w http.ResponseWriter) error {
	for _, cookie := range r.cookies {
		http.SetCookie(w, cookie)
	}
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Location", r.location)
	w.WriteHeader(http.StatusSeeOther)
	return nil
}

func (a *authorizationAPI) CompleteSnapTradeAuthorization(ctx context.Context, request generated.CompleteSnapTradeAuthorizationRequestObject) (generated.CompleteSnapTradeAuthorizationResponseObject, error) {
	cookies, _ := ctx.Value(callbackCookieKey{}).(callbackCookies)
	value := func(input *string) string {
		if input == nil {
			return ""
		}
		return *input
	}
	result := auth.CallbackResult{Route: "/connect/result"}
	var err error
	if a.completer != nil {
		result, err = a.completer.Complete(ctx, auth.CallbackInput{State: value(request.Params.State), Code: value(request.Params.Code), ProviderError: value(request.Params.Error), Binding: cookies.attempt, ExistingSession: cookies.session})
	}
	expireAttempt := &http.Cookie{Name: "findur_oauth_attempt", Value: "", Path: "/api/auth/snaptrade/callback", MaxAge: -1, Secure: true, HttpOnly: true, SameSite: http.SameSiteLaxMode}
	set := []*http.Cookie{expireAttempt}
	if err == nil && result.Success && result.Session != "" {
		set = append(set,
			&http.Cookie{Name: "findur_session", Value: result.Session, Path: "/", MaxAge: int(auth.SessionAbsoluteLifetime / time.Second), Secure: true, HttpOnly: true, SameSite: http.SameSiteLaxMode},
			&http.Cookie{Name: "findur_csrf", Value: result.CSRF, Path: "/", MaxAge: int(auth.SessionAbsoluteLifetime / time.Second), Secure: true, HttpOnly: false, SameSite: http.SameSiteStrictMode})
	}
	category := "restart_required"
	if err == nil && result.Success {
		category = "succeeded"
	}
	a.logger.InfoContext(ctx, "authorization callback completed", "request_id", requestIDFromContext(ctx), "category", category)
	return callbackRedirect{location: result.Route, cookies: set}, nil
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
