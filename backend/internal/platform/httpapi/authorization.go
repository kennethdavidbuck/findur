package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
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

type sessionLifecycle interface {
	Authenticate(context.Context, string) (auth.Actor, error)
	AuthorizeUnsafe(context.Context, string, string) (auth.Actor, error)
	RevokeCurrent(context.Context, string) error
}

type callbackCookies struct{ attempt, session string }
type callbackCookieKey struct{}
type httpRequestContextKey struct{}

func requestFromContext(ctx context.Context) *http.Request {
	request, _ := ctx.Value(httpRequestContextKey{}).(*http.Request)
	return request
}

const (
	authorizationRequestLimit = 4 << 10
	attemptCookieName         = "findur_oauth_attempt"
	sessionCookieName         = "findur_session"
	csrfCookieName            = "findur_csrf"
	authorizationPath         = "/api/auth/snaptrade/authorize"
	authorizationStatusPath   = "/api/auth/status"
	callbackSucceededCategory = "succeeded"
	callbackRestartCategory   = "restart_required"
	requestCanceledCategory   = "request_canceled"
	deadlineExceededCategory  = "deadline_exceeded"
	sessionCheckCategory      = "session_check_failed"
	cacheControlHeader        = "Cache-Control"
	noStoreDirective          = "no-store"
	privateNoStoreDirective   = "private, no-store"
	contentTypeHeader         = "Content-Type"
	jsonMediaType             = "application/json"
	formMediaType             = "application/x-www-form-urlencoded"
)

type authorizationAPI struct {
	logger                 *slog.Logger
	initiator              authorizationInitiator
	completer              authorizationCompleter
	status                 authorizationStatusProvider
	sessions               sessionLifecycle
	authorizationAvailable bool
	publicOrigin           string
}

func registerAuthorizationAPI(mux *http.ServeMux, logger *slog.Logger, initiator authorizationInitiator, completer authorizationCompleter, sessions sessionLifecycle, authorizationAvailable bool, publicOrigin string) {
	api := &authorizationAPI{logger: logger, initiator: initiator, completer: completer, sessions: sessions, authorizationAvailable: authorizationAvailable, publicOrigin: publicOrigin}
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
	spec, err := generated.GetSpec()
	if err != nil {
		panic("generated OpenAPI specification is invalid")
	}
	validateRequests := requestvalidator.OapiRequestValidatorWithOptions(spec, &requestvalidator.Options{
		ErrorHandlerWithOpts: func(_ context.Context, _ error, w http.ResponseWriter, _ *http.Request, _ requestvalidator.ErrorHandlerOpts) {
			writeGeneratedError(w, http.StatusBadRequest, generated.InvalidRequest)
		},
	})
	generated.HandlerWithOptions(strict, generated.StdHTTPServerOptions{
		BaseRouter: mux,
		ErrorHandlerFunc: func(w http.ResponseWriter, _ *http.Request, _ error) {
			writeGeneratedError(w, http.StatusBadRequest, generated.InvalidRequest)
		},
		Middlewares: []generated.MiddlewareFunc{
			validateRequests,
			noStoreMiddleware,
			callbackContextMiddleware,
			contentTypeMiddleware,
		},
	})
	mux.HandleFunc(authorizationPath, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Allow", http.MethodPost)
		writeGeneratedError(w, http.StatusMethodNotAllowed, generated.InvalidRequest)
	})
}

func noStoreMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(cacheControlHeader, noStoreDirective)
		next.ServeHTTP(w, r)
	})
}

func callbackContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == auth.SnapTradeCallbackPath || r.URL.Path == authorizationStatusPath || r.URL.Path == "/api/auth/logout" {
			cookies := callbackCookies{
				attempt: cookieValue(r, attemptCookieName),
				session: cookieValue(r, sessionCookieName),
			}
			r = r.WithContext(context.WithValue(r.Context(), callbackCookieKey{}, cookies))
		}
		r.Body = http.MaxBytesReader(w, r.Body, authorizationRequestLimit)
		r = r.WithContext(context.WithValue(r.Context(), httpRequestContextKey{}, r))
		next.ServeHTTP(w, r)
	})
}

func contentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentType := strings.TrimSpace(strings.Split(r.Header.Get(contentTypeHeader), ";")[0])
		if contentType != "" && contentType != jsonMediaType && contentType != formMediaType {
			writeGeneratedError(w, http.StatusUnsupportedMediaType, generated.InvalidRequest)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (a *authorizationAPI) GetAuthorizationStatus(ctx context.Context, _ generated.GetAuthorizationStatusRequestObject) (generated.GetAuthorizationStatusResponseObject, error) {
	cookies, _ := ctx.Value(callbackCookieKey{}).(callbackCookies)
	status := auth.AuthorizationStatus{AuthorizationAvailable: a.authorizationAvailable}
	if a.sessions != nil {
		_, err := a.sessions.Authenticate(ctx, cookies.session)
		status.Authenticated = err == nil
		if err != nil && !errors.Is(err, auth.ErrUnauthenticated) {
			a.logger.WarnContext(ctx, "authorization status unavailable", "request_id", requestIDFromContext(ctx), "category", sessionCheckCategory)
		}
	} else if a.status != nil {
		var err error
		status, err = a.status.Status(ctx, cookies.session)
		if err != nil {
			a.logger.WarnContext(ctx, "authorization status unavailable", "request_id", requestIDFromContext(ctx), "category", sessionCheckCategory)
		}
	}
	return generated.GetAuthorizationStatus200JSONResponse{Body: generated.AuthorizationStatus{AuthorizationAvailable: status.AuthorizationAvailable, Authenticated: status.Authenticated}, Headers: generated.GetAuthorizationStatus200ResponseHeaders{CacheControl: noStoreDirective}}, nil
}

type logoutResponse struct{ cookies []*http.Cookie }

func (r logoutResponse) VisitLogoutCurrentSessionResponse(w http.ResponseWriter) error {
	for _, cookie := range r.cookies {
		http.SetCookie(w, cookie)
	}
	w.Header().Set(cacheControlHeader, privateNoStoreDirective)
	w.WriteHeader(http.StatusNoContent)
	return nil
}

type logoutErrorResponse struct {
	status  int
	code    generated.ErrorCode
	cookies []*http.Cookie
}

func (r logoutErrorResponse) VisitLogoutCurrentSessionResponse(w http.ResponseWriter) error {
	for _, cookie := range r.cookies {
		http.SetCookie(w, cookie)
	}
	w.Header().Set(cacheControlHeader, privateNoStoreDirective)
	w.Header().Set(contentTypeHeader, jsonMediaType)
	w.WriteHeader(r.status)
	return json.NewEncoder(w).Encode(generated.Error{Code: r.code})
}

func (a *authorizationAPI) LogoutCurrentSession(ctx context.Context, request generated.LogoutCurrentSessionRequestObject) (generated.LogoutCurrentSessionResponseObject, error) {
	cookies, _ := ctx.Value(callbackCookieKey{}).(callbackCookies)
	if a.sessions == nil || cookies.session == "" {
		return logoutUnauthorized(), nil
	}
	httpRequest := requestFromContext(ctx)
	if httpRequest == nil || !sameOriginRequest(httpRequest, a.publicOrigin) || httpRequest.Header.Get("Sec-Fetch-Site") != "same-origin" {
		return logoutForbidden(), nil
	}
	if _, err := a.sessions.AuthorizeUnsafe(ctx, cookies.session, optionalString(request.Params.XCSRFToken)); err != nil {
		switch {
		case errors.Is(err, auth.ErrUnauthenticated):
			return logoutUnauthorized(), nil
		case errors.Is(err, auth.ErrForbidden):
			return logoutForbidden(), nil
		default:
			return nil, err
		}
	}
	if err := a.sessions.RevokeCurrent(ctx, cookies.session); err != nil {
		if errors.Is(err, auth.ErrUnauthenticated) {
			return logoutUnauthorized(), nil
		}
		return nil, err
	}
	return logoutResponse{cookies: expiredSessionCookies()}, nil
}

func logoutUnauthorized() generated.LogoutCurrentSessionResponseObject {
	return logoutErrorResponse{status: http.StatusUnauthorized, code: generated.Unauthenticated, cookies: expiredSessionCookies()}
}

func logoutForbidden() generated.LogoutCurrentSessionResponseObject {
	return logoutErrorResponse{status: http.StatusForbidden, code: generated.Forbidden}
}

func sameOriginRequest(request *http.Request, publicOrigin string) bool {
	if publicOrigin == "" {
		return false
	}
	if origin := request.Header.Get("Origin"); origin != "" {
		return origin == publicOrigin
	}
	referer, err := url.Parse(request.Header.Get("Referer"))
	return err == nil && referer.Scheme+"://"+referer.Host == publicOrigin
}

type callbackRedirect struct {
	location string
	cookies  []*http.Cookie
}

func (r callbackRedirect) VisitCompleteSnapTradeAuthorizationResponse(w http.ResponseWriter) error {
	for _, cookie := range r.cookies {
		http.SetCookie(w, cookie)
	}
	w.Header().Set(cacheControlHeader, noStoreDirective)
	w.Header().Set("Location", r.location)
	w.WriteHeader(http.StatusSeeOther)
	return nil
}

func (a *authorizationAPI) CompleteSnapTradeAuthorization(ctx context.Context, request generated.CompleteSnapTradeAuthorizationRequestObject) (generated.CompleteSnapTradeAuthorizationResponseObject, error) {
	cookies, _ := ctx.Value(callbackCookieKey{}).(callbackCookies)
	result := auth.CallbackResult{Route: auth.AuthorizationResultRoute}
	var err error
	if a.completer != nil {
		result, err = a.completer.Complete(ctx, callbackInput(request.Params, cookies))
	}
	set := []*http.Cookie{expiredAttemptCookie()}
	if err == nil && result.Success && result.Session != "" {
		set = append(set, sessionCookies(result)...)
	}
	category := callbackRestartCategory
	if err == nil && result.Success {
		category = callbackSucceededCategory
	}
	a.logger.InfoContext(ctx, "authorization callback completed", "request_id", requestIDFromContext(ctx), "category", category)
	return callbackRedirect{location: result.Route, cookies: set}, nil
}

func (a *authorizationAPI) BeginSnapTradeAuthorization(ctx context.Context, request generated.BeginSnapTradeAuthorizationRequestObject) (generated.BeginSnapTradeAuthorizationResponseObject, error) {
	if a.initiator == nil {
		return unavailableResponse(generated.AuthorizationUnavailable), nil
	}
	result, err := a.initiator.Begin(ctx, requestedReturn(request))
	if err != nil {
		code := generated.InitializationFailed
		category := string(code)
		switch {
		case errors.Is(err, auth.ErrUnavailable):
			code = generated.AuthorizationUnavailable
			category = string(code)
		case errors.Is(err, context.Canceled):
			category = requestCanceledCategory
		case errors.Is(err, context.DeadlineExceeded):
			category = deadlineExceededCategory
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
	cookie := (&http.Cookie{Name: attemptCookieName, Value: result.BrowserBinding, Path: auth.SnapTradeCallbackPath, MaxAge: maxAge, Secure: true, HttpOnly: true, SameSite: http.SameSiteLaxMode}).String()
	return generated.BeginSnapTradeAuthorization303Response{Headers: generated.BeginSnapTradeAuthorization303ResponseHeaders{Location: result.AuthorizationURL, SetCookie: cookie}}, nil
}

func callbackInput(params generated.CompleteSnapTradeAuthorizationParams, cookies callbackCookies) auth.CallbackInput {
	return auth.CallbackInput{
		State:           optionalString(params.State),
		Code:            optionalString(params.Code),
		ProviderError:   optionalString(params.Error),
		Binding:         cookies.attempt,
		ExistingSession: cookies.session,
	}
}

func optionalString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func requestedReturn(request generated.BeginSnapTradeAuthorizationRequestObject) string {
	if request.JSONBody != nil {
		return optionalReturn(request.JSONBody.ReturnTo)
	}
	if request.FormdataBody != nil {
		return optionalReturn(request.FormdataBody.ReturnTo)
	}
	return ""
}

func optionalReturn(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func cookieValue(request *http.Request, name string) string {
	cookie, err := request.Cookie(name)
	if err != nil {
		return ""
	}
	return cookie.Value
}

func expiredAttemptCookie() *http.Cookie {
	return &http.Cookie{Name: attemptCookieName, Path: auth.SnapTradeCallbackPath, MaxAge: -1, Secure: true, HttpOnly: true, SameSite: http.SameSiteLaxMode}
}

func sessionCookies(result auth.CallbackResult) []*http.Cookie {
	maxAge := int(auth.SessionAbsoluteLifetime / time.Second)
	return []*http.Cookie{
		{Name: sessionCookieName, Value: result.Session, Path: "/", MaxAge: maxAge, Secure: true, HttpOnly: true, SameSite: http.SameSiteLaxMode},
		{Name: csrfCookieName, Value: result.CSRF, Path: "/", MaxAge: maxAge, Secure: true, SameSite: http.SameSiteStrictMode},
	}
}

func expiredSessionCookies() []*http.Cookie {
	return []*http.Cookie{
		{Name: sessionCookieName, Path: "/", MaxAge: -1, Secure: true, HttpOnly: true, SameSite: http.SameSiteLaxMode},
		{Name: csrfCookieName, Path: "/", MaxAge: -1, Secure: true, SameSite: http.SameSiteStrictMode},
	}
}

func unavailableResponse(code generated.ErrorCode) generated.BeginSnapTradeAuthorization503JSONResponse {
	return generated.BeginSnapTradeAuthorization503JSONResponse{Body: generated.Error{Code: code}, Headers: generated.BeginSnapTradeAuthorization503ResponseHeaders{CacheControl: noStoreDirective}}
}

func writeGeneratedError(w http.ResponseWriter, status int, code generated.ErrorCode) {
	w.Header().Set(cacheControlHeader, noStoreDirective)
	w.Header().Set(contentTypeHeader, jsonMediaType)
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(generated.Error{Code: code})
}
