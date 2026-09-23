package httpapi

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/kennethdavidbuck/findur/backend/internal/portfolio"
)

type webhookLifecycleStub struct {
	event   portfolio.WebhookEvent
	changed bool
	err     error
	calls   int
}

func (s *webhookLifecycleStub) Apply(_ context.Context, event portfolio.WebhookEvent) (bool, error) {
	s.calls++
	s.event = event
	return s.changed, s.err
}

func TestSnapTradeWebhookAuthenticationAndResponses(t *testing.T) {
	const key = "synthetic-consumer-key"
	valid := validWebhookBody(t, portfolio.WebhookConnectionBroken, "connection", "", time.Now())
	tests := []struct {
		name       string
		body       string
		signature  func(string) string
		serviceErr error
		wantStatus int
		wantCalls  int
	}{
		{name: "valid canonical signature", body: valid, signature: func(body string) string { return webhookSignature(t, key, body) }, wantStatus: http.StatusNoContent, wantCalls: 1},
		{name: "missing signature", body: valid, wantStatus: http.StatusUnauthorized},
		{name: "bad signature", body: valid, signature: func(string) string { return base64.StdEncoding.EncodeToString(make([]byte, sha256.Size)) }, wantStatus: http.StatusUnauthorized},
		{name: "malformed JSON", body: `{`, signature: func(string) string { return "unused" }, wantStatus: http.StatusBadRequest},
		{name: "wrong schema", body: strings.Replace(valid, "oauth_v1", "api_v1", 1), signature: func(body string) string { return webhookSignature(t, key, body) }, wantStatus: http.StatusBadRequest},
		{name: "wrong client", body: strings.Replace(valid, `"client"`, `"other"`, 1), signature: func(body string) string { return webhookSignature(t, key, body) }, wantStatus: http.StatusBadRequest},
		{name: "invalid supported event", body: validWebhookBody(t, portfolio.WebhookConnectionBroken, "", "", time.Now()), signature: func(body string) string { return webhookSignature(t, key, body) }, wantStatus: http.StatusBadRequest},
		{name: "inventory busy", body: valid, signature: func(body string) string { return webhookSignature(t, key, body) }, serviceErr: portfolio.ErrInventoryBusy, wantStatus: http.StatusServiceUnavailable, wantCalls: 1},
		{name: "persistence failure", body: valid, signature: func(body string) string { return webhookSignature(t, key, body) }, serviceErr: errors.New("private database failure"), wantStatus: http.StatusServiceUnavailable, wantCalls: 1},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := &webhookLifecycleStub{changed: true, err: test.serviceErr}
			handler, err := NewSnapTradeWebhookHandler(service, []byte(key), "client", true, slog.New(slog.NewTextHandler(io.Discard, nil)))
			if err != nil {
				t.Fatal(err)
			}
			request := httptest.NewRequest(http.MethodPost, SnapTradeWebhookPath, strings.NewReader(test.body))
			request.Header.Set(contentTypeHeader, jsonMediaType)
			if test.signature != nil {
				request.Header.Set(snapTradeSignatureHeader, test.signature(test.body))
			}
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			if response.Code != test.wantStatus || service.calls != test.wantCalls {
				t.Fatalf("status=%d calls=%d body=%q", response.Code, service.calls, response.Body.String())
			}
		})
	}
}

func TestSnapTradeWebhookIsPublicAndDoesNotLogPayloadOrSecrets(t *testing.T) {
	const key = "do-not-log-consumer-key"
	body := validWebhookBody(t, portfolio.WebhookAccountRemoved, "", "private-account", time.Now())
	var logs bytes.Buffer
	service := &webhookLifecycleStub{}
	webhook, err := NewSnapTradeWebhookHandler(service, []byte(key), "client", true, slog.New(slog.NewJSONHandler(&logs, nil)))
	if err != nil {
		t.Fatal(err)
	}
	readiness := NewReadiness(pingFunc(func(context.Context) error { return nil }), 0)
	readiness.SetReady(true)
	handler := newHandler(slog.New(slog.NewJSONHandler(&logs, nil)), readiness, "development", nil, nil, nil, nil, nil, nil, nil, nil, nil, webhook, false, "")
	request := httptest.NewRequest(http.MethodPost, SnapTradeWebhookPath, strings.NewReader(body))
	request.Header.Set(contentTypeHeader, jsonMediaType)
	request.Header.Set(snapTradeSignatureHeader, webhookSignature(t, key, body))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusNoContent || service.calls != 1 {
		t.Fatalf("status=%d calls=%d", response.Code, service.calls)
	}
	for _, secret := range []string{key, "private-delivery", "private-subject", "private-account", request.Header.Get(snapTradeSignatureHeader)} {
		if strings.Contains(logs.String(), secret) {
			t.Fatalf("logs contain private value %q: %s", secret, logs.String())
		}
	}
}

