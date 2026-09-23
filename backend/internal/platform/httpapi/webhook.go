package httpapi

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf16"

	"github.com/kennethdavidbuck/findur/backend/internal/portfolio"
)

const (
	// SnapTradeWebhookPath is the public receiver configured in SnapTrade.
	SnapTradeWebhookPath       = "/api/webhooks/snaptrade"
	snapTradeWebhookSchema     = "oauth_v1"
	snapTradeSignatureHeader   = "Signature"
	snapTradeWebhookBodyLimit  = 64 << 10
	maxWebhookIdentifierLength = 256
	webhookTimestampTolerance  = 5 * time.Minute
)

type webhookLifecycle interface {
	Apply(context.Context, portfolio.WebhookEvent) (bool, error)
}

type snapTradeWebhookPayload struct {
	SchemaVersion  string `json:"schemaVersion"`
	WebhookID      string `json:"webhookId"`
	OAuthClientID  string `json:"oauthClientId"`
	EventTimestamp string `json:"eventTimestamp"`
	UserID         string `json:"userId"`
	EventType      string `json:"eventType"`
	ConnectionID   string `json:"connectionId"`
	AccountID      string `json:"accountId"`
}

type snapTradeWebhookHandler struct {
	service     webhookLifecycle
	consumerKey []byte
	clientID    string
	process     bool
	logger      *slog.Logger
}

// NewSnapTradeWebhookHandler constructs the public signature-authenticated receiver.
func NewSnapTradeWebhookHandler(service webhookLifecycle, consumerKey []byte, clientID string, processingEnabled bool, logger *slog.Logger) (http.Handler, error) {
	if service == nil || len(consumerKey) == 0 || strings.TrimSpace(clientID) == "" || logger == nil {
		return nil, errors.New("incomplete SnapTrade webhook configuration")
	}
	keyCopy := append([]byte(nil), consumerKey...)
	return &snapTradeWebhookHandler{service: service, consumerKey: keyCopy, clientID: clientID, process: processingEnabled, logger: logger}, nil
}

