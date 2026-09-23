package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	providergenerated "github.com/kennethdavidbuck/findur/backend/internal/generated/providerapi"
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
	if strings.Join(paths, ",") != "/authorizations,/accounts" {
		t.Fatalf("request order = %v", paths)
	}
	if len(connections) != 1 || len(connections[0].Accounts) != 1 {
		t.Fatalf("connections = %+v", connections)
	}
	account := connections[0].Accounts[0]
	if account.MaskedLabel != "Retirement (•••• 8443)" || !account.Eligible || account.Category != "investment" || account.SyncState != "complete" ||
		account.TotalBalanceAmount == nil || *account.TotalBalanceAmount != "15363.23" || account.TotalBalanceCurrency == nil || *account.TotalBalanceCurrency != "CAD" {
		t.Fatalf("account = %+v", account)
	}
	serialized := strings.Join([]string{account.ID, string(account.Category), account.Type, account.MaskedLabel, string(account.SyncState)}, " ")
	for _, forbidden := range []string{"Q6542138443", "15363.23", "PRIVATE"} {
		if strings.Contains(serialized, forbidden) {
			t.Fatalf("normalized inventory retained %q: %s", forbidden, serialized)
		}
	}

	withoutTotal := strings.Replace(fixtureBody(t, "success-accounts.json"), `"balance": { "total": { "amount": 15363.23, "currency": "CAD" } }`, `"balance": {}`, 1)
	client, _ = NewInventoryClient(baseURL, roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if strings.HasSuffix(request.URL.Path, "/accounts") {
			return jsonResponse(http.StatusOK, withoutTotal), nil
		}
		return jsonResponse(http.StatusOK, fixtureBody(t, "success-connections.json")), nil
	}), time.Now)
	connections, err = client.Load(context.Background(), "access-token")
	if err != nil || connections[0].Accounts[0].TotalBalanceAmount != nil || connections[0].Accounts[0].TotalBalanceCurrency != nil {
		t.Fatalf("missing total connections=%+v err=%v", connections, err)
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
		eligible       bool
		selectable     bool
	}{
		{name: "missing status", accounts: strings.Replace(accounts, `"status": "open",`, "", 1), want: portfolio.UsabilityReady, eligible: true, selectable: true},
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
			var rows []providergenerated.Account
			if err := json.Unmarshal([]byte(test.accounts), &rows); err != nil || len(rows) != 1 {
				t.Fatalf("decode test account: %v", err)
			}
			account, valid := normalizeAccount(rows[0], rows[0].BrokerageAuthorization.String())
			if !valid {
				t.Fatal("account failed normalization")
			}
			if account.Selectable != test.selectable || account.Eligible != test.eligible || account.UsabilityReason != test.want {
				t.Fatalf("account=%+v want reason=%s", account, test.want)
			}
		})
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
		{name: "partial account total", connections: fixtureBody(t, "success-connections.json"), accounts: strings.Replace(fixtureBody(t, "success-accounts.json"), `"amount": 15363.23, "currency": "CAD"`, `"amount": 15363.23`, 1)},
		{name: "invalid account total", connections: fixtureBody(t, "success-connections.json"), accounts: strings.Replace(fixtureBody(t, "success-accounts.json"), `15363.23`, `"not-a-decimal"`, 1)},
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

func TestInventorySelectsInvestmentAccountsWithProviderMaskedNumbers(t *testing.T) {
	var rows []string
	for index, accountType := range []string{"NP", "Fidelity Credit Card", "NP", "TODJ"} {
		status, category := `"open"`, "INVESTMENT"
		if index == 1 {
			status, category = "null", "LOC"
		}
		rows = append(rows, fmt.Sprintf(`{
			"id":"10000000-0000-4000-8000-%012d",
			"brokerage_authorization":"87b24961-b51e-4db8-9226-f198f6518a89",
			"name":"Synthetic account %d","number":"*****%04d",
			"created_date":"2026-05-09T01:44:59.843365Z","funding_date":null,"opening_date":null,
			"sync_status":{"holdings":{"last_successful_sync":"2026-09-21T10:45:43.428758+00:00","initial_sync_completed":true},
			"transactions":{"last_successful_sync":"2026-09-20","first_transaction_date":null,"initial_sync_completed":true}},
			"balance":{"total":{"amount":100,"currency":"USD"}},"raw_data":null,
			"raw_type":%q,"status":%s,"is_paper":false,"account_category":%q
		}`, index+1, index+1, 1001+index, accountType, status, category))
	}
	baseURL, _ := url.Parse("https://api.snaptrade.example")
	client, _ := NewInventoryClient(baseURL, roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if strings.HasSuffix(request.URL.Path, "/accounts") {
			return jsonResponse(http.StatusOK, "["+strings.Join(rows, ",")+"]"), nil
		}
		return jsonResponse(http.StatusOK, fixtureBody(t, "success-connections.json")), nil
	}), time.Now)
	connections, err := client.Load(context.Background(), "access-token")
	if err != nil || len(connections) != 1 || len(connections[0].Accounts) != 4 {
		t.Fatalf("connections=%+v error=%v", connections, err)
	}
	selectable := make([]portfolio.Account, 0, 3)
	for _, account := range connections[0].Accounts {
		if account.Selectable {
			selectable = append(selectable, account)
			continue
		}
		if account.UsabilityReason != portfolio.UsabilityUnsupportedCategory {
			t.Fatalf("unexpected passive account=%+v", account)
		}
	}
	for index, account := range selectable {
		sourceIndex := []int{0, 2, 3}[index]
		wantLabel := fmt.Sprintf("Synthetic account %d (•••• %04d)", sourceIndex+1, 1001+sourceIndex)
		if account.ID != globalAccountID(sourceIndex+1) || account.MaskedLabel != wantLabel || !account.Selectable || !account.Eligible || account.UsabilityReason != portfolio.UsabilityReady ||
			account.TotalBalanceAmount == nil || *account.TotalBalanceAmount != "100" || account.TotalBalanceCurrency == nil || *account.TotalBalanceCurrency != "USD" {
			t.Fatalf("account=%+v want label=%q", account, wantLabel)
		}
	}
}

