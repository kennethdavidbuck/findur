package portfolio

import (
	"context"
	"errors"
	"strings"
	"time"
)

// WebhookEventType is a normalized SnapTrade inventory lifecycle event.
type WebhookEventType string

// Supported SnapTrade inventory lifecycle event types.
const (
	WebhookConnectionAdded     WebhookEventType = "CONNECTION_ADDED"
	WebhookConnectionBroken    WebhookEventType = "CONNECTION_BROKEN"
	WebhookConnectionFixed     WebhookEventType = "CONNECTION_FIXED"
	WebhookConnectionDeleted   WebhookEventType = "CONNECTION_DELETED"
	WebhookNewAccountAvailable WebhookEventType = "NEW_ACCOUNT_AVAILABLE"
	WebhookAccountRemoved      WebhookEventType = "ACCOUNT_REMOVED"
)

// WebhookEvent contains only the provider-independent identity needed to apply
// an inventory lifecycle notification.
type WebhookEvent struct {
	Subject      string
	Type         WebhookEventType
	ConnectionID string
	AccountID    string
}

var (
	// ErrInvalidWebhookEvent means a supported event omitted a required identity.
	ErrInvalidWebhookEvent = errors.New("invalid portfolio webhook event")
	// ErrInventoryBusy preserves an active inventory worker's claim.
	ErrInventoryBusy = errors.New("portfolio inventory is busy")
)

// InventoryLifecycleRepository exposes entity operations without knowledge of
// webhook transports or event names.
type InventoryLifecycleRepository interface {
	AddConnection(context.Context, string, string, time.Time) (bool, error)
	SetConnectionStatus(context.Context, string, string, ConnectionStatus, time.Time) (bool, error)
	RemoveConnection(context.Context, string, string, time.Time) (bool, error)
	AddAccount(context.Context, string, string, string, time.Time) (bool, error)
	RemoveAccount(context.Context, string, string, time.Time) (bool, error)
}

// WebhookService keeps provider transport concerns outside lifecycle mutation.
type WebhookService struct {
	repository InventoryLifecycleRepository
	clock      func() time.Time
}

// NewWebhookService validates and constructs the webhook application service.
func NewWebhookService(repository InventoryLifecycleRepository, clock func() time.Time) (*WebhookService, error) {
	if repository == nil || clock == nil {
		return nil, errors.New("incomplete portfolio webhook configuration")
	}
	return &WebhookService{repository: repository, clock: clock}, nil
}

// ValidateWebhookEvent normalizes transport-provided identifiers and validates
// the identities required by supported lifecycle events. Unsupported events are
// acknowledged without imposing lifecycle-specific identifier requirements.
func ValidateWebhookEvent(event WebhookEvent) (WebhookEvent, bool, error) {
	event.Subject = strings.TrimSpace(event.Subject)
	event.ConnectionID = strings.TrimSpace(event.ConnectionID)
	event.AccountID = strings.TrimSpace(event.AccountID)
	if event.Subject == "" || len(event.Subject) > 256 {
		return WebhookEvent{}, false, ErrInvalidWebhookEvent
	}
	switch event.Type {
	case WebhookConnectionAdded, WebhookConnectionBroken, WebhookConnectionFixed, WebhookConnectionDeleted:
		if event.ConnectionID == "" || len(event.ConnectionID) > 128 {
			return WebhookEvent{}, false, ErrInvalidWebhookEvent
		}
	case WebhookNewAccountAvailable:
		if event.ConnectionID == "" || len(event.ConnectionID) > 128 || event.AccountID == "" || len(event.AccountID) > 128 {
			return WebhookEvent{}, false, ErrInvalidWebhookEvent
		}
	case WebhookAccountRemoved:
		if event.AccountID == "" || len(event.AccountID) > 128 {
			return WebhookEvent{}, false, ErrInvalidWebhookEvent
		}
	default:
		return event, false, nil
	}
	return event, true, nil
}

// Apply validates supported events, acknowledges unsupported events without a
// mutation, and delegates persistence through the transport-neutral port.
func (s *WebhookService) Apply(ctx context.Context, event WebhookEvent) (bool, error) {
	var supported bool
	var err error
	event, supported, err = ValidateWebhookEvent(event)
	if err != nil {
		return false, err
	}
	if !supported {
		return false, nil
	}
	now := s.clock().UTC()
	switch event.Type {
	case WebhookConnectionAdded:
		return s.repository.AddConnection(ctx, event.Subject, event.ConnectionID, now)
	case WebhookConnectionBroken:
		return s.repository.SetConnectionStatus(ctx, event.Subject, event.ConnectionID, ConnectionStatusDisabled, now)
	case WebhookConnectionFixed:
		return s.repository.SetConnectionStatus(ctx, event.Subject, event.ConnectionID, ConnectionStatusActive, now)
	case WebhookConnectionDeleted:
		return s.repository.RemoveConnection(ctx, event.Subject, event.ConnectionID, now)
	case WebhookNewAccountAvailable:
		return s.repository.AddAccount(ctx, event.Subject, event.ConnectionID, event.AccountID, now)
	case WebhookAccountRemoved:
		return s.repository.RemoveAccount(ctx, event.Subject, event.AccountID, now)
	default:
		return false, nil
	}
}
