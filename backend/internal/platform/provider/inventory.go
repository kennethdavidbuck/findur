package provider

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	providergenerated "github.com/kennethdavidbuck/findur/backend/internal/generated/providerapi"
	"github.com/kennethdavidbuck/findur/backend/internal/portfolio"
)

const (
	providerResponseLimit  = 1 << 20
	conservativeRetryDelay = time.Minute
	maxConnections         = 100
	maxAccounts            = 500
	retryAfterHeader       = "Retry-After"
	accountRemainingHeader = "X-RateLimit-Account-Remaining"
	accountResetHeader     = "X-RateLimit-Account-Reset"
	authorizationHeader    = "Authorization"
	acceptHeader           = "Accept"
	bearerPrefix           = "Bearer "
	jsonMediaType          = "application/json"
	unknownValue           = "unknown"
	defaultAccountLabel    = "Account"
	defaultBrokerageLabel  = "Connected institution"
)

// InventoryClient is the purpose-limited OAuth bearer adapter generated from
// the pinned SnapTrade inventory overlay.
type InventoryClient struct {
	baseURL string
	http    doer
	clock   func() time.Time
}

type normalizedConnection struct {
	id         uuid.UUID
	connection portfolio.Connection
}

// NewInventoryClient validates and constructs the purpose-limited bearer client.
func NewInventoryClient(baseURL *url.URL, httpClient doer, clock func() time.Time) (*InventoryClient, error) {
	if baseURL == nil || httpClient == nil || clock == nil || baseURL.Scheme == "" || baseURL.Host == "" {
		return nil, errors.New("incomplete inventory provider configuration")
	}
	return &InventoryClient{baseURL: strings.TrimRight(baseURL.String(), "/"), http: httpClient, clock: clock}, nil
}

// Load lists connections first, then only the accounts belonging to each usable connection.
func (c *InventoryClient) Load(ctx context.Context, bearer string) ([]portfolio.Connection, error) {
	if bearer == "" {
		return nil, &portfolio.ProviderError{State: portfolio.StateUnauthorized}
	}
	rawConnections, err := c.fetchConnections(ctx, bearer)
	if err != nil {
		return nil, err
	}
	return c.loadConnectionAccounts(ctx, bearer, rawConnections)
}

func (c *InventoryClient) fetchConnections(ctx context.Context, bearer string) ([]providergenerated.BrokerageAuthorization, error) {
	request, err := providergenerated.NewConnectionsListBrokerageAuthorizationsRequest(c.baseURL)
	if err != nil {
		return nil, err
	}
	request = request.WithContext(ctx)
	setBearer(request, bearer)
	var rawConnections []providergenerated.BrokerageAuthorization
	if err := c.doJSON(request, &rawConnections); err != nil {
		return nil, err
	}
	if rawConnections == nil || len(rawConnections) > maxConnections {
		return nil, &portfolio.ProviderError{State: portfolio.StateMalformed}
	}
	return rawConnections, nil
}

func (c *InventoryClient) loadConnectionAccounts(ctx context.Context, bearer string, rawConnections []providergenerated.BrokerageAuthorization) ([]portfolio.Connection, error) {
	normalized, partialErr := normalizeConnections(rawConnections)
	connections := make([]portfolio.Connection, 0, len(normalized))
	accountIDs := make(map[string]struct{})
	fetchFailed := false
	for _, candidate := range normalized {
		connection := candidate.connection
		if connection.Status == portfolio.ConnectionStatusDisabled {
			connections = append(connections, connection)
			continue
		}
		if fetchFailed {
			connections = append(connections, unavailableConnection(connection))
			continue
		}
		rawAccounts, err := c.fetchAccounts(ctx, bearer, candidate.id)
		if err != nil {
			connections = append(connections, unavailableConnection(connection))
			partialErr = err
			fetchFailed = true
			continue
		}
		connection, err = normalizeAccounts(connection, rawAccounts, accountIDs)
		if err != nil {
			partialErr = err
		}
		connections = append(connections, connection)
	}
	return connections, partialErr
}

