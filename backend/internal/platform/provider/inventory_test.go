package provider

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/kennethdavidbuck/findur/backend/internal/portfolio"
)

func TestInventoryListsConnectionsBeforeAccountsAndMinimizesOutput(t *testing.T) {
	baseURL, _ := url.Parse("https://api.snaptrade.example")
	var paths []string
	client, err := NewInventoryClient(baseURL, roundTripFunc(func(request *http.Request) (*http.Response, error) {
		paths = append(paths, request.URL.Path)
		if request.Header.Get("Authorization") != "Bearer access-token" || request.URL.RawQuery != "" {
			t.Fatalf("unsafe request: %s auth=%q query=%q", request.URL, request.Header.Get("Authorization"), request.URL.RawQuery)
		}
		for _, forbidden := range []string{"SnapTrade-Client-Id", "userId", "userSecret", "Signature"} {
			if request.Header.Get(forbidden) != "" {
				t.Fatalf("forbidden provider header %q", forbidden)
			}
		}
		body := fixtureBody(t, "success-connections.json")
		if strings.HasSuffix(request.URL.Path, "/accounts") {
			body = fixtureBody(t, "success-accounts.json")
		}
		return jsonResponse(http.StatusOK, body), nil
	}), time.Now)
	if err != nil {
		t.Fatal(err)
	}

	connections, err := client.Load(context.Background(), "access-token")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(paths, ",") != "/authorizations,/authorizations/87b24961-b51e-4db8-9226-f198f6518a89/accounts" {
		t.Fatalf("request order = %v", paths)
	}
	if len(connections) != 1 || len(connections[0].Accounts) != 1 {
		t.Fatalf("connections = %+v", connections)
	}
	account := connections[0].Accounts[0]
	if account.MaskedLabel != "Retirement (•••• 8443)" || !account.Eligible || account.Category != "investment" || account.SyncState != "complete" {
		t.Fatalf("account = %+v", account)
	}
	serialized := strings.Join([]string{account.ID, string(account.Category), account.Type, account.MaskedLabel, string(account.SyncState)}, " ")
	for _, forbidden := range []string{"Q6542138443", "15363.23", "PRIVATE"} {
		if strings.Contains(serialized, forbidden) {
			t.Fatalf("normalized inventory retained %q: %s", forbidden, serialized)
		}
	}
}

func TestInventoryClassifiesProviderFailures(t *testing.T) {
	now := time.Date(2026, 9, 20, 12, 0, 0, 0, time.UTC)
	for _, test := range []struct {
		name    string
		status  int
		fixture string
		want    portfolio.State
	}{
		{name: "unauthorized", status: 401, fixture: "unauthorized.json", want: portfolio.StateUnauthorized},
		{name: "forbidden is ambiguous", status: 403, fixture: "disabled.json", want: portfolio.StateUnavailable},
		{name: "not found is ambiguous", status: 404, fixture: "disabled.json", want: portfolio.StateUnavailable},
		{name: "rate limited", status: 429, fixture: "rate-limited.json", want: portfolio.StateRateLimited},
		{name: "transient", status: 503, fixture: "transient.json", want: portfolio.StateUnavailable},
		{name: "malformed", status: 200, fixture: "malformed.json", want: portfolio.StateMalformed},
	} {
		t.Run(test.name, func(t *testing.T) {
			baseURL, _ := url.Parse("https://api.snaptrade.example")
			client, err := NewInventoryClient(baseURL, roundTripFunc(func(*http.Request) (*http.Response, error) {
				response := jsonResponse(test.status, fixtureBody(t, test.fixture))
				if test.status == 429 {
					response.Header.Set("Retry-After", "60")
				}
				return response, nil
			}), func() time.Time { return now })
			if err != nil {
				t.Fatal(err)
			}
			_, err = client.Load(context.Background(), "access-token")
			var providerErr *portfolio.ProviderError
			if !errors.As(err, &providerErr) || providerErr.State != test.want {
				t.Fatalf("error = %v, want %s", err, test.want)
			}
			if test.status == 429 && (providerErr.RetryAt == nil || !providerErr.RetryAt.Equal(now.Add(time.Minute))) {
				t.Fatalf("retryAt = %v", providerErr.RetryAt)
			}
		})
	}
}

