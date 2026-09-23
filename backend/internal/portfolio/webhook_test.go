package portfolio

import (
	"context"
	"errors"
	"testing"
	"time"
)

type webhookRepositoryStub struct {
	event WebhookEvent
	err   error
	calls int
}

func (s *webhookRepositoryStub) record(event WebhookEvent) (bool, error) {
	s.calls++
	s.event = event
	return true, s.err
}

func (s *webhookRepositoryStub) AddConnection(_ context.Context, subject, connection string, _ time.Time) (bool, error) {
	return s.record(WebhookEvent{Subject: subject, Type: WebhookConnectionAdded, ConnectionID: connection})
}
func (s *webhookRepositoryStub) SetConnectionStatus(_ context.Context, subject, connection string, status ConnectionStatus, _ time.Time) (bool, error) {
	eventType := WebhookConnectionFixed
	if status == ConnectionStatusDisabled {
		eventType = WebhookConnectionBroken
	}
	return s.record(WebhookEvent{Subject: subject, Type: eventType, ConnectionID: connection})
}
func (s *webhookRepositoryStub) RemoveConnection(_ context.Context, subject, connection string, _ time.Time) (bool, error) {
	return s.record(WebhookEvent{Subject: subject, Type: WebhookConnectionDeleted, ConnectionID: connection})
}
func (s *webhookRepositoryStub) AddAccount(_ context.Context, subject, connection, account string, _ time.Time) (bool, error) {
	return s.record(WebhookEvent{Subject: subject, Type: WebhookNewAccountAvailable, ConnectionID: connection, AccountID: account})
}
func (s *webhookRepositoryStub) RemoveAccount(_ context.Context, subject, account string, _ time.Time) (bool, error) {
	return s.record(WebhookEvent{Subject: subject, Type: WebhookAccountRemoved, AccountID: account})
}

func TestWebhookServiceValidatesAndRoutesEvents(t *testing.T) {
	tests := []struct {
		name      string
		event     WebhookEvent
		wantCalls int
		wantErr   error
	}{
		{name: "connection added", event: WebhookEvent{Subject: "owner", Type: WebhookConnectionAdded, ConnectionID: "connection"}, wantCalls: 1},
		{name: "connection broken", event: WebhookEvent{Subject: "owner", Type: WebhookConnectionBroken, ConnectionID: "connection"}, wantCalls: 1},
		{name: "connection fixed", event: WebhookEvent{Subject: "owner", Type: WebhookConnectionFixed, ConnectionID: "connection"}, wantCalls: 1},
		{name: "connection deleted", event: WebhookEvent{Subject: "owner", Type: WebhookConnectionDeleted, ConnectionID: "connection"}, wantCalls: 1},
		{name: "new account", event: WebhookEvent{Subject: "owner", Type: WebhookNewAccountAvailable, ConnectionID: "connection", AccountID: "account"}, wantCalls: 1},
		{name: "account removed", event: WebhookEvent{Subject: "owner", Type: WebhookAccountRemoved, AccountID: "account"}, wantCalls: 1},
		{name: "unsupported acknowledged", event: WebhookEvent{Subject: "owner", Type: "ACCOUNT_HOLDINGS_UPDATED"}},
		{name: "missing subject", event: WebhookEvent{Type: WebhookConnectionAdded, ConnectionID: "connection"}, wantErr: ErrInvalidWebhookEvent},
		{name: "missing connection", event: WebhookEvent{Subject: "owner", Type: WebhookConnectionDeleted}, wantErr: ErrInvalidWebhookEvent},
		{name: "missing account connection", event: WebhookEvent{Subject: "owner", Type: WebhookNewAccountAvailable, AccountID: "account"}, wantErr: ErrInvalidWebhookEvent},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := &webhookRepositoryStub{}
			service, err := NewWebhookService(repository, time.Now)
			if err != nil {
				t.Fatal(err)
			}
			_, err = service.Apply(context.Background(), test.event)
			if !errors.Is(err, test.wantErr) || repository.calls != test.wantCalls {
				t.Fatalf("calls=%d err=%v", repository.calls, err)
			}
		})
	}
}

func TestWebhookServicePropagatesRepositoryFailure(t *testing.T) {
	want := errors.New("database unavailable")
	repository := &webhookRepositoryStub{err: want}
	service, err := NewWebhookService(repository, time.Now)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Apply(context.Background(), WebhookEvent{Subject: "owner", Type: WebhookConnectionFixed, ConnectionID: "connection"}); !errors.Is(err, want) {
		t.Fatalf("error=%v", err)
	}
}
