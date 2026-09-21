package provider

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/kennethdavidbuck/findur/backend/internal/portfolio"
)

func TestAccountDataUsesOnlyAllowlistedReadsAndBoundsActivities(t *testing.T) {
	now := time.Date(2026, 9, 20, 12, 0, 0, 0, time.UTC)
	baseURL, _ := url.Parse("https://api.snaptrade.example")
	var requests []*http.Request
	client, err := NewInventoryClient(baseURL, roundTripFunc(func(request *http.Request) (*http.Response, error) {
		requests = append(requests, request)
		if request.Method != http.MethodGet || request.Header.Get("Authorization") != "Bearer access-token" {
			t.Fatalf("unsafe request: %s %s", request.Method, request.URL)
		}
		switch {
		case strings.HasSuffix(request.URL.Path, "/balances"):
			return jsonResponse(http.StatusOK, `[{"cash":100.123456789,"buying_power":"150.50","currency":{"code":"usd"}}]`), nil
		case strings.HasSuffix(request.URL.Path, "/positions/all"):
			return jsonResponse(http.StatusOK, `{"data_freshness":{"as_of":"2026-09-20T11:55:00Z"},"results":[{"instrument":{"id":"87b24961-b51e-4db8-9226-f198f6518a89","symbol":"AAPL","kind":"stock"},"currency":"USD","units":"1.25","price":"123.456789","cost_basis":"100.01"}]}`), nil
		case strings.HasSuffix(request.URL.Path, "/activities"):
			query := request.URL.Query()
			if query.Get("limit") != "500" || query.Get("offset") != "0" || query.Get("startDate") != "2026-08-21" || query.Get("endDate") != "2026-09-20" {
				t.Fatalf("unbounded activity query: %s", request.URL.RawQuery)
			}
			return jsonResponse(http.StatusOK, `{"data":[{"id":"activity-1","type":"BUY","trade_date":"2026-09-19T10:00:00Z","currency":{"code":"CAD"},"amount":-10.25,"units":"1"}]}`), nil
		default:
			t.Fatalf("forbidden provider endpoint: %s", request.URL.Path)
			return nil, nil
		}
	}), func() time.Time { return now })
	if err != nil {
		t.Fatal(err)
	}

	result, err := client.LoadAccountData(context.Background(), "access-token", "917c8734-8470-4a3e-a18f-57c3f2ee6631", now)
	if err != nil {
		t.Fatal(err)
	}
	if len(requests) != 3 || len(result.Balances.Rows) != 1 || len(result.Positions.Rows) != 1 || len(result.Activities.Rows) != 1 {
		t.Fatalf("requests=%d result=%+v", len(requests), result)
	}
	if result.Balances.Rows[0].Cash == nil || *result.Balances.Rows[0].Cash != "100.123456789" || result.Balances.Rows[0].Currency != "USD" {
		t.Fatalf("precision/currency lost: %+v", result.Balances.Rows[0])
	}
	if !result.Positions.ObservedAt.Equal(time.Date(2026, 9, 20, 11, 55, 0, 0, time.UTC)) || result.Balances.ObservedAt != nil || result.Activities.ObservedAt != nil {
		t.Fatalf("dataset observation metadata=%+v", result)
	}
	if result.Balances.RetrievedAt.IsZero() || result.Positions.RetrievedAt.IsZero() || result.Activities.RetrievedAt.IsZero() {
		t.Fatalf("dataset retrieval metadata=%+v", result)
	}
}