func TestInventoryPreservesAccountUsabilityPrecedence(t *testing.T) {
	accounts := fixtureBody(t, "success-accounts.json")
	for _, test := range []struct {
		name, accounts string
		want           portfolio.UsabilityReason
		selectable     bool
	}{
		{name: "missing status", accounts: strings.Replace(accounts, `"status": "open",`, "", 1), want: portfolio.UsabilityProvisionalStatus, selectable: true},
		{name: "missing category", accounts: strings.Replace(accounts, `"account_category": "INVESTMENT",`, "", 1), want: portfolio.UsabilityProvisionalCategory, selectable: true},
		{
			name: "closed before unavailable holdings",
			accounts: strings.NewReplacer(
				`"status": "open"`, `"status": "closed"`,
				`"initial_sync_completed": true`, `"initial_sync_completed": true, "holdings_unavailable": true`,
			).Replace(accounts),
			want: portfolio.UsabilityAccountClosed,
		},
		{
			name:     "unavailable holdings before sync reason",
			accounts: strings.Replace(accounts, `"initial_sync_completed": true`, `"initial_sync_completed": true, "holdings_unavailable": true`, 1),
			want:     portfolio.UsabilityAccountUnavailable,
		},
		{
			name: "unsupported category before missing status",
			accounts: strings.NewReplacer(
				`"status": "open",`, "",
				`"account_category": "INVESTMENT"`, `"account_category": "DEPOSIT"`,
			).Replace(accounts),
			want: portfolio.UsabilityUnsupportedCategory,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			baseURL, _ := url.Parse("https://api.snaptrade.example")
			client, _ := NewInventoryClient(baseURL, roundTripFunc(func(request *http.Request) (*http.Response, error) {
				if strings.HasSuffix(request.URL.Path, "/accounts") {
					return jsonResponse(http.StatusOK, test.accounts), nil
				}
				return jsonResponse(http.StatusOK, fixtureBody(t, "success-connections.json")), nil
			}), time.Now)
			connections, err := client.Load(context.Background(), "access-token")
			if err != nil || len(connections) != 1 || len(connections[0].Accounts) != 1 {
				t.Fatalf("connections=%+v err=%v", connections, err)
			}
			account := connections[0].Accounts[0]
			if account.Selectable != test.selectable || account.Eligible || account.UsabilityReason != test.want {
				t.Fatalf("account=%+v want reason=%s", account, test.want)
			}
		})
	}
}

func TestInventoryMakesMissingInitialHoldingsUnselectable(t *testing.T) {
	accounts := strings.Replace(fixtureBody(t, "success-accounts.json"), `"initial_sync_completed": true`, `"initial_sync_completed": false`, 1)
	baseURL, _ := url.Parse("https://api.snaptrade.example")
	client, _ := NewInventoryClient(baseURL, roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if strings.HasSuffix(request.URL.Path, "/accounts") {
			return jsonResponse(http.StatusOK, accounts), nil
		}
		return jsonResponse(http.StatusOK, fixtureBody(t, "success-connections.json")), nil
	}), time.Now)
	connections, err := client.Load(context.Background(), "access-token")
	if err != nil || len(connections) != 1 || len(connections[0].Accounts) != 1 {
		t.Fatalf("connections=%+v err=%v", connections, err)
	}
	account := connections[0].Accounts[0]
	if account.Selectable || account.Eligible || account.UsabilityReason != portfolio.UsabilitySyncPending {
		t.Fatalf("pending initial holdings account=%+v", account)
	}
}