func TestInventoryGroupsGlobalAccountsInTwoRequests(t *testing.T) {
	// More than 500 accounts and more than 1 MiB exercise the global inventory
	// bounds without increasing the bounds on other provider responses.
	var rows []string
	for index := range maxAccounts {
		connection := index%3 + 1
		rows = append(rows, globalAccountRow(index+1, connection, strings.Repeat("Synthetic ", 40)))
	}
	body := "[" + strings.Join(rows, ",") + "]"
	if len(body) <= providerResponseLimit {
		t.Fatal("large global fixture must exceed the ordinary response limit")
	}
	client, paths := globalInventoryClient(t, globalConnections(), body, http.StatusOK)
	connections, err := client.Load(context.Background(), "access-token")
	if err != nil || len(connections) != 4 || strings.Join(*paths, ",") != "/authorizations,/accounts" {
		t.Fatalf("connections=%d paths=%v err=%v", len(connections), *paths, err)
	}
	count := 0
	for index, connection := range connections {
		if index == 3 {
			if connection.Status != portfolio.ConnectionStatusDisabled || len(connection.Accounts) != 0 {
				t.Fatalf("disabled connection=%+v", connection)
			}
			continue
		}
		if connection.Status != portfolio.ConnectionStatusActive {
			t.Fatalf("healthy connection=%+v", connection)
		}
		for _, account := range connection.Accounts {
			if !account.Selectable || account.UsabilityReason != portfolio.UsabilityReady {
				t.Fatalf("account=%+v", account)
			}
			count++
		}
	}
	if count != maxAccounts {
		t.Fatalf("account count=%d want=%d", count, maxAccounts)
	}
}

