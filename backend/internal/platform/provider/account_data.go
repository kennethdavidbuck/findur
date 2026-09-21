package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	providergenerated "github.com/kennethdavidbuck/findur/backend/internal/generated/providerapi"
	"github.com/kennethdavidbuck/findur/backend/internal/portfolio"
)

const (
	activityWindowDays = 30
	activityRowLimit   = int32(500)
	maxDatasetRows     = 5000
	maxNumericWhole    = int64(131072)
	maxNumericScale    = int64(16383)
	maxPositionSymbol  = 120
	maxPositionKind    = 60
	maxActivityID      = 160
	maxActivityType    = 80
)

var (
	decimalPattern  = regexp.MustCompile(`^-?(?:0|[1-9][0-9]*)(?:\.[0-9]+)?(?:[eE][+-]?[0-9]+)?$`)
	currencyPattern = regexp.MustCompile(`^[A-Z]{3}$`)
)

type rawDecimal string

func (d *rawDecimal) UnmarshalJSON(value []byte) error {
	if bytes.Equal(value, []byte("null")) {
		return nil
	}
	if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
		var text string
		if err := json.Unmarshal(value, &text); err != nil {
			return err
		}
		value = []byte(text)
	}
	if !decimalPattern.Match(value) || len(value) > 100 || !postgresNumericSafe(value) {
		return errors.New("invalid decimal")
	}
	*d = rawDecimal(value)
	return nil
}

type rawBalance struct {
	Cash        *rawDecimal `json:"cash"`
	BuyingPower *rawDecimal `json:"buying_power"`
	Currency    *struct {
		Code *string `json:"code"`
	} `json:"currency"`
}

type rawPosition struct {
	Instrument struct {
		ID     uuid.UUID `json:"id"`
		Symbol string    `json:"symbol"`
		Kind   string    `json:"kind"`
	} `json:"instrument"`
	Currency  *string     `json:"currency"`
	Units     *rawDecimal `json:"units"`
	Price     *rawDecimal `json:"price"`
	CostBasis *rawDecimal `json:"cost_basis"`
}

type rawPositions struct {
	DataFreshness struct {
		AsOf time.Time `json:"as_of"`
	} `json:"data_freshness"`
	Results []rawPosition `json:"results"`
}

type rawActivity struct {
	ID        *string    `json:"id"`
	Type      *string    `json:"type"`
	TradeDate *time.Time `json:"trade_date"`
	Currency  *struct {
		Code *string `json:"code"`
	} `json:"currency"`
	Amount *rawDecimal `json:"amount"`
	Fee    *rawDecimal `json:"fee"`
	Price  *rawDecimal `json:"price"`
	Units  *rawDecimal `json:"units"`
}

type rawActivities struct {
	Data *[]rawActivity `json:"data"`
}

// LoadAccountData fetches exactly balances, all positions, and a bounded
// thirty-day activity window. It retains only the typed subset Findur needs.
func (c *InventoryClient) LoadAccountData(ctx context.Context, bearer, accountID string, now time.Time) (portfolio.AccountData, error) {
	if bearer == "" {
		return portfolio.AccountData{}, &portfolio.ProviderError{State: portfolio.StateUnauthorized}
	}
	id, err := uuid.Parse(accountID)
	if err != nil || id == uuid.Nil {
		return portfolio.AccountData{}, &portfolio.ProviderError{State: portfolio.StateMalformed}
	}
	balances, err := c.loadBalances(ctx, bearer, id)
	if err != nil {
		return portfolio.AccountData{}, err
	}
	positions, err := c.loadPositions(ctx, bearer, id)
	if err != nil {
		return portfolio.AccountData{}, err
	}
	activities, err := c.loadActivities(ctx, bearer, id, now)
	if err != nil {
		return portfolio.AccountData{}, err
	}
	return portfolio.AccountData{Balances: balances, Positions: positions, Activities: activities}, nil
}

func (c *InventoryClient) loadBalances(ctx context.Context, bearer string, accountID uuid.UUID) (portfolio.BalanceDataset, error) {
	request, err := providergenerated.NewAccountInformationGetUserAccountBalanceRequest(c.baseURL, accountID)
	if err != nil {
		return portfolio.BalanceDataset{}, err
	}
	request = request.WithContext(ctx)
	setBearer(request, bearer)
	var raw []rawBalance
	if err := c.doJSON(request, &raw); err != nil {
		return portfolio.BalanceDataset{}, err
	}
	if raw == nil || len(raw) > maxDatasetRows {
		return portfolio.BalanceDataset{}, &portfolio.ProviderError{State: portfolio.StateMalformed}
	}
	result := make([]portfolio.Balance, 0, len(raw))
	for _, row := range raw {
		currency, ok := normalizedCurrency(row.Currency)
		if !ok {
			return portfolio.BalanceDataset{}, &portfolio.ProviderError{State: portfolio.StateMalformed}
		}
		result = append(result, portfolio.Balance{Currency: currency, Cash: decimalPointer(row.Cash), BuyingPower: decimalPointer(row.BuyingPower)})
	}
	return portfolio.BalanceDataset{RetrievedAt: c.clock().UTC(), Rows: result}, nil
}