func TestInventoryRateLimitRetryTiming(t *testing.T) {
	now := time.Date(2026, 9, 20, 12, 0, 0, 0, time.UTC)
	for _, test := range []struct {
		name       string
		retryAfter string
		reset      string
		want       *time.Time
	}{
		{name: "delta seconds", retryAfter: "60", want: timePointer(now.Add(time.Minute))},
		{name: "HTTP date", retryAfter: now.Add(2 * time.Minute).Format(http.TimeFormat), want: timePointer(now.Add(2 * time.Minute))},
		{name: "account reset fallback", reset: "90", want: timePointer(now.Add(90 * time.Second))},
		{name: "retry after takes precedence", retryAfter: "30", reset: "90", want: timePointer(now.Add(30 * time.Second))},
		{name: "invalid timing gets conservative floor", retryAfter: "later", want: timePointer(now.Add(time.Minute))},
		{name: "overflowing delta saturates safely", retryAfter: "999999999999999999999999999999999", want: timePointer(now.Add(time.Duration(1<<63 - 1)))},
	} {
		t.Run(test.name, func(t *testing.T) {
			baseURL, _ := url.Parse("https://api.snaptrade.example")
			client, _ := NewInventoryClient(baseURL, roundTripFunc(func(*http.Request) (*http.Response, error) {
				return nil, errors.New("unused")
			}), func() time.Time { return now })
			response := jsonResponse(http.StatusTooManyRequests, "{}")
			t.Cleanup(func() { _ = response.Body.Close() })
			response.Header.Set("Retry-After", test.retryAfter)
			if test.reset != "" {
				response.Header.Set("X-RateLimit-Account-Remaining", "0")
				response.Header.Set("X-RateLimit-Account-Reset", test.reset)
			}
			var providerErr *portfolio.ProviderError
			if !errors.As(client.responseError(response), &providerErr) {
				t.Fatal("expected categorical provider error")
			}
			if test.want == nil && providerErr.RetryAt != nil {
				t.Fatalf("retryAt = %v, want nil", providerErr.RetryAt)
			}
			if test.want != nil && (providerErr.RetryAt == nil || !providerErr.RetryAt.Equal(*test.want)) {
				t.Fatalf("retryAt = %v, want %v", providerErr.RetryAt, test.want)
			}
		})
	}
}

