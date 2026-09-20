package httpapi

import (
	"context"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type fixtureProvider interface {
	Accounts(context.Context) (*http.Response, error)
}

// Diagnostics contains synthetic integration-only handlers. A nil value exposes no routes.
type Diagnostics struct {
	proxy    http.Handler
	provider fixtureProvider
}

// NewDiagnostics creates integration-only diagnostic handlers.
func NewDiagnostics(target *url.URL, provider fixtureProvider) *Diagnostics {
	return newDiagnostics(target, provider, http.DefaultTransport)
}

func newDiagnostics(target *url.URL, provider fixtureProvider, transport http.RoundTripper) *Diagnostics {
	reverseProxy := httputil.NewSingleHostReverseProxy(target)
	reverseProxy.Transport = transport
	reverseProxy.ErrorHandler = func(w http.ResponseWriter, _ *http.Request, _ error) {
		writeStatus(w, http.StatusBadGateway, "unavailable", "fixture")
	}
	return &Diagnostics{
		proxy:    http.StripPrefix("/api/__fixture/proxy", reverseProxy),
		provider: provider,
	}
}

func (d *Diagnostics) register(mux *http.ServeMux) {
	mux.Handle("/api/__fixture/proxy/", d.proxy)
	mux.HandleFunc("GET /api/__fixture/provider/accounts", func(w http.ResponseWriter, r *http.Request) {
		response, err := d.provider.Accounts(r.Context())
		if err != nil {
			writeStatus(w, http.StatusBadGateway, "unavailable", "fixture")
			return
		}
		defer func() { _ = response.Body.Close() }()
		for _, header := range []string{"Content-Type", "Cache-Control", "Retry-After"} {
			if value := response.Header.Get(header); value != "" {
				w.Header().Set(header, value)
			}
		}
		w.WriteHeader(response.StatusCode)
		_, _ = io.Copy(w, response.Body)
	})
}