func TestSnapTradeWebhookRejectsOversizedBody(t *testing.T) {
	service := &webhookLifecycleStub{}
	handler, err := NewSnapTradeWebhookHandler(service, []byte("key"), "client", true, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPost, SnapTradeWebhookPath, strings.NewReader(`{"padding":"`+strings.Repeat("x", snapTradeWebhookBodyLimit)+`"}`))
	request.Header.Set(contentTypeHeader, jsonMediaType)
	request.Header.Set(snapTradeSignatureHeader, "unused")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest || service.calls != 0 {
		t.Fatalf("status=%d calls=%d", response.Code, service.calls)
	}
}

func TestSnapTradeWebhookDefaultsCanRunLogOnly(t *testing.T) {
	body := validWebhookBody(t, portfolio.WebhookConnectionBroken, "connection", "", time.Now())
	service := &webhookLifecycleStub{}
	handler, err := NewSnapTradeWebhookHandler(service, []byte("key"), "client", false, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPost, SnapTradeWebhookPath, strings.NewReader(body))
	request.Header.Set(contentTypeHeader, jsonMediaType)
	request.Header.Set(snapTradeSignatureHeader, webhookSignature(t, "key", body))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusNoContent || service.calls != 0 {
		t.Fatalf("status=%d calls=%d", response.Code, service.calls)
	}
}

func TestSnapTradeWebhookMapsEverySupportedEvent(t *testing.T) {
	tests := []struct {
		name       string
		typeValue  portfolio.WebhookEventType
		connection string
		account    string
		want       portfolio.WebhookEvent
	}{
		{name: "connection added", typeValue: portfolio.WebhookConnectionAdded, connection: " connection ", want: portfolio.WebhookEvent{Subject: "subject", Type: portfolio.WebhookConnectionAdded, ConnectionID: "connection"}},
		{name: "connection broken", typeValue: portfolio.WebhookConnectionBroken, connection: " connection ", want: portfolio.WebhookEvent{Subject: "subject", Type: portfolio.WebhookConnectionBroken, ConnectionID: "connection"}},
		{name: "connection fixed", typeValue: portfolio.WebhookConnectionFixed, connection: " connection ", want: portfolio.WebhookEvent{Subject: "subject", Type: portfolio.WebhookConnectionFixed, ConnectionID: "connection"}},
		{name: "connection deleted", typeValue: portfolio.WebhookConnectionDeleted, connection: " connection ", want: portfolio.WebhookEvent{Subject: "subject", Type: portfolio.WebhookConnectionDeleted, ConnectionID: "connection"}},
		{name: "new account", typeValue: portfolio.WebhookNewAccountAvailable, connection: " connection ", account: " account ", want: portfolio.WebhookEvent{Subject: "subject", Type: portfolio.WebhookNewAccountAvailable, ConnectionID: "connection", AccountID: "account"}},
		{name: "account removed", typeValue: portfolio.WebhookAccountRemoved, account: " account ", want: portfolio.WebhookEvent{Subject: "subject", Type: portfolio.WebhookAccountRemoved, AccountID: "account"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := &webhookLifecycleStub{}
			handler, err := NewSnapTradeWebhookHandler(service, []byte("key"), "client", true, slog.New(slog.NewTextHandler(io.Discard, nil)))
			if err != nil {
				t.Fatal(err)
			}
			body := webhookBody(t, " subject ", test.typeValue, test.connection, test.account, time.Now())
			request := httptest.NewRequest(http.MethodPost, SnapTradeWebhookPath, strings.NewReader(body))
			request.Header.Set(contentTypeHeader, jsonMediaType)
			request.Header.Set(snapTradeSignatureHeader, webhookSignature(t, "key", body))
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			if response.Code != http.StatusNoContent || service.calls != 1 || service.event != test.want {
				t.Fatalf("status=%d calls=%d event=%+v", response.Code, service.calls, service.event)
			}
		})
	}
}

func TestSnapTradeWebhookValidatesSupportedIdentifiersBeforeModeBranch(t *testing.T) {
	tests := []struct {
		name       string
		typeValue  portfolio.WebhookEventType
		connection string
		account    string
	}{
		{name: "connection added needs connection", typeValue: portfolio.WebhookConnectionAdded},
		{name: "connection broken needs connection", typeValue: portfolio.WebhookConnectionBroken},
		{name: "connection fixed needs connection", typeValue: portfolio.WebhookConnectionFixed},
		{name: "connection deleted needs connection", typeValue: portfolio.WebhookConnectionDeleted},
		{name: "new account needs connection", typeValue: portfolio.WebhookNewAccountAvailable, account: "account"},
		{name: "new account needs account", typeValue: portfolio.WebhookNewAccountAvailable, connection: "connection"},
		{name: "account removed needs account", typeValue: portfolio.WebhookAccountRemoved},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := &webhookLifecycleStub{}
			handler, err := NewSnapTradeWebhookHandler(service, []byte("key"), "client", false, slog.New(slog.NewTextHandler(io.Discard, nil)))
			if err != nil {
				t.Fatal(err)
			}
			body := webhookBody(t, "subject", test.typeValue, test.connection, test.account, time.Now())
			request := httptest.NewRequest(http.MethodPost, SnapTradeWebhookPath, strings.NewReader(body))
			request.Header.Set(contentTypeHeader, jsonMediaType)
			request.Header.Set(snapTradeSignatureHeader, webhookSignature(t, "key", body))
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			if response.Code != http.StatusBadRequest || service.calls != 0 {
				t.Fatalf("status=%d calls=%d", response.Code, service.calls)
			}
		})
	}

	service := &webhookLifecycleStub{}
	handler, err := NewSnapTradeWebhookHandler(service, []byte("key"), "client", false, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	body := webhookBody(t, "subject", "UNSUPPORTED_EVENT", "", "", time.Now())
	request := httptest.NewRequest(http.MethodPost, SnapTradeWebhookPath, strings.NewReader(body))
	request.Header.Set(contentTypeHeader, jsonMediaType)
	request.Header.Set(snapTradeSignatureHeader, webhookSignature(t, "key", body))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusNoContent || service.calls != 0 {
		t.Fatalf("unsupported status=%d calls=%d", response.Code, service.calls)
	}
}