func normalizeConnections(rawConnections []providergenerated.BrokerageAuthorization) ([]normalizedConnection, error) {
	connections := make([]normalizedConnection, 0, len(rawConnections))
	connectionIDs := make(map[string]struct{}, len(rawConnections))
	var partialErr error
	for _, raw := range rawConnections {
		connection, usable := normalizeConnection(raw)
		if !usable || !rememberUnique(connectionIDs, connection.ID) {
			partialErr = &portfolio.ProviderError{State: portfolio.StateMalformed}
			continue
		}
		connections = append(connections, normalizedConnection{id: *raw.Id, connection: connection})
	}
	return connections, partialErr
}

func (c *InventoryClient) fetchAccounts(ctx context.Context, bearer string, connectionID uuid.UUID) ([]providergenerated.Account, error) {
	request, err := providergenerated.NewConnectionsListBrokerageAuthorizationAccountsRequest(c.baseURL, connectionID)
	if err != nil {
		return nil, err
	}
	request = request.WithContext(ctx)
	setBearer(request, bearer)
	var accounts []providergenerated.Account
	if err := c.doJSON(request, &accounts); err != nil {
		return nil, err
	}
	if accounts == nil || len(accounts) > maxAccounts {
		return nil, &portfolio.ProviderError{State: portfolio.StateMalformed}
	}
	return accounts, nil
}

func normalizeAccounts(connection portfolio.Connection, rawAccounts []providergenerated.Account, accountIDs map[string]struct{}) (portfolio.Connection, error) {
	var partialErr error
	for _, raw := range rawAccounts {
		account, ok := normalizeAccount(raw, connection.ID)
		if !ok || !rememberUnique(accountIDs, account.ID) {
			connection = unavailableConnection(connection)
			partialErr = &portfolio.ProviderError{State: portfolio.StateMalformed}
			continue
		}
		connection.Accounts = append(connection.Accounts, account)
	}
	if partialErr != nil {
		connection = unavailableConnection(connection)
	}
	return connection, partialErr
}

func rememberUnique(seen map[string]struct{}, id string) bool {
	if _, duplicate := seen[id]; duplicate {
		return false
	}
	seen[id] = struct{}{}
	return true
}

func unavailableConnection(connection portfolio.Connection) portfolio.Connection {
	connection.Status, connection.Available, connection.Eligible = portfolio.ConnectionStatusUnavailable, false, false
	for index := range connection.Accounts {
		connection.Accounts[index].Available = false
		connection.Accounts[index].Eligible = false
		connection.Accounts[index].Selectable = false
		connection.Accounts[index].UsabilityReason = portfolio.UsabilityConnectionUnavailable
	}
	return connection
}

func (c *InventoryClient) doJSON(request *http.Request, target any) error {
	response, err := c.http.Do(request)
	if err != nil {
		return &portfolio.ProviderError{State: portfolio.StateUnavailable}
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != http.StatusOK {
		_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, providerResponseLimit))
		return c.responseError(response)
	}
	limited := &io.LimitedReader{R: response.Body, N: providerResponseLimit + 1}
	decoder := json.NewDecoder(limited)
	if err := decoder.Decode(target); err != nil {
		return &portfolio.ProviderError{State: portfolio.StateMalformed}
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return &portfolio.ProviderError{State: portfolio.StateMalformed}
	}
	if limited.N <= 0 {
		return &portfolio.ProviderError{State: portfolio.StateMalformed}
	}
	return nil
}

func (c *InventoryClient) responseError(response *http.Response) error {
	switch response.StatusCode {
	case http.StatusUnauthorized:
		return &portfolio.ProviderError{State: portfolio.StateUnauthorized}
	case http.StatusForbidden, http.StatusNotFound:
		return &portfolio.ProviderError{State: portfolio.StateUnavailable}
	case http.StatusTooManyRequests:
		now := c.clock().UTC()
		retryAt := parseRetryAfter(response.Header.Get(retryAfterHeader), now)
		if retryAt == nil && response.Header.Get(accountRemainingHeader) == "0" {
			retryAt = parseRetryAfter(response.Header.Get(accountResetHeader), now)
		}
		if retryAt == nil {
			fallback := now.Add(conservativeRetryDelay)
			retryAt = &fallback
		}
		return &portfolio.ProviderError{State: portfolio.StateRateLimited, RetryAt: retryAt}
	default:
		return &portfolio.ProviderError{State: portfolio.StateUnavailable}
	}
}