func TestAccountDataRejectsMissingProviderObservationProof(t *testing.T) {
	baseURL, _ := url.Parse("https://api.snaptrade.example")
	client, _ := NewInventoryClient(baseURL, roundTripFunc(func(request *http.Request) (*http.Response, error) {
		switch {
		case strings.HasSuffix(request.URL.Path, "/balances"):
			return jsonResponse(http.StatusOK, `[]`), nil
		case strings.HasSuffix(request.URL.Path, "/positions/all"):
			return jsonResponse(http.StatusOK, `{"results":[]}`), nil
		default:
			t.Fatalf("provider continued after missing observation proof: %s", request.URL.Path)
			return nil, nil
		}
	}), time.Now)

	if _, err := client.LoadAccountData(context.Background(), "access-token", "917c8734-8470-4a3e-a18f-57c3f2ee6631", time.Now()); err == nil {
		t.Fatal("missing provider observation proof was accepted")
	}
}

func TestAccountDataRejectsPartialMalformedDataset(t *testing.T) {
	baseURL, _ := url.Parse("https://api.snaptrade.example")
	client, _ := NewInventoryClient(baseURL, roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if strings.HasSuffix(request.URL.Path, "/balances") {
			return jsonResponse(http.StatusOK, `[{"cash":100,"currency":null}]`), nil
		}
		t.Fatalf("provider continued after malformed required dataset: %s", request.URL.Path)
		return nil, nil
	}), time.Now)

	if _, err := client.LoadAccountData(context.Background(), "access-token", "917c8734-8470-4a3e-a18f-57c3f2ee6631", time.Now()); err == nil {
		t.Fatal("malformed dataset was accepted")
	}
}

func TestAccountDataRejectsPostgresUnsafeProviderValuesAsMalformed(t *testing.T) {
	baseURL, _ := url.Parse("https://api.snaptrade.example")
	validBalances := `[{"cash":"1.25","buying_power":"2.5","currency":{"code":"USD"}}]`
	validPositions := `{"data_freshness":{"as_of":"2026-09-20T11:55:00Z"},"results":[{"instrument":{"id":"87b24961-b51e-4db8-9226-f198f6518a89","symbol":"AAPL","kind":"stock"},"currency":"USD","units":"1.25","price":"123.45","cost_basis":"100.01"}]}`
	validActivities := `{"data":[{"id":"activity-1","type":"BUY","currency":{"code":"USD"},"amount":"10.25"}]}`
	for _, test := range []struct {
		name, balances, positions, activities string
	}{
		{name: "positive decimal exponent", balances: `[{"cash":"1e131072","currency":{"code":"USD"}}]`, positions: validPositions, activities: validActivities},
		{name: "negative decimal exponent", balances: `[{"cash":"1e-16384","currency":{"code":"USD"}}]`, positions: validPositions, activities: validActivities},
		{name: "position kind length", balances: validBalances, positions: strings.Replace(validPositions, `"kind":"stock"`, `"kind":"`+strings.Repeat("k", 61)+`"`, 1), activities: validActivities},
		{name: "activity type length", balances: validBalances, positions: validPositions, activities: strings.Replace(validActivities, `"type":"BUY"`, `"type":"`+strings.Repeat("T", 81)+`"`, 1)},
	} {
		t.Run(test.name, func(t *testing.T) {
			client, _ := NewInventoryClient(baseURL, roundTripFunc(func(request *http.Request) (*http.Response, error) {
				switch {
				case strings.HasSuffix(request.URL.Path, "/balances"):
					return jsonResponse(http.StatusOK, test.balances), nil
				case strings.HasSuffix(request.URL.Path, "/positions/all"):
					return jsonResponse(http.StatusOK, test.positions), nil
				case strings.HasSuffix(request.URL.Path, "/activities"):
					return jsonResponse(http.StatusOK, test.activities), nil
				default:
					t.Fatalf("unexpected provider path: %s", request.URL.Path)
					return nil, nil
				}
			}), time.Now)
			_, err := client.LoadAccountData(context.Background(), "access-token", "917c8734-8470-4a3e-a18f-57c3f2ee6631", time.Now())
			var providerErr *portfolio.ProviderError
			if !errors.As(err, &providerErr) || providerErr.State != portfolio.StateMalformed {
				t.Fatalf("error=%v", err)
			}
		})
	}
}
