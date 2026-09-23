package postgres

import (
	"testing"
	"time"

	"github.com/kennethdavidbuck/findur/backend/internal/portfolio"
)

func TestResourceDiagnosticProjectsInitialSyncAsPending(t *testing.T) {
	diagnostic := resourceDiagnostic(nil, nil, nil, nil)
	if diagnostic == nil || diagnostic.Reason != portfolio.DiagnosticSyncPending || diagnostic.RecommendedAction != portfolio.DiagnosticActionWait {
		t.Fatalf("diagnostic=%+v", diagnostic)
	}

	lastSuccess := time.Date(2026, 9, 23, 12, 0, 0, 0, time.UTC)
	if diagnostic := resourceDiagnostic(nil, nil, nil, &lastSuccess); diagnostic != nil {
		t.Fatalf("successful resource diagnostic=%+v want nil", diagnostic)
	}
}

func TestUnusableAccountDiagnosticDistinguishesConnectionRepair(t *testing.T) {
	tests := []struct {
		name            string
		connectionState string
		accountState    string
		wantReason      portfolio.ResourceDiagnosticReason
		wantAction      portfolio.ResourceDiagnosticAction
	}{
		{name: "disabled brokerage connection", connectionState: string(portfolio.ConnectionStatusDisabled), accountState: string(portfolio.AccountSyncStateComplete), wantReason: portfolio.DiagnosticConnectionDisabled, wantAction: portfolio.DiagnosticActionReconnect},
		{name: "account synchronization pending", connectionState: string(portfolio.ConnectionStatusActive), accountState: string(portfolio.AccountSyncStatePending), wantReason: portfolio.DiagnosticSyncPending, wantAction: portfolio.DiagnosticActionWait},
		{name: "account synchronization unavailable", connectionState: string(portfolio.ConnectionStatusActive), accountState: string(portfolio.AccountSyncStateUnavailable), wantReason: portfolio.DiagnosticProviderUnavailable, wantAction: portfolio.DiagnosticActionRetry},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			reason, action := unusableAccountDiagnostic(test.connectionState, test.accountState)
			if reason != test.wantReason || action != test.wantAction {
				t.Fatalf("reason=%q action=%q want reason=%q action=%q", reason, action, test.wantReason, test.wantAction)
			}
		})
	}
}

func TestSyncFailureDiagnosticReasonIsBounded(t *testing.T) {
	tests := []struct {
		failure string
		want    portfolio.ResourceDiagnosticReason
	}{
		{failure: "authorization_required", want: portfolio.DiagnosticAuthorizationRequired},
		{failure: "rate_limited", want: portfolio.DiagnosticProviderUnavailable},
		{failure: "provider_unavailable", want: portfolio.DiagnosticProviderUnavailable},
		{failure: "unusable_data", want: portfolio.DiagnosticUnknown},
		{failure: "unexpected", want: portfolio.DiagnosticUnknown},
	}
	for _, test := range tests {
		t.Run(test.failure, func(t *testing.T) {
			if got := syncFailureDiagnosticReason(test.failure); got != test.want {
				t.Fatalf("reason=%q want=%q", got, test.want)
			}
		})
	}
}

func TestSyncFailureDiagnosticActionMatchesRecovery(t *testing.T) {
	tests := []struct {
		failure string
		want    portfolio.ResourceDiagnosticAction
	}{
		{failure: "authorization_required", want: portfolio.DiagnosticActionReconnect},
		{failure: "rate_limited", want: portfolio.DiagnosticActionWait},
		{failure: "provider_unavailable", want: portfolio.DiagnosticActionRetry},
		{failure: "unusable_data", want: portfolio.DiagnosticActionRetry},
	}
	for _, test := range tests {
		t.Run(test.failure, func(t *testing.T) {
			if got := syncFailureDiagnosticAction(test.failure); got != test.want {
				t.Fatalf("action=%q want=%q", got, test.want)
			}
		})
	}
}

func TestDiagnosticColumnsAreResourceScoped(t *testing.T) {
	tests := []struct {
		resource              portfolio.AccountResource
		reason, action, retry string
	}{
		{resource: portfolio.AccountResourceBalances, reason: "balances_failure_reason", action: "balances_failure_action", retry: "balances_retry_at"},
		{resource: portfolio.AccountResourcePositions, reason: "positions_failure_reason", action: "positions_failure_action", retry: "positions_retry_at"},
		{resource: portfolio.AccountResourceActivities, reason: "activities_failure_reason", action: "activities_failure_action", retry: "activities_retry_at"},
	}
	for _, test := range tests {
		t.Run(string(test.resource), func(t *testing.T) {
			reason, action, retry, err := diagnosticColumns(test.resource)
			if err != nil || reason != test.reason || action != test.action || retry != test.retry {
				t.Fatalf("reason=%q action=%q retry=%q err=%v", reason, action, retry, err)
			}
		})
	}
}