func setBearer(request *http.Request, bearer string) {
	request.Header.Set(acceptHeader, jsonMediaType)
	request.Header.Set(authorizationHeader, bearerPrefix+bearer)
}

func normalizeConnection(raw providergenerated.BrokerageAuthorization) (portfolio.Connection, bool) {
	if raw.Id == nil || *raw.Id == uuid.Nil || raw.Disabled == nil {
		return portfolio.Connection{}, false
	}
	label := defaultBrokerageLabel
	if raw.Brokerage != nil {
		if raw.Brokerage.DisplayName != nil {
			label = portfolio.SafeLabel(*raw.Brokerage.DisplayName, label)
		} else if raw.Brokerage.Name != nil {
			label = portfolio.SafeLabel(*raw.Brokerage.Name, label)
		}
	}
	result := portfolio.Connection{
		ID:             raw.Id.String(),
		BrokerageLabel: label,
		Status:         portfolio.ConnectionStatusActive,
		SyncMode:       portfolio.SyncModeUnknown,
		Available:      true,
		Eligible:       true,
	}
	if *raw.Disabled {
		result.Status, result.Available, result.Eligible = portfolio.ConnectionStatusDisabled, false, false
	}
	if raw.DataFreshnessMode != nil {
		institution := raw.DataFreshnessMode.Institution
		snaptrade := raw.DataFreshnessMode.Snaptrade
		switch {
		case institution == providergenerated.BrokerageAuthorizationDataFreshnessModeInstitutionDelayed || snaptrade == providergenerated.BrokerageAuthorizationDataFreshnessModeSnaptradeDelayed:
			result.SyncMode = portfolio.SyncModeDelayed
		case institution == providergenerated.BrokerageAuthorizationDataFreshnessModeInstitutionRealtime && snaptrade == providergenerated.BrokerageAuthorizationDataFreshnessModeSnaptradeRealtime:
			result.SyncMode = portfolio.SyncModeRealtime
		}
	}
	return result, true
}

func normalizeAccount(raw providergenerated.Account, connectionID string) (portfolio.Account, bool) {
	if raw.Id == uuid.Nil || raw.BrokerageAuthorization == uuid.Nil || raw.BrokerageAuthorization.String() != connectionID {
		return portfolio.Account{}, false
	}

	result := initialAccount(raw)
	statusOpen, statusKnown := false, false
	if raw.Status != nil {
		switch *raw.Status {
		case providergenerated.Open:
			statusOpen, statusKnown, result.Available = true, true, true
		case providergenerated.Closed:
			statusKnown, result.Available = true, false
		case providergenerated.Unavailable, providergenerated.Archived:
			statusKnown, result.Available = true, false
		default:
			return portfolio.Account{}, false
		}
	}
	if raw.SyncStatus.Holdings != nil {
		holdings := raw.SyncStatus.Holdings
		switch {
		case holdings.HoldingsUnavailable != nil && *holdings.HoldingsUnavailable:
			result.SyncState, result.Available = portfolio.AccountSyncStateUnavailable, false
		case holdings.InitialSyncCompleted != nil && *holdings.InitialSyncCompleted:
			result.SyncState = portfolio.AccountSyncStateComplete
		case holdings.InitialSyncCompleted != nil:
			result.SyncState = portfolio.AccountSyncStatePending
		}
	}

	result.Eligible = result.Available && statusOpen &&
		result.Category == portfolio.AccountCategoryInvestment && result.SyncState == portfolio.AccountSyncStateComplete
	result.Selectable = result.Available && (statusOpen || !statusKnown) &&
		(result.Category == portfolio.AccountCategoryInvestment || result.Category == portfolio.AccountCategoryUnknown) &&
		result.SyncState == portfolio.AccountSyncStateComplete
	closed := raw.Status != nil && *raw.Status == providergenerated.Closed
	result.UsabilityReason = accountUsabilityReason(result, statusKnown, closed)
	return result, true
}