func TestInventoryNormalizesMaskingFreshnessAndRejectsDuplicateIDs(t *testing.T) {
	baseURL, _ := url.Parse("https://api.snaptrade.example")
	t.Run("short number and known number in label stay hidden", func(t *testing.T) {
		accounts := strings.Replace(fixtureBody(t, "success-accounts.json"), `"name": "Retirement"`, `"name": "Account 1234"`, 1)
		accounts = strings.Replace(accounts, `"number": "Q6542138443"`, `"number": "1234"`, 1)
		client, _ := NewInventoryClient(baseURL, roundTripFunc(func(request *http.Request) (*http.Response, error) {
			if strings.HasSuffix(request.URL.Path, "/accounts") {
				return jsonResponse(http.StatusOK, accounts), nil
			}
			return jsonResponse(http.StatusOK, fixtureBody(t, "success-connections.json")), nil
		}), time.Now)
		connections, err := client.Load(context.Background(), "access-token")
		if err != nil {
			t.Fatal(err)
		}
		label := connections[0].Accounts[0].MaskedLabel
		if strings.Contains(label, "1234") || strings.Contains(label, "(••••") {
			t.Fatalf("short account number leaked in %q", label)
		}
	})

	t.Run("known full number is removed from provider display name", func(t *testing.T) {
		accounts := strings.Replace(fixtureBody(t, "success-accounts.json"), `"name": "Retirement"`, `"name": "Account 1234 5678"`, 1)
		accounts = strings.Replace(accounts, `"number": "Q6542138443"`, `"number": "1234-5678"`, 1)
		client, _ := NewInventoryClient(baseURL, roundTripFunc(func(request *http.Request) (*http.Response, error) {
			if strings.HasSuffix(request.URL.Path, "/accounts") {
				return jsonResponse(http.StatusOK, accounts), nil
			}
			return jsonResponse(http.StatusOK, fixtureBody(t, "success-connections.json")), nil
		}), time.Now)
		connections, err := client.Load(context.Background(), "access-token")
		if err != nil {
			t.Fatal(err)
		}
		label := connections[0].Accounts[0].MaskedLabel
		if strings.Contains(label, "1234 5678") || !strings.Contains(label, "5678") {
			t.Fatalf("known account number was not safely minimized in %q", label)
		}
	})

	t.Run("institution delay controls effective freshness", func(t *testing.T) {
		connectionsJSON := strings.Replace(fixtureBody(t, "success-connections.json"), `"institution": "realtime", "snaptrade": "delayed"`, `"institution": "delayed", "snaptrade": "realtime"`, 1)
		client, _ := NewInventoryClient(baseURL, roundTripFunc(func(request *http.Request) (*http.Response, error) {
			if strings.HasSuffix(request.URL.Path, "/accounts") {
				return jsonResponse(http.StatusOK, fixtureBody(t, "success-accounts.json")), nil
			}
			return jsonResponse(http.StatusOK, connectionsJSON), nil
		}), time.Now)
		connections, err := client.Load(context.Background(), "access-token")
		if err != nil || connections[0].SyncMode != "delayed" {
			t.Fatalf("connections=%+v err=%v", connections, err)
		}
	})

	for _, test := range []struct {
		name, connections, accounts string
	}{
		{name: "duplicate connection", connections: `[{"id":"87b24961-b51e-4db8-9226-f198f6518a89","disabled":true},{"id":"87b24961-b51e-4db8-9226-f198f6518a89","disabled":true}]`},
		{name: "duplicate account", connections: fixtureBody(t, "success-connections.json"), accounts: `[{"id":"917c8734-8470-4a3e-a18f-57c3f2ee6631","brokerage_authorization":"87b24961-b51e-4db8-9226-f198f6518a89","name":"One","number":"12345","institution_name":"Broker","status":"open","sync_status":{},"balance":{},"is_paper":false},{"id":"917c8734-8470-4a3e-a18f-57c3f2ee6631","brokerage_authorization":"87b24961-b51e-4db8-9226-f198f6518a89","name":"Two","number":"67890","institution_name":"Broker","status":"open","sync_status":{},"balance":{},"is_paper":false}]`},
	} {
		t.Run(test.name, func(t *testing.T) {
			client, _ := NewInventoryClient(baseURL, roundTripFunc(func(request *http.Request) (*http.Response, error) {
				body := test.connections
				if strings.HasSuffix(request.URL.Path, "/accounts") {
					body = test.accounts
				}
				return jsonResponse(http.StatusOK, body), nil
			}), time.Now)
			_, err := client.Load(context.Background(), "access-token")
			var providerErr *portfolio.ProviderError
			if !errors.As(err, &providerErr) || providerErr.State != portfolio.StateMalformed {
				t.Fatalf("error=%v", err)
			}
		})
	}
}

func TestInventoryPreservesAllValidConnectionLifecycleRowsAfterFirstAccountFailure(t *testing.T) {
	baseURL, _ := url.Parse("https://api.snaptrade.example")
	connectionsJSON := `[
		{"id":"87b24961-b51e-4db8-9226-f198f6518a89","disabled":false,"brokerage":{"display_name":"First"}},
		{"disabled":false,"brokerage":{"display_name":"Malformed"}},
		{"id":"28d80e82-a089-4956-933a-efb7f6b8e099","disabled":false,"brokerage":{"display_name":"Later"}},
		{"id":"28d80e82-a089-4956-933a-efb7f6b8e099","disabled":false,"brokerage":{"display_name":"Duplicate"}},
		{"id":"a877a12a-bf36-414e-b41e-3c20b6866928","disabled":true,"brokerage":{"display_name":"Disabled"}}
	]`
	var paths []string
	client, _ := NewInventoryClient(baseURL, roundTripFunc(func(request *http.Request) (*http.Response, error) {
		paths = append(paths, request.URL.Path)
		if request.URL.Path == "/authorizations" {
			return jsonResponse(http.StatusOK, connectionsJSON), nil
		}
		return jsonResponse(http.StatusServiceUnavailable, `{}`), nil
	}), time.Now)
	connections, err := client.Load(context.Background(), "access-token")
	var providerErr *portfolio.ProviderError
	if !errors.As(err, &providerErr) || providerErr.State != portfolio.StateUnavailable {
		t.Fatalf("error=%v", err)
	}
	if len(connections) != 3 || connections[0].BrokerageLabel != "First" || connections[1].BrokerageLabel != "Later" || connections[2].Status != portfolio.ConnectionStatusDisabled {
		t.Fatalf("connections=%+v", connections)
	}
	if connections[0].Status != portfolio.ConnectionStatusUnavailable || connections[1].Status != portfolio.ConnectionStatusUnavailable {
		t.Fatalf("incomplete active connections=%+v", connections)
	}
	if len(paths) != 2 || !strings.HasSuffix(paths[1], "/87b24961-b51e-4db8-9226-f198f6518a89/accounts") {
		t.Fatalf("provider calls=%v", paths)
	}
}