func TestInventoryPublishesAccountAvailabilityWithoutMakingUnavailableAccountsSelectable(t *testing.T) {
	base := globalAccountRow(20, 1, "Synthetic candidate")
	for _, test := range []struct {
		name, row  string
		accepted   bool
		wantReason portfolio.UsabilityReason
	}{
		{name: "open investment", row: base, accepted: true},
		{name: "cash investment", row: strings.Replace(base, `"name":`, `"raw_type":"Cash","name":`, 1), accepted: true},
		{name: "transactions pending with usable holdings", row: strings.Replace(base, `"holdings":`, `"transactions":{"initial_sync_completed":false},"holdings":`, 1), accepted: true},
		{name: "holdings explicitly available", row: strings.Replace(base, `"initial_sync_completed":true`, `"initial_sync_completed":true,"holdings_unavailable":false`, 1), accepted: true},
		{name: "null status investment", row: strings.Replace(base, `"open"`, `null`, 1), accepted: true},
		{name: "null status and null category", row: strings.NewReplacer(`"open"`, `null`, `"INVESTMENT"`, `null`).Replace(base), accepted: true, wantReason: portfolio.UsabilityProvisionalStatus},
		{name: "null status and missing category", row: strings.NewReplacer(`"open"`, `null`, `"account_category":"INVESTMENT",`, ``).Replace(base), accepted: true, wantReason: portfolio.UsabilityProvisionalStatus},
		{name: "null status and future category", row: strings.NewReplacer(`"open"`, `null`, `"INVESTMENT"`, `"FUTURE"`).Replace(base)},
		{name: "missing status investment", row: strings.Replace(base, `"status":"open",`, ``, 1), accepted: true},
		{name: "closed", row: strings.Replace(base, `"open"`, `"closed"`, 1)},
		{name: "archived", row: strings.Replace(base, `"open"`, `"archived"`, 1)},
		{name: "unavailable", row: strings.Replace(base, `"open"`, `"unavailable"`, 1)},
		{name: "deposit", row: strings.Replace(base, `"INVESTMENT"`, `"DEPOSIT"`, 1)},
		{name: "credit", row: strings.Replace(base, `"INVESTMENT"`, `"LOC"`, 1)},
		{name: "null category", row: strings.Replace(base, `"INVESTMENT"`, `null`, 1), accepted: true, wantReason: portfolio.UsabilityProvisionalCategory},
		{name: "future category", row: strings.Replace(base, `"INVESTMENT"`, `"FUTURE"`, 1)},
		{name: "missing category", row: strings.Replace(base, `"account_category":"INVESTMENT",`, ``, 1), accepted: true, wantReason: portfolio.UsabilityProvisionalCategory},
		{name: "missing holdings", row: strings.Replace(base, `"holdings":{"initial_sync_completed":true}`, ``, 1)},
		{name: "missing initial sync flag", row: strings.Replace(base, `"initial_sync_completed":true`, ``, 1)},
		{name: "null initial sync flag", row: strings.Replace(base, `"initial_sync_completed":true`, `"initial_sync_completed":null`, 1)},
		{name: "pending holdings", row: strings.Replace(base, `"initial_sync_completed":true`, `"initial_sync_completed":false`, 1)},
		{name: "unavailable holdings despite complete sync", row: strings.Replace(base, `"initial_sync_completed":true`, `"initial_sync_completed":true,"holdings_unavailable":true`, 1)},
		{name: "null status and pending holdings", row: strings.NewReplacer(`"open"`, `null`, `"initial_sync_completed":true`, `"initial_sync_completed":false`).Replace(base)},
		{name: "disabled connection", row: strings.Replace(base, globalConnectionID(1), globalConnectionID(4), 1)},
	} {
		t.Run(test.name, func(t *testing.T) {
			body := "[" + globalAccountRow(1, 1, "Healthy baseline") + "," + test.row + "]"
			client, paths := globalInventoryClient(t, globalConnections(), body, http.StatusOK)
			connections, err := client.Load(context.Background(), "access-token")
			candidateConnection := 0
			if test.name == "disabled connection" {
				candidateConnection = 3
			}
			if err != nil || len(connections) != 4 || len(connections[0].Accounts) != 2-(candidateConnection/3) || len(connections[3].Accounts) != candidateConnection/3 || strings.Join(*paths, ",") != "/authorizations,/accounts" {
				t.Fatalf("connections=%+v paths=%v err=%v", connections, *paths, err)
			}
			candidate := connections[candidateConnection].Accounts[len(connections[candidateConnection].Accounts)-1]
			if candidate.ID != globalAccountID(20) || candidate.Selectable != test.accepted {
				t.Fatalf("candidate account=%+v accepted=%t", candidate, test.accepted)
			}
			if test.wantReason != "" && candidate.UsabilityReason != test.wantReason {
				t.Fatalf("candidate reason=%q want=%q", candidate.UsabilityReason, test.wantReason)
			}
			if !test.accepted && candidate.UsabilityReason == portfolio.UsabilityReady {
				t.Fatalf("unavailable candidate lacks passive reason: %+v", candidate)
			}
		})
	}
}

func TestInventoryRejectsMalformedGlobalAccountsAtomically(t *testing.T) {
	base := globalAccountRow(20, 2, "Malformed candidate")
	for _, test := range []struct {
		name, rows string
	}{
		{name: "unknown connection", rows: strings.Replace(base, globalConnectionID(2), globalConnectionID(9), 1)},
		{name: "invalid connection UUID", rows: strings.Replace(base, globalConnectionID(2), "not-a-uuid", 1)},
		{name: "invalid account UUID", rows: strings.Replace(base, globalAccountID(20), "not-a-uuid", 1)},
		{name: "zero account UUID", rows: strings.Replace(base, globalAccountID(20), "00000000-0000-0000-0000-000000000000", 1)},
		{name: "invalid date", rows: strings.Replace(base, `"name":`, `"created_date":"not-a-date","name":`, 1)},
		{name: "invalid number type", rows: strings.Replace(base, `"number":"*****1001"`, `"number":1001`, 1)},
		{name: "unknown status", rows: strings.Replace(base, `"open"`, `"invented"`, 1)},
		{name: "duplicate account", rows: base + "," + base},
		{name: "duplicate across connections", rows: base + "," + strings.Replace(base, globalConnectionID(2), globalConnectionID(3), 1)},
		{name: "duplicate excluded account", rows: base + "," + strings.Replace(base, `"INVESTMENT"`, `"LOC"`, 1)},
		{name: "malformed disabled row", rows: strings.NewReplacer(globalConnectionID(2), globalConnectionID(4), globalAccountID(20), "not-a-uuid").Replace(base)},
	} {
		t.Run(test.name, func(t *testing.T) {
			body := "[" + globalAccountRow(1, 1, "Healthy first") + "," + test.rows + "," + globalAccountRow(3, 3, "Healthy last") + "]"
			client, paths := globalInventoryClient(t, globalConnections(), body, http.StatusOK)
			connections, err := client.Load(context.Background(), "access-token")
			var providerErr *portfolio.ProviderError
			if !errors.As(err, &providerErr) || providerErr.State != portfolio.StateMalformed || len(connections) != 0 || strings.Join(*paths, ",") != "/authorizations,/accounts" {
				t.Fatalf("partial inventory escaped: connections=%+v paths=%v err=%v", connections, *paths, err)
			}
		})
	}
}

