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
		{name: "missing status", accounts: strings.Replace(accounts, `"status": "open",`, "", 1), want: portfolio.UsabilityReady, selectable: true},
		{name: "missing category", accounts: strings.Replace(accounts, `"account_category": "INVESTMENT",`, "", 1), want: portfolio.UsabilityProvisionalCategory},
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
			if account.Selectable != test.selectable || account.Eligible != test.selectable || account.UsabilityReason != test.want {
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
	if err != nil || len(connections) != 1 || len(connections[0].Accounts) != 3 {
		t.Fatalf("connections=%+v error=%v", connections, err)
	}
	for index, account := range connections[0].Accounts {
		sourceIndex := []int{0, 2, 3}[index]
		wantLabel := fmt.Sprintf("Synthetic account %d (•••• %04d)", sourceIndex+1, 1001+sourceIndex)
		if account.ID != globalAccountID(sourceIndex+1) || account.MaskedLabel != wantLabel || !account.Selectable || !account.Eligible || account.UsabilityReason != portfolio.UsabilityReady {
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

func TestInventoryPublishesOnlyUsableInvestmentAccounts(t *testing.T) {
	base := globalAccountRow(20, 1, "Synthetic candidate")
	for _, test := range []struct {
		name, row string
		accepted  bool
	}{
		{name: "open investment", row: base, accepted: true},
		{name: "cash investment", row: strings.Replace(base, `"name":`, `"raw_type":"Cash","name":`, 1), accepted: true},
		{name: "transactions pending with usable holdings", row: strings.Replace(base, `"holdings":`, `"transactions":{"initial_sync_completed":false},"holdings":`, 1), accepted: true},
		{name: "holdings explicitly available", row: strings.Replace(base, `"initial_sync_completed":true`, `"initial_sync_completed":true,"holdings_unavailable":false`, 1), accepted: true},
		{name: "null status investment", row: strings.Replace(base, `"open"`, `null`, 1), accepted: true},
		{name: "null status and null category", row: strings.NewReplacer(`"open"`, `null`, `"INVESTMENT"`, `null`).Replace(base)},
		{name: "null status and missing category", row: strings.NewReplacer(`"open"`, `null`, `"account_category":"INVESTMENT",`, ``).Replace(base)},
		{name: "null status and future category", row: strings.NewReplacer(`"open"`, `null`, `"INVESTMENT"`, `"FUTURE"`).Replace(base)},
		{name: "missing status investment", row: strings.Replace(base, `"status":"open",`, ``, 1), accepted: true},
		{name: "closed", row: strings.Replace(base, `"open"`, `"closed"`, 1)},
		{name: "archived", row: strings.Replace(base, `"open"`, `"archived"`, 1)},
		{name: "unavailable", row: strings.Replace(base, `"open"`, `"unavailable"`, 1)},
		{name: "deposit", row: strings.Replace(base, `"INVESTMENT"`, `"DEPOSIT"`, 1)},
		{name: "credit", row: strings.Replace(base, `"INVESTMENT"`, `"LOC"`, 1)},
		{name: "null category", row: strings.Replace(base, `"INVESTMENT"`, `null`, 1)},
		{name: "future category", row: strings.Replace(base, `"INVESTMENT"`, `"FUTURE"`, 1)},
		{name: "missing category", row: strings.Replace(base, `"account_category":"INVESTMENT",`, ``, 1)},
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
			wantCount := 1
			if test.accepted {
				wantCount++
			}
			if err != nil || len(connections) != 4 || len(connections[0].Accounts) != wantCount || len(connections[3].Accounts) != 0 || strings.Join(*paths, ",") != "/authorizations,/accounts" {
				t.Fatalf("connections=%+v paths=%v err=%v", connections, *paths, err)
			}
			for index, account := range connections[0].Accounts {
				wantID := globalAccountID(1)
				if index == 1 {
					wantID = globalAccountID(20)
				}
				if account.ID != wantID || !account.Selectable || !account.Eligible || account.UsabilityReason != portfolio.UsabilityReady {
					t.Fatalf("accepted account=%+v", account)
				}
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