func TestInventoryRejectsNullListsAndUnknownAccountStatus(t *testing.T) {
	baseURL, _ := url.Parse("https://api.snaptrade.example")
	for _, test := range []struct {
		name, connections, accounts string
	}{
		{name: "null connections", connections: `null`},
		{name: "null accounts", connections: fixtureBody(t, "success-connections.json"), accounts: `null`},
		{name: "unknown account status", connections: fixtureBody(t, "success-connections.json"), accounts: strings.Replace(fixtureBody(t, "success-accounts.json"), `"status": "open"`, `"status": "invented"`, 1)},
	} {
		t.Run(test.name, func(t *testing.T) {
			client, _ := NewInventoryClient(baseURL, roundTripFunc(func(request *http.Request) (*http.Response, error) {
				body := test.connections
				if strings.HasSuffix(request.URL.Path, "/accounts") {
					body = test.accounts
				}
				return jsonResponse(http.StatusOK, body), nil
			}), time.Now)
			connections, err := client.Load(context.Background(), "access-token")
			var providerErr *portfolio.ProviderError
			if !errors.As(err, &providerErr) || providerErr.State != portfolio.StateMalformed {
				t.Fatalf("connections=%+v err=%v", connections, err)
			}
			if test.name == "unknown account status" && (len(connections) != 1 || connections[0].Status != portfolio.ConnectionStatusUnavailable) {
				t.Fatalf("unknown status connection=%+v", connections)
			}
		})
	}
}

func TestInventoryMakesEveryRetainedAccountUnavailableWhenMalformedRowComesFirst(t *testing.T) {
	baseURL, _ := url.Parse("https://api.snaptrade.example")
	accounts := `[
		{"id":"00000000-0000-0000-0000-000000000000","brokerage_authorization":"87b24961-b51e-4db8-9226-f198f6518a89","name":"Malformed","number":"12345","institution_name":"Broker","status":"open","sync_status":{},"balance":{},"is_paper":false},
		{"id":"917c8734-8470-4a3e-a18f-57c3f2ee6631","brokerage_authorization":"87b24961-b51e-4db8-9226-f198f6518a89","name":"Valid","number":"67890","institution_name":"Broker","account_category":"INVESTMENT","raw_type":"Margin","status":"open","sync_status":{"holdings":{"initial_sync_completed":true}},"balance":{},"is_paper":false}
	]`
	client, _ := NewInventoryClient(baseURL, roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if strings.HasSuffix(request.URL.Path, "/accounts") {
			return jsonResponse(http.StatusOK, accounts), nil
		}
		return jsonResponse(http.StatusOK, fixtureBody(t, "success-connections.json")), nil
	}), time.Now)

	connections, err := client.Load(context.Background(), "access-token")
	var providerErr *portfolio.ProviderError
	if !errors.As(err, &providerErr) || providerErr.State != portfolio.StateMalformed {
		t.Fatalf("error=%v", err)
	}
	if len(connections) != 1 || connections[0].Status != portfolio.ConnectionStatusUnavailable || len(connections[0].Accounts) != 1 {
		t.Fatalf("connections=%+v", connections)
	}
	account := connections[0].Accounts[0]
	if account.Selectable || account.Eligible || account.UsabilityReason != portfolio.UsabilityConnectionUnavailable {
		t.Fatalf("retained account remained usable after partial normalization: %+v", account)
	}
}