func TestInventoryGlobalAccountFetchFailuresPublishNothing(t *testing.T) {
	for _, test := range []struct {
		name   string
		status int
		body   string
		want   portfolio.State
	}{
		{name: "unauthorized", status: http.StatusUnauthorized, want: portfolio.StateUnauthorized},
		{name: "rate limited", status: http.StatusTooManyRequests, want: portfolio.StateRateLimited},
		{name: "forbidden", status: http.StatusForbidden, want: portfolio.StateUnavailable},
		{name: "transient", status: http.StatusServiceUnavailable, want: portfolio.StateUnavailable},
		{name: "transport", want: portfolio.StateUnavailable},
		{name: "null", status: http.StatusOK, body: `null`, want: portfolio.StateMalformed},
		{name: "malformed", status: http.StatusOK, body: `{}`, want: portfolio.StateMalformed},
	} {
		t.Run(test.name, func(t *testing.T) {
			client, paths := globalInventoryClient(t, globalConnections(), test.body, test.status)
			connections, err := client.Load(context.Background(), "access-token")
			var providerErr *portfolio.ProviderError
			if !errors.As(err, &providerErr) || providerErr.State != test.want || len(connections) != 0 || strings.Join(*paths, ",") != "/authorizations,/accounts" {
				t.Fatalf("connections=%+v paths=%v error=%v", connections, *paths, err)
			}
			if test.want == portfolio.StateRateLimited && providerErr.RetryAt == nil {
				t.Fatal("missing provider retry timing")
			}

		})
	}
}

func TestInventoryGlobalAccountBounds(t *testing.T) {
	t.Run("account count", func(t *testing.T) {
		rows := make([]string, maxAccounts+1)
		for index := range rows {
			rows[index] = globalAccountRow(index+1, 1, "Synthetic")
		}
		assertMalformedGlobalInventory(t, globalConnections(), "["+strings.Join(rows, ",")+"]")
	})
	t.Run("account pagination envelope", func(t *testing.T) {
		assertMalformedGlobalInventory(t, globalConnections(), `{"results":[`+globalAccountRow(1, 1, "First page")+`],"next":"next-page","pagination":{"has_more":true}}`)
	})
	t.Run("connection pagination envelope", func(t *testing.T) {
		assertMalformedGlobalInventory(t, `{"results":`+globalConnections()+`,"next":"next-page","pagination":{"has_more":true}}`, `[]`)
	})
}

func TestInventoryResponseByteBoundaries(t *testing.T) {
	for _, test := range []struct {
		name     string
		accounts bool
		limit    int
	}{
		{name: "connections", limit: providerResponseLimit},
		{name: "accounts", accounts: true, limit: accountListResponseLimit},
	} {
		for _, extra := range []int{0, 1} {
			t.Run(fmt.Sprintf("%s limit plus %d", test.name, extra), func(t *testing.T) {
				connectionsBody := globalConnections()
				accountsBody := "[" + globalAccountRow(1, 1, "Valid account") + "]"
				if test.accounts {
					accountsBody += strings.Repeat(" ", test.limit+extra-len(accountsBody))
				} else {
					connectionsBody += strings.Repeat(" ", test.limit+extra-len(connectionsBody))
				}
				client, paths := globalInventoryClient(t, connectionsBody, accountsBody, http.StatusOK)
				connections, err := client.Load(context.Background(), "access-token")
				if extra == 0 {
					if err != nil || len(connections) != 4 || len(connections[0].Accounts) != 1 || !connections[0].Accounts[0].Selectable || strings.Join(*paths, ",") != "/authorizations,/accounts" {
						t.Fatalf("exact-boundary valid response rejected: connections=%+v paths=%v err=%v", connections, *paths, err)
					}
					return
				}
				wantPaths := "/authorizations"
				if test.accounts {
					wantPaths += ",/accounts"
				}
				var providerErr *portfolio.ProviderError
				if !errors.As(err, &providerErr) || providerErr.State != portfolio.StateMalformed || len(connections) != 0 || strings.Join(*paths, ",") != wantPaths {
					t.Fatalf("over-boundary response accepted: connections=%+v paths=%v err=%v", connections, *paths, err)
				}
			})
		}
	}
}