func initialAccount(raw providergenerated.Account) portfolio.Account {
	result := portfolio.Account{
		ID:          raw.Id.String(),
		Category:    portfolio.AccountCategoryUnknown,
		Type:        unknownValue,
		MaskedLabel: defaultAccountLabel,
		SyncState:   portfolio.AccountSyncStateUnknown,
		Available:   true,
	}
	if raw.Name != nil {
		result.MaskedLabel = safeAccountLabel(*raw.Name, raw.Number)
	}
	if raw.Number != "" {
		if suffix := maskedSuffix(raw.Number); suffix != "" {
			result.MaskedLabel += " (•••• " + suffix + ")"
		}
	}
	if raw.AccountCategory != nil {
		switch *raw.AccountCategory {
		case providergenerated.INVESTMENT:
			result.Category = portfolio.AccountCategoryInvestment
		case providergenerated.DEPOSIT:
			result.Category = portfolio.AccountCategoryDeposit
		case providergenerated.LOC:
			result.Category = portfolio.AccountCategoryCredit
		}
	}
	if raw.RawType != nil {
		result.Type = portfolio.SafeLabel(*raw.RawType, unknownValue)
	}
	return result
}

func accountUsabilityReason(account portfolio.Account, statusKnown, closed bool) portfolio.UsabilityReason {
	// Account closure and unavailability take precedence over sync and category
	// reasons, including when unavailable holdings made the account unavailable.
	switch {
	case !account.Available && statusKnown && closed:
		return portfolio.UsabilityAccountClosed
	case !account.Available:
		return portfolio.UsabilityAccountUnavailable
	case account.SyncState == portfolio.AccountSyncStateUnavailable:
		return portfolio.UsabilitySyncUnavailable
	case account.Category == portfolio.AccountCategoryDeposit || account.Category == portfolio.AccountCategoryCredit:
		return portfolio.UsabilityUnsupportedCategory
	case !statusKnown:
		return portfolio.UsabilityProvisionalStatus
	case account.Category == portfolio.AccountCategoryUnknown:
		return portfolio.UsabilityProvisionalCategory
	case account.SyncState != portfolio.AccountSyncStateComplete:
		return portfolio.UsabilitySyncPending
	default:
		return portfolio.UsabilityReady
	}
}

func maskedSuffix(value string) string {
	var safe []rune
	for _, character := range value {
		if unicode.IsLetter(character) || unicode.IsDigit(character) {
			safe = append(safe, character)
		}
	}
	if len(safe) <= 4 {
		return ""
	}
	return string(safe[len(safe)-4:])
}

func safeAccountLabel(label, accountNumber string) string {
	result := portfolio.SafeLabel(label, defaultAccountLabel)
	if accountNumber == "" {
		return result
	}
	if sensitive := alphanumeric(accountNumber); sensitive != "" {
		result = regexp.MustCompile(accountNumberPattern(sensitive)).ReplaceAllString(result, "••••")
	}
	return portfolio.SafeLabel(result, defaultAccountLabel)
}

func accountNumberPattern(value string) string {
	characters := []rune(value)
	parts := make([]string, 0, len(characters))
	for _, character := range characters {
		parts = append(parts, regexp.QuoteMeta(string(character)))
	}
	return `(?i)` + strings.Join(parts, `[^\pL\pN]*`)
}

func alphanumeric(value string) string {
	var safe []rune
	for _, character := range value {
		if unicode.IsLetter(character) || unicode.IsDigit(character) {
			safe = append(safe, character)
		}
	}
	return string(safe)
}

func parseRetryAfter(value string, now time.Time) *time.Time {
	value = strings.TrimSpace(value)
	if decimalDigits(value) {
		const maxDelay = time.Duration(1<<63 - 1)
		seconds, err := strconv.ParseUint(value, 10, 64)
		delay := maxDelay
		if err == nil && seconds <= uint64(maxDelay/time.Second) {
			delay = time.Duration(seconds) * time.Second
		}
		result := now.Add(delay)
		return &result
	}
	if parsed, err := http.ParseTime(value); err == nil && parsed.After(now) {
		parsed = parsed.UTC()
		return &parsed
	}
	return nil
}

func decimalDigits(value string) bool {
	if value == "" {
		return false
	}
	for _, character := range value {
		if character < '0' || character > '9' {
			return false
		}
	}
	return true
}