func TestInventoryReturnsTrustworthyPartialConnectionWithCategoricalAccountFailure(t *testing.T) {
	baseURL, _ := url.Parse("https://api.snaptrade.example")
	for _, test := range []struct {
		name   string
		status int
		body   string
		want   portfolio.State
	}{
		{name: "transient", status: http.StatusServiceUnavailable, body: `{}`, want: portfolio.StateUnavailable},
		{name: "malformed", status: http.StatusOK, body: `{"not":"accounts"}`, want: portfolio.StateMalformed},
	} {
		t.Run(test.name, func(t *testing.T) {
			client, _ := NewInventoryClient(baseURL, roundTripFunc(func(request *http.Request) (*http.Response, error) {
				if strings.HasSuffix(request.URL.Path, "/accounts") {
					return jsonResponse(test.status, test.body), nil
				}
				return jsonResponse(http.StatusOK, fixtureBody(t, "success-connections.json")), nil
			}), time.Now)
			connections, err := client.Load(context.Background(), "access-token")
			var providerErr *portfolio.ProviderError
			if !errors.As(err, &providerErr) || providerErr.State != test.want || len(connections) != 1 || connections[0].Status != portfolio.ConnectionStatusUnavailable {
				t.Fatalf("connections=%+v err=%v", connections, err)
			}
		})
	}
}

func TestInventoryKeepsDisabledConnectionCategoricalAndSkipsAccounts(t *testing.T) {
	baseURL, _ := url.Parse("https://api.snaptrade.example")
	calls := 0
	client, _ := NewInventoryClient(baseURL, roundTripFunc(func(*http.Request) (*http.Response, error) {
		calls++
		return jsonResponse(http.StatusOK, fixtureBody(t, "disabled.json")), nil
	}), time.Now)
	connections, err := client.Load(context.Background(), "access-token")
	if err != nil || calls != 1 || len(connections) != 1 || connections[0].Status != "disabled" || connections[0].Available {
		t.Fatalf("connections=%+v calls=%d err=%v", connections, calls, err)
	}
}

func TestInventoryAcceptsEmptyFixtureWithoutAdditionalCalls(t *testing.T) {
	baseURL, _ := url.Parse("https://api.snaptrade.example")
	calls := 0
	client, _ := NewInventoryClient(baseURL, roundTripFunc(func(*http.Request) (*http.Response, error) {
		calls++
		return jsonResponse(http.StatusOK, fixtureBody(t, "empty.json")), nil
	}), time.Now)
	connections, err := client.Load(context.Background(), "access-token")
	if err != nil || calls != 1 || len(connections) != 0 {
		t.Fatalf("connections=%+v calls=%d err=%v", connections, calls, err)
	}
}

func TestInventoryClassifiesTransportFailureWithoutDetail(t *testing.T) {
	baseURL, _ := url.Parse("https://api.snaptrade.example")
	client, _ := NewInventoryClient(baseURL, roundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("token=private")
	}), time.Now)
	_, err := client.Load(context.Background(), "access-token")
	if err == nil || strings.Contains(err.Error(), "private") {
		t.Fatalf("error = %v", err)
	}
}

func jsonResponse(status int, body string) *http.Response {
	return &http.Response{StatusCode: status, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body))}
}

func fixtureBody(t *testing.T, name string) string {
	t.Helper()
	path := filepath.Join("..", "..", "..", "..", "test", "fixtures", "wiremock", "__files", "inventory", name)
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(body)
}

func timePointer(value time.Time) *time.Time { return &value }