func TestInventoryConnectionCountBoundaries(t *testing.T) {
	for _, count := range []int{maxConnections, maxConnections + 1} {
		t.Run(fmt.Sprintf("%d connections", count), func(t *testing.T) {
			rows := make([]string, count)
			for index := range rows {
				rows[index] = fmt.Sprintf(`{"id":%q,"disabled":false}`, globalConnectionID(index+1))
			}
			client, paths := globalInventoryClient(t, "["+strings.Join(rows, ",")+"]", "["+globalAccountRow(1, count, "Last connection account")+"]", http.StatusOK)
			connections, err := client.Load(context.Background(), "access-token")
			if count == maxConnections {
				if err != nil || len(connections) != count || len(connections[count-1].Accounts) != 1 || !connections[count-1].Accounts[0].Selectable || strings.Join(*paths, ",") != "/authorizations,/accounts" {
					t.Fatalf("valid connection limit rejected: count=%d paths=%v err=%v", len(connections), *paths, err)
				}
				return
			}
			var providerErr *portfolio.ProviderError
			if !errors.As(err, &providerErr) || providerErr.State != portfolio.StateMalformed || len(connections) != 0 || strings.Join(*paths, ",") != "/authorizations" {
				t.Fatalf("connection overflow was not rejected before accounts: count=%d paths=%v err=%v", len(connections), *paths, err)
			}
		})
	}
}

func TestInventoryRejectsAmbiguousConnectionsAtomically(t *testing.T) {
	for _, test := range []struct {
		name, row string
	}{
		{name: "duplicate connection", row: fmt.Sprintf(`{"id":%q,"disabled":true}`, globalConnectionID(1))},
		{name: "missing connection ID", row: `{"disabled":false}`},
	} {
		t.Run(test.name, func(t *testing.T) {
			connectionsBody := strings.TrimSuffix(globalConnections(), "]") + "," + test.row + "]"
			client, paths := globalInventoryClient(t, connectionsBody, `[]`, http.StatusOK)
			connections, err := client.Load(context.Background(), "access-token")
			var providerErr *portfolio.ProviderError
			if !errors.As(err, &providerErr) || providerErr.State != portfolio.StateMalformed || len(connections) != 0 || strings.Join(*paths, ",") != "/authorizations" {
				t.Fatalf("connections=%+v paths=%v err=%v", connections, *paths, err)
			}
		})
	}
}

func assertMalformedGlobalInventory(t *testing.T, connectionsBody, accountsBody string) {
	t.Helper()
	client, _ := globalInventoryClient(t, connectionsBody, accountsBody, http.StatusOK)
	connections, err := client.Load(context.Background(), "access-token")
	var providerErr *portfolio.ProviderError
	if !errors.As(err, &providerErr) || providerErr.State != portfolio.StateMalformed {
		t.Fatalf("error=%v", err)
	}
	for _, connection := range connections {
		if len(connection.Accounts) != 0 {
			t.Fatalf("out-of-bounds inventory retained accounts: %+v", connection)
		}
	}
}

func globalInventoryClient(t *testing.T, connectionsBody, accountsBody string, accountsStatus int) (*InventoryClient, *[]string) {
	t.Helper()
	baseURL, _ := url.Parse("https://api.snaptrade.example")
	var paths []string
	client, err := NewInventoryClient(baseURL, roundTripFunc(func(request *http.Request) (*http.Response, error) {
		paths = append(paths, request.URL.Path)
		if request.Header.Get("Authorization") != "Bearer access-token" || request.URL.RawQuery != "" {
			t.Fatalf("unexpected request credentials or query")
		}
		switch request.URL.Path {
		case "/authorizations":
			return jsonResponse(http.StatusOK, connectionsBody), nil
		case "/accounts":
			if accountsStatus == 0 {
				return nil, errors.New("synthetic transport failure")
			}
			return jsonResponse(accountsStatus, accountsBody), nil
		default:
			t.Fatalf("unexpected provider path=%q", request.URL.Path)
			return nil, errors.New("unexpected provider path")
		}
	}), time.Now)
	if err != nil {
		t.Fatal(err)
	}
	return client, &paths
}