func (c *InventoryClient) loadPositions(ctx context.Context, bearer string, accountID uuid.UUID) (portfolio.PositionDataset, error) {
	request, err := providergenerated.NewAccountInformationGetAllAccountPositionsRequest(c.baseURL, accountID)
	if err != nil {
		return portfolio.PositionDataset{}, err
	}
	request = request.WithContext(ctx)
	setBearer(request, bearer)
	var raw rawPositions
	if err := c.doJSON(request, &raw); err != nil {
		return portfolio.PositionDataset{}, err
	}
	if raw.Results == nil || len(raw.Results) > maxDatasetRows || raw.DataFreshness.AsOf.IsZero() {
		return portfolio.PositionDataset{}, &portfolio.ProviderError{State: portfolio.StateMalformed}
	}
	result := make([]portfolio.Position, 0, len(raw.Results))
	for _, row := range raw.Results {
		currency, ok := normalizedCode(row.Currency)
		symbol, symbolOK := normalizedRequiredText(row.Instrument.Symbol, maxPositionSymbol)
		kind, kindOK := normalizedRequiredText(row.Instrument.Kind, maxPositionKind)
		if !ok || row.Instrument.ID == uuid.Nil || !symbolOK || !kindOK {
			return portfolio.PositionDataset{}, &portfolio.ProviderError{State: portfolio.StateMalformed}
		}
		result = append(result, portfolio.Position{InstrumentID: row.Instrument.ID.String(), Symbol: symbol, Kind: kind, Currency: currency, Units: decimalPointer(row.Units), Price: decimalPointer(row.Price), CostBasis: decimalPointer(row.CostBasis)})
	}
	return portfolio.PositionDataset{ObservedAt: raw.DataFreshness.AsOf.UTC(), RetrievedAt: c.clock().UTC(), Rows: result}, nil
}

func (c *InventoryClient) loadActivities(ctx context.Context, bearer string, accountID uuid.UUID, now time.Time) (portfolio.ActivityDataset, error) {
	start, end := now.UTC().AddDate(0, 0, -activityWindowDays), now.UTC()
	offset := int32(0)
	params := &providergenerated.AccountInformationGetAccountActivitiesParams{
		StartDate: &providergenerated.ReportingDate{Time: start},
		EndDate:   &providergenerated.ReportingDate{Time: end},
		Offset:    &offset,
		Limit:     pointer(activityRowLimit),
	}
	request, err := providergenerated.NewAccountInformationGetAccountActivitiesRequest(c.baseURL, accountID, params)
	if err != nil {
		return portfolio.ActivityDataset{}, err
	}
	request = request.WithContext(ctx)
	setBearer(request, bearer)
	var raw rawActivities
	if err := c.doJSON(request, &raw); err != nil {
		return portfolio.ActivityDataset{}, err
	}
	if raw.Data == nil || len(*raw.Data) > int(activityRowLimit) {
		return portfolio.ActivityDataset{}, &portfolio.ProviderError{State: portfolio.StateMalformed}
	}
	result := make([]portfolio.Activity, 0, len(*raw.Data))
	for _, row := range *raw.Data {
		currency, ok := normalizedCurrency(row.Currency)
		if !ok || row.ID == nil || row.Type == nil {
			return portfolio.ActivityDataset{}, &portfolio.ProviderError{State: portfolio.StateMalformed}
		}
		activityID, idOK := normalizedRequiredText(*row.ID, maxActivityID)
		activityType, typeOK := normalizedRequiredText(*row.Type, maxActivityType)
		if !idOK || !typeOK {
			return portfolio.ActivityDataset{}, &portfolio.ProviderError{State: portfolio.StateMalformed}
		}
		result = append(result, portfolio.Activity{ID: activityID, Type: activityType, Currency: currency, TradeDate: row.TradeDate, Amount: decimalPointer(row.Amount), Fee: decimalPointer(row.Fee), Price: decimalPointer(row.Price), Units: decimalPointer(row.Units)})
	}
	return portfolio.ActivityDataset{RetrievedAt: c.clock().UTC(), Rows: result}, nil
}

func normalizedCurrency(value *struct {
	Code *string `json:"code"`
}) (string, bool) {
	if value == nil {
		return "", false
	}
	return normalizedCode(value.Code)
}

func normalizedCode(value *string) (string, bool) {
	if value == nil {
		return "", false
	}
	code := strings.ToUpper(strings.TrimSpace(*value))
	return code, currencyPattern.MatchString(code)
}

func normalizedRequiredText(value string, maxLength int) (string, bool) {
	value = strings.Join(strings.Fields(value), " ")
	return value, value != "" && utf8.RuneCountInString(value) <= maxLength
}

func postgresNumericSafe(value []byte) bool {
	text := strings.TrimPrefix(string(value), "-")
	exponent := int64(0)
	if index := strings.IndexAny(text, "eE"); index >= 0 {
		parsed, err := strconv.ParseInt(text[index+1:], 10, 64)
		if err != nil {
			return false
		}
		exponent = parsed
		text = text[:index]
	}
	wholeDigits, fractionalDigits := int64(len(text)), int64(0)
	if index := strings.IndexByte(text, '.'); index >= 0 {
		wholeDigits = int64(index)
		fractionalDigits = int64(len(text) - index - 1)
	}
	return wholeDigits+exponent <= maxNumericWhole && fractionalDigits-exponent <= maxNumericScale
}

func decimalPointer(value *rawDecimal) *string {
	if value == nil {
		return nil
	}
	result := string(*value)
	return &result
}

func pointer[T any](value T) *T { return &value }