func (h *snapTradeWebhookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(cacheControlHeader, noStoreDirective)
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	mediaType, _, err := mime.ParseMediaType(r.Header.Get(contentTypeHeader))
	if err != nil || mediaType != jsonMediaType {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	canonical, payload, err := decodeSnapTradeWebhook(w, r)
	if err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if !h.validSignature(r.Header.Values(snapTradeSignatureHeader), canonical) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if !validSnapTradeWebhookEnvelope(payload, h.clientID, time.Now().UTC()) {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	event, _, err := portfolio.ValidateWebhookEvent(portfolio.WebhookEvent{
		Subject:      payload.UserID,
		Type:         portfolio.WebhookEventType(payload.EventType),
		ConnectionID: payload.ConnectionID,
		AccountID:    payload.AccountID,
	})
	if err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if !h.process {
		h.logger.InfoContext(r.Context(), "SnapTrade webhook accepted in log-only mode", "event", "snaptrade_webhook_accepted", "event_type", payload.EventType, "processing_enabled", false)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	changed, err := h.service.Apply(r.Context(), event)
	if err != nil {
		status := http.StatusServiceUnavailable
		if errors.Is(err, portfolio.ErrInvalidWebhookEvent) {
			status = http.StatusBadRequest
		}
		h.logger.WarnContext(r.Context(), "SnapTrade webhook rejected", "event", "snaptrade_webhook_rejected", "event_type", payload.EventType, "status", status)
		http.Error(w, http.StatusText(status), status)
		return
	}
	h.logger.InfoContext(r.Context(), "SnapTrade webhook accepted", "event", "snaptrade_webhook_accepted", "event_type", payload.EventType, "changed", changed)
	w.WriteHeader(http.StatusNoContent)
}

func decodeSnapTradeWebhook(w http.ResponseWriter, r *http.Request) ([]byte, snapTradeWebhookPayload, error) {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, snapTradeWebhookBodyLimit))
	decoder.UseNumber()
	var document map[string]any
	if err := decoder.Decode(&document); err != nil || document == nil {
		return nil, snapTradeWebhookPayload{}, errors.New("invalid JSON")
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return nil, snapTradeWebhookPayload{}, errors.New("multiple JSON values")
	}
	canonical, err := canonicalWebhookJSON(document)
	if err != nil {
		return nil, snapTradeWebhookPayload{}, err
	}
	var payload snapTradeWebhookPayload
	if err := json.Unmarshal(canonical, &payload); err != nil {
		return nil, snapTradeWebhookPayload{}, err
	}
	return canonical, payload, nil
}

func canonicalWebhookJSON(document any) ([]byte, error) {
	result := make([]byte, 0, 512)
	return appendCanonicalJSON(result, document)
}

func appendCanonicalJSON(dst []byte, value any) ([]byte, error) {
	switch typed := value.(type) {
	case nil:
		return append(dst, "null"...), nil
	case bool:
		return strconv.AppendBool(dst, typed), nil
	case string:
		return appendCanonicalJSONString(dst, typed), nil
	case json.Number:
		if _, err := typed.Int64(); err != nil {
			if _, err := typed.Float64(); err != nil {
				return nil, err
			}
		}
		return append(dst, typed.String()...), nil
	case []any:
		dst = append(dst, '[')
		for index, item := range typed {
			if index > 0 {
				dst = append(dst, ',')
			}
			var err error
			dst, err = appendCanonicalJSON(dst, item)
			if err != nil {
				return nil, err
			}
		}
		return append(dst, ']'), nil
	case map[string]any:
		keys := make([]string, 0, len(typed))
		for key := range typed {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		dst = append(dst, '{')
		for index, key := range keys {
			if index > 0 {
				dst = append(dst, ',')
			}
			dst = appendCanonicalJSONString(dst, key)
			dst = append(dst, ':')
			var err error
			dst, err = appendCanonicalJSON(dst, typed[key])
			if err != nil {
				return nil, err
			}
		}
		return append(dst, '}'), nil
	default:
		return nil, errors.New("unsupported JSON value")
	}
}

func appendCanonicalJSONString(dst []byte, value string) []byte {
	const hexadecimal = "0123456789abcdef"
	dst = append(dst, '"')
	for _, character := range value {
		switch character {
		case '"', '\\':
			dst = append(dst, '\\', byte(character))
		case '\b':
			dst = append(dst, `\b`...)
		case '\f':
			dst = append(dst, `\f`...)
		case '\n':
			dst = append(dst, `\n`...)
		case '\r':
			dst = append(dst, `\r`...)
		case '\t':
			dst = append(dst, `\t`...)
		default:
			switch {
			case character >= 0x20 && character <= 0x7e:
				dst = append(dst, byte(character))
			case character <= 0xffff:
				dst = appendUnicodeEscape(dst, uint16(character), hexadecimal)
			default:
				high, low := utf16.EncodeRune(character)
				dst = appendUnicodeEscape(dst, uint16(high), hexadecimal)
				dst = appendUnicodeEscape(dst, uint16(low), hexadecimal)
			}
		}
	}
	return append(dst, '"')
}

func appendUnicodeEscape(dst []byte, value uint16, hexadecimal string) []byte {
	return append(dst, '\\', 'u', hexadecimal[value>>12], hexadecimal[value>>8&0xf], hexadecimal[value>>4&0xf], hexadecimal[value&0xf])
}

func (h *snapTradeWebhookHandler) validSignature(values []string, canonical []byte) bool {
	if len(values) != 1 {
		return false
	}
	provided, err := base64.StdEncoding.DecodeString(strings.TrimSpace(values[0]))
	if err != nil || len(provided) != sha256.Size {
		return false
	}
	mac := hmac.New(sha256.New, h.consumerKey)
	_, _ = mac.Write(canonical)
	return hmac.Equal(provided, mac.Sum(nil))
}

func validSnapTradeWebhookEnvelope(payload snapTradeWebhookPayload, clientID string, now time.Time) bool {
	if payload.SchemaVersion != snapTradeWebhookSchema || payload.OAuthClientID != clientID {
		return false
	}
	for _, value := range []string{payload.WebhookID, payload.UserID, payload.EventType} {
		if strings.TrimSpace(value) == "" || len(value) > maxWebhookIdentifierLength {
			return false
		}
	}
	eventTimestamp, err := time.Parse(time.RFC3339Nano, payload.EventTimestamp)
	if err != nil {
		return false
	}
	return !eventTimestamp.Before(now.Add(-webhookTimestampTolerance)) && !eventTimestamp.After(now.Add(webhookTimestampTolerance))
}