func TestSnapTradeWebhookBoundsEventTimestamp(t *testing.T) {
	tests := []struct {
		name       string
		timestamp  time.Time
		wantStatus int
		wantCalls  int
	}{
		{name: "recent", timestamp: time.Now().Add(-time.Minute), wantStatus: http.StatusNoContent, wantCalls: 1},
		{name: "older than five minutes", timestamp: time.Now().Add(-webhookTimestampTolerance - time.Second), wantStatus: http.StatusBadRequest},
		{name: "materially future dated", timestamp: time.Now().Add(webhookTimestampTolerance + time.Second), wantStatus: http.StatusBadRequest},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := &webhookLifecycleStub{}
			handler, err := NewSnapTradeWebhookHandler(service, []byte("key"), "client", true, slog.New(slog.NewTextHandler(io.Discard, nil)))
			if err != nil {
				t.Fatal(err)
			}
			body := validWebhookBody(t, portfolio.WebhookConnectionBroken, "connection", "", test.timestamp)
			request := httptest.NewRequest(http.MethodPost, SnapTradeWebhookPath, strings.NewReader(body))
			request.Header.Set(contentTypeHeader, jsonMediaType)
			request.Header.Set(snapTradeSignatureHeader, webhookSignature(t, "key", body))
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			if response.Code != test.wantStatus || service.calls != test.wantCalls {
				t.Fatalf("status=%d calls=%d", response.Code, service.calls)
			}
		})
	}
}

func TestCanonicalWebhookJSONMatchesPythonKnownVector(t *testing.T) {
	const body = `{"z":"café","html":"<tag>&","nested":{"é":"雪","a":true},"emoji":"😀"}`
	const wantCanonical = `{"emoji":"\ud83d\ude00","html":"<tag>&","nested":{"a":true,"\u00e9":"\u96ea"},"z":"caf\u00e9"}`
	const wantSignature = "N64dEwvzvFRz4KMR56nIKTOujXAV64DF8fymAnIRj2M="
	request := httptest.NewRequest(http.MethodPost, SnapTradeWebhookPath, strings.NewReader(body))
	canonical, _, err := decodeSnapTradeWebhook(httptest.NewRecorder(), request)
	if err != nil {
		t.Fatal(err)
	}
	if string(canonical) != wantCanonical {
		t.Fatalf("canonical=%q", canonical)
	}
	mac := hmac.New(sha256.New, []byte("known-key"))
	_, _ = mac.Write(canonical)
	if got := base64.StdEncoding.EncodeToString(mac.Sum(nil)); got != wantSignature {
		t.Fatalf("signature=%q", got)
	}
}

func TestWebhookRouteHasDedicatedCategory(t *testing.T) {
	if got := routeCategory(SnapTradeWebhookPath); got != "snaptrade_webhook" {
		t.Fatalf("route category=%q", got)
	}
}

func validWebhookBody(t *testing.T, eventType portfolio.WebhookEventType, connectionID, accountID string, timestamp time.Time) string {
	t.Helper()
	return webhookBody(t, "subject", eventType, connectionID, accountID, timestamp)
}

func webhookBody(t *testing.T, subject string, eventType portfolio.WebhookEventType, connectionID, accountID string, timestamp time.Time) string {
	t.Helper()
	document := map[string]any{
		"schemaVersion":  snapTradeWebhookSchema,
		"webhookId":      "delivery",
		"oauthClientId":  "client",
		"eventTimestamp": timestamp.UTC().Format(time.RFC3339Nano),
		"userId":         subject,
		"eventType":      eventType,
	}
	if connectionID != "" {
		document["connectionId"] = connectionID
	}
	if accountID != "" {
		document["accountId"] = accountID
	}
	body, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	return string(body)
}

func webhookSignature(t *testing.T, key, body string) string {
	t.Helper()
	decoder := json.NewDecoder(strings.NewReader(body))
	decoder.UseNumber()
	var document map[string]any
	if err := decoder.Decode(&document); err != nil {
		t.Fatal(err)
	}
	canonical, err := canonicalWebhookJSON(document)
	if err != nil {
		t.Fatal(err)
	}
	mac := hmac.New(sha256.New, []byte(key))
	_, _ = mac.Write(canonical)
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}