func globalConnections() string {
	var rows []string
	for index := range 4 {
		rows = append(rows, fmt.Sprintf(`{"id":%q,"disabled":%t}`, globalConnectionID(index+1), index == 3))
	}
	return "[" + strings.Join(rows, ",") + "]"
}

func globalAccountRow(id, connection int, name string) string {
	return fmt.Sprintf(`{"id":%q,"brokerage_authorization":%q,"name":%q,"number":"*****1001","status":"open","account_category":"INVESTMENT","sync_status":{"holdings":{"initial_sync_completed":true}}}`, globalAccountID(id), globalConnectionID(connection), name)
}

func globalAccountID(index int) string    { return fmt.Sprintf("10000000-0000-4000-8000-%012d", index) }
func globalConnectionID(index int) string { return fmt.Sprintf("00000000-0000-4000-8000-%012d", index) }

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
			if len(connections) != 0 {
				t.Fatalf("unknown status connection=%+v", connections)
			}
		})
	}
}

func TestInventoryAccountFailurePublishesNoPartialConnections(t *testing.T) {
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
			if !errors.As(err, &providerErr) || providerErr.State != test.want || len(connections) != 0 {
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

func TestRequestGateHonorsRetryAfterAndFailureThresholds(t *testing.T) {
	now := time.Date(2026, 9, 21, 12, 0, 0, 0, time.UTC)
	gate := &requestGate{}
	gate.record(http.StatusTooManyRequests, "90", now)
	if err := gate.wait(context.Background(), now.Add(89*time.Second)); err == nil {
		t.Fatal("429 circuit admitted before Retry-After")
	}
	if err := gate.wait(context.Background(), now.Add(90*time.Second)); err != nil {
		t.Fatalf("429 circuit remained open: %v", err)
	}

	dateGate := &requestGate{}
	retryAt := now.Add(2 * time.Minute)
	dateGate.record(http.StatusTooManyRequests, retryAt.Format(http.TimeFormat), now)
	if !dateGate.openUntil.Equal(retryAt) {
		t.Fatalf("date Retry-After=%v want=%v", dateGate.openUntil, retryAt)
	}

	failureGate := &requestGate{}
	failureGate.record(http.StatusBadGateway, "", now)
	failureGate.record(0, "", now)
	if err := failureGate.wait(context.Background(), now); err != nil {
		t.Fatalf("circuit opened before threshold: %v", err)
	}
	failureGate.record(http.StatusServiceUnavailable, "", now)
	if err := failureGate.wait(context.Background(), now); err == nil {
		t.Fatal("repeated transport/5xx failures did not open circuit")
	}
	if err := failureGate.wait(context.Background(), now.Add(30*time.Second)); err != nil {
		t.Fatalf("failure circuit did not recover: %v", err)
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

func TestConnectionDiagnosticUsesBoundedReasonAndActionMatrix(t *testing.T) {
	tests := []struct {
		name       string
		connection portfolio.Connection
		returned   int
		reasons    map[portfolio.UsabilityReason]int
		wantReason portfolio.ResourceDiagnosticReason
		wantAction portfolio.ResourceDiagnosticAction
		wantNil    bool
	}{
		{name: "active usable", connection: portfolio.Connection{Status: portfolio.ConnectionStatusActive, Accounts: []portfolio.Account{{ID: "usable", Selectable: true}}}, returned: 1, wantNil: true},
		{name: "no accounts", connection: portfolio.Connection{Status: portfolio.ConnectionStatusActive}, wantReason: portfolio.DiagnosticNoAccountsReturned, wantAction: portfolio.DiagnosticActionRetry},
		{name: "unsupported only", connection: portfolio.Connection{Status: portfolio.ConnectionStatusActive}, returned: 1, reasons: map[portfolio.UsabilityReason]int{portfolio.UsabilityUnsupportedCategory: 1}, wantReason: portfolio.DiagnosticNoSupportedAccounts, wantAction: portfolio.DiagnosticActionNone},
		{name: "disabled", connection: portfolio.Connection{Status: portfolio.ConnectionStatusDisabled}, wantReason: portfolio.DiagnosticConnectionDisabled, wantAction: portfolio.DiagnosticActionReconnect},
		{name: "holdings pending", connection: portfolio.Connection{Status: portfolio.ConnectionStatusActive}, returned: 1, reasons: map[portfolio.UsabilityReason]int{portfolio.UsabilitySyncPending: 1}, wantReason: portfolio.DiagnosticSyncPending, wantAction: portfolio.DiagnosticActionWait},
		{name: "provider unavailable", connection: portfolio.Connection{Status: portfolio.ConnectionStatusActive}, returned: 1, reasons: map[portfolio.UsabilityReason]int{portfolio.UsabilitySyncUnavailable: 1}, wantReason: portfolio.DiagnosticProviderUnavailable, wantAction: portfolio.DiagnosticActionRetry},
		{name: "mixed pending and unavailable", connection: portfolio.Connection{Status: portfolio.ConnectionStatusActive}, returned: 2, reasons: map[portfolio.UsabilityReason]int{portfolio.UsabilitySyncPending: 1, portfolio.UsabilitySyncUnavailable: 1}, wantReason: portfolio.DiagnosticUnknown, wantAction: portfolio.DiagnosticActionRetry},
		{name: "mixed pending and unsupported", connection: portfolio.Connection{Status: portfolio.ConnectionStatusActive}, returned: 2, reasons: map[portfolio.UsabilityReason]int{portfolio.UsabilitySyncPending: 1, portfolio.UsabilityUnsupportedCategory: 1}, wantReason: portfolio.DiagnosticUnknown, wantAction: portfolio.DiagnosticActionRetry},
		{name: "mixed unavailable and unsupported", connection: portfolio.Connection{Status: portfolio.ConnectionStatusActive}, returned: 2, reasons: map[portfolio.UsabilityReason]int{portfolio.UsabilityAccountUnavailable: 1, portfolio.UsabilityUnsupportedCategory: 1}, wantReason: portfolio.DiagnosticUnknown, wantAction: portfolio.DiagnosticActionRetry},
		{name: "unknown combination", connection: portfolio.Connection{Status: portfolio.ConnectionStatusActive}, returned: 1, reasons: map[portfolio.UsabilityReason]int{portfolio.UsabilityAccountClosed: 1}, wantReason: portfolio.DiagnosticUnknown, wantAction: portfolio.DiagnosticActionRetry},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			diagnostic := connectionDiagnostic(test.connection, test.returned, test.reasons)
			if test.wantNil {
				if diagnostic != nil {
					t.Fatalf("diagnostic=%+v want nil", diagnostic)
				}
				return
			}
			if diagnostic == nil || diagnostic.Reason != test.wantReason || diagnostic.RecommendedAction != test.wantAction {
				t.Fatalf("diagnostic=%+v want reason=%q action=%q", diagnostic, test.wantReason, test.wantAction)
			}
		})
	}
}

func TestDefaultStackFixtureContainsRichScenarioInventory(t *testing.T) {
	var rawConnections []providergenerated.BrokerageAuthorization
	if err := json.Unmarshal([]byte(fixtureBody(t, "stack-connections.json")), &rawConnections); err != nil {
		t.Fatal(err)
	}
	var rawAccounts []rawInventoryAccount
	if err := json.Unmarshal([]byte(fixtureBody(t, "stack-accounts.json")), &rawAccounts); err != nil {
		t.Fatal(err)
	}
	connections, err := normalizeConnections(rawConnections)
	if err != nil {
		t.Fatal(err)
	}
	if err := groupAccounts(connections, rawAccounts); err != nil {
		t.Fatal(err)
	}
	selectable := 0
	labels := make(map[string]bool)
	passiveLabels := make(map[string]portfolio.UsabilityReason)
	for _, connection := range connections {
		for _, account := range connection.Accounts {
			if account.Selectable {
				selectable++
				labels[account.MaskedLabel] = true
			} else {
				passiveLabels[account.MaskedLabel] = account.UsabilityReason
			}
		}
	}
	if len(connections) != 22 || len(rawAccounts) != 29 || selectable != 26 {
		t.Fatalf("connections=%d accounts=%d selectable=%d", len(connections), len(rawAccounts), selectable)
	}
	for _, label := range []string{
		"Healthy Realtime — Full Data (•••• X001)",
		"Cash Only — No Positions (•••• CASH)",
		"Positions — Refresh Failed (•••• 0004)",
	} {
		if !labels[label] {
			t.Errorf("missing selectable scenario label %q", label)
		}
	}
	for label, reason := range map[string]portfolio.UsabilityReason{
		"Unsupported Account Type (•••• 0021)":   portfolio.UsabilityUnsupportedCategory,
		"Connection Repair Required (•••• 0023)": portfolio.UsabilityConnectionDisabled,
	} {
		if passiveLabels[label] != reason {
			t.Errorf("passive scenario %q reason=%q want=%q", label, passiveLabels[label], reason)
		}
	}
}

type wireMockContractMapping struct {
	Priority int `json:"priority"`
	Request  struct {
		URLPath        string `json:"urlPath"`
		URLPathPattern string `json:"urlPathPattern"`
	} `json:"request"`
	Response struct {
		Status       int             `json:"status"`
		BodyFileName string          `json:"bodyFileName"`
		JSONBody     json.RawMessage `json:"jsonBody"`
	} `json:"response"`
}

func TestDefaultStackFixtureMapsEverySelectableAccountResource(t *testing.T) {
	var contract struct {
		Mappings []wireMockContractMapping `json:"mappings"`
	}
	contractPath := filepath.Join("..", "..", "..", "..", "test", "fixtures", "wiremock", "mappings", "portfolio-inventory.json")
	body, err := os.ReadFile(contractPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(body, &contract); err != nil {
		t.Fatal(err)
	}
	var rawConnections []providergenerated.BrokerageAuthorization
	if err := json.Unmarshal([]byte(fixtureBody(t, "stack-connections.json")), &rawConnections); err != nil {
		t.Fatal(err)
	}
	var rawAccounts []rawInventoryAccount
	if err := json.Unmarshal([]byte(fixtureBody(t, "stack-accounts.json")), &rawAccounts); err != nil {
		t.Fatal(err)
	}
	connections, err := normalizeConnections(rawConnections)
	if err != nil {
		t.Fatal(err)
	}
	if err := groupAccounts(connections, rawAccounts); err != nil {
		t.Fatal(err)
	}

	resources := map[string]string{"balances": "balances", "positions": "positions/all", "activities": "activities"}
	for _, connection := range connections {
		for _, account := range connection.Accounts {
			if !account.Selectable {
				continue
			}
			for resource, suffix := range resources {
				path := "/accounts/" + account.ID + "/" + suffix
				matched, found := resolveWireMockMapping(t, contract.Mappings, path)
				if !found {
					t.Errorf("selectable account %q has no %s route", account.MaskedLabel, resource)
					continue
				}
				got := wireMockResponseKind(t, matched.Response.Status, matched.Response.BodyFileName, matched.Response.JSONBody, resource)
				want := defaultScenarioResourceKind(account.ID, resource)
				if got != want {
					t.Errorf("account %q %s resolved %q, want %q", account.MaskedLabel, resource, got, want)
				}
			}
		}
	}
}

func resolveWireMockMapping(t *testing.T, mappings []wireMockContractMapping, path string) (wireMockContractMapping, bool) {
	t.Helper()
	var selected wireMockContractMapping
	selectedPriority := int(^uint(0) >> 1)
	found := false
	for _, mapping := range mappings {
		matches := mapping.Request.URLPath == path
		if !matches && mapping.Request.URLPathPattern != "" {
			matches = regexp.MustCompile("^(?:" + mapping.Request.URLPathPattern + ")$").MatchString(path)
		}
		priority := mapping.Priority
		if priority == 0 {
			priority = 5
		}
		if matches && priority < selectedPriority {
			selected, selectedPriority, found = mapping, priority, true
		}
	}
	return selected, found
}

func wireMockResponseKind(t *testing.T, status int, bodyFileName string, raw json.RawMessage, resource string) string {
	t.Helper()
	if status >= http.StatusBadRequest {
		return "named-failure"
	}
	if bodyFileName != "" {
		path := filepath.Join("..", "..", "..", "..", "test", "fixtures", "wiremock", "__files", bodyFileName)
		var err error
		raw, err = os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
	}
	if resource == "balances" {
		var rows []json.RawMessage
		if err := json.Unmarshal(raw, &rows); err != nil {
			t.Fatal(err)
		}
		if len(rows) > 0 {
			return "populated"
		}
		return "empty"
	}
	var envelope map[string]json.RawMessage
	if err := json.Unmarshal(raw, &envelope); err != nil {
		t.Fatal(err)
	}
	key := "results"
	if resource == "activities" {
		key = "data"
	}
	var rows []json.RawMessage
	if err := json.Unmarshal(envelope[key], &rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) > 0 {
		return "populated"
	}
	return "empty"
}

func defaultScenarioResourceKind(accountID, resource string) string {
	if accountID == "b0000000-0000-4000-8000-000000000001" && resource == "positions" ||
		accountID == "b0000000-0000-4000-8000-000000000003" && resource == "activities" ||
		accountID == "b0000000-0000-4000-8000-000000000004" && resource == "balances" {
		return "named-failure"
	}
	if accountID == "03867fbb-41b4-4a05-8815-c96f94f8ba6b" {
		return "populated"
	}
	if accountID == "7e7dcb86-7d52-4f46-8fcf-91d5c9f81629" || accountID == "50bb0405-5efd-473f-a742-78a82bb1db53" {
		if resource == "balances" {
			return "populated"
		}
		return "empty"
	}
	number, err := strconv.Atoi(strings.TrimPrefix(accountID, "b0000000-0000-4000-8000-000000000"))
	if err != nil {
		return "empty"
	}
	if resource == "balances" && number >= 1 && number <= 9 ||
		resource == "positions" && (number >= 1 && number <= 5 || number >= 11 && number <= 13) ||
		resource == "activities" && number >= 6 && number <= 14 {
		return "populated"
	}
	return "empty"
}
