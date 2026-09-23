// Package auth owns pre-login authorization attempts and their security policy.
package auth

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/url"
	"slices"
	"time"

	coreoidc "github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

// Authorization protocol, route, and lifetime constants shared by its adapters.
const (
	AttemptLifetime          = 10 * time.Minute
	SnapTradeProvider        = "snaptrade"
	SnapTradeCallbackPath    = "/api/auth/snaptrade/callback"
	AuthorizationResultRoute = "/onboarding/accounts"
	AuthorizationDeniedRoute = "/connect?authorization=denied"
	AuthorizationRetryRoute  = "/connect?authorization=restart_required"
	DefaultReturnRoute       = "/connect"
	PortfolioReturnRoute     = "/portfolio"
	ScopeOpenID              = "openid"
	ScopeRead                = "read"
	randomBytes              = 32
)

// Stable authorization errors exposed to adapters without sensitive detail.
var (
	ErrUnavailable    = errors.New("authorization initiation unavailable")
	ErrInitialization = errors.New("authorization initialization failed")
	ErrNotClaimable   = errors.New("authorization attempt is expired or already claimed")
)

// InitializationStage is a bounded, non-sensitive failure location suitable for logs.
type InitializationStage string

// Initialization failure stages safe for categorical logging.
const (
	StageDiscovery  InitializationStage = "oidc_discovery"
	StageGeneration InitializationStage = "attempt_generation"
	StageStorage    InitializationStage = "attempt_persistence"
)

type initializationError struct {
	stage InitializationStage
	cause error
}

func (e *initializationError) Error() string        { return ErrInitialization.Error() }
func (e *initializationError) Unwrap() error        { return e.cause }
func (e *initializationError) Is(target error) bool { return target == ErrInitialization }

// InitializationStageOf returns only the safe stage category, never the underlying error.
func InitializationStageOf(err error) InitializationStage {
	var failure *initializationError
	if errors.As(err, &failure) {
		return failure.stage
	}
	return ""
}

func initializationFailure(stage InitializationStage, cause error) error {
	return &initializationError{stage: stage, cause: cause}
}

// Discovery is the validated subset of OIDC discovery needed to initiate authorization.
type Discovery struct {
	AuthorizationEndpoint string
}

// DiscoveryProvider returns validated, cacheable provider metadata.
type DiscoveryProvider interface {
	Discover(context.Context) (Discovery, error)
}

// Attempt is the minimum persisted correlation state. Secret values are hashed or encrypted.
type Attempt struct {
	StateHash          []byte
	NonceHash          []byte
	BrowserBindingHash []byte
	EncryptedVerifier  []byte
	ReturnRoute        string
	ExpiresAt          time.Time
}

// AttemptRepository provides short transactions around attempt state.
type AttemptRepository interface {
	Create(context.Context, Attempt) error
	Delete(context.Context, []byte) error
	Claim(context.Context, []byte, []byte, time.Time) error
	Cleanup(context.Context, time.Time) (int64, error)
}

// BeginResult contains only browser-safe, one-time values.
type BeginResult struct {
	AuthorizationURL string
	BrowserBinding   string
	ExpiresAt        time.Time
}

type attemptSecrets struct {
	state, nonce, verifier, binding string
}

// Config is validated authorization initiation policy.
type Config struct {
	Enabled          bool
	ClientID         string
	CallbackURL      string
	AllowedReturns   []string
	DefaultReturn    string
	HashKey          []byte
	EncryptionKey    []byte
	Random           io.Reader
	Clock            func() time.Time
	OperationTimeout time.Duration
}

// Service creates and atomically claims short-lived OAuth attempts.
type Service struct {
	config     Config
	repository AttemptRepository
	discovery  DiscoveryProvider
	aead       cipher.AEAD
}

// NewService validates dependencies and builds an authorization service.
func NewService(config Config, repository AttemptRepository, discovery DiscoveryProvider) (*Service, error) {
	if config.Random == nil || config.Clock == nil || config.OperationTimeout <= 0 || len(config.HashKey) < 32 || len(config.EncryptionKey) != 32 {
		return nil, errors.New("invalid authorization service policy")
	}
	if subtle.ConstantTimeCompare(config.HashKey, config.EncryptionKey) == 1 {
		return nil, errors.New("authorization hashing and encryption keys must be independent")
	}
	if repository == nil || discovery == nil || config.ClientID == "" || config.CallbackURL == "" || config.DefaultReturn == "" {
		return nil, errors.New("incomplete authorization service configuration")
	}
	if !slices.Contains(config.AllowedReturns, config.DefaultReturn) {
		return nil, errors.New("default return route is not allowlisted")
	}
	block, err := aes.NewCipher(config.EncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("create verifier cipher: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create verifier envelope: %w", err)
	}
	return &Service{config: config, repository: repository, discovery: discovery, aead: aead}, nil
}

// Enabled reports the independent product gate; it is never inferred from credentials.
func (s *Service) Enabled() bool { return s != nil && s.config.Enabled }

// Begin creates one attempt only after explicit invocation.
func (s *Service) Begin(ctx context.Context, requestedReturn string) (BeginResult, error) {
	if !s.Enabled() {
		return BeginResult{}, ErrUnavailable
	}
	ctx, cancel := context.WithTimeout(ctx, s.config.OperationTimeout)
	defer cancel()
	metadata, err := s.discovery.Discover(ctx)
	if err != nil {
		return BeginResult{}, initializationFailure(StageDiscovery, err)
	}
	secrets, err := s.attemptSecrets()
	if err != nil {
		return BeginResult{}, initializationFailure(StageGeneration, err)
	}
	authorizationURL, err := s.authorizationURL(metadata.AuthorizationEndpoint, secrets)
	if err != nil {
		return BeginResult{}, initializationFailure(StageDiscovery, err)
	}
	return s.persistAttempt(ctx, requestedReturn, authorizationURL, secrets)
}

func (s *Service) attemptSecrets() (attemptSecrets, error) {
	state, err := s.secret()
	if err != nil {
		return attemptSecrets{}, err
	}
	nonce, err := s.secret()
	if err != nil {
		return attemptSecrets{}, err
	}
	verifier, err := s.secret()
	if err != nil {
		return attemptSecrets{}, err
	}
	binding, err := s.secret()
	if err != nil {
		return attemptSecrets{}, err
	}
	return attemptSecrets{state: state, nonce: nonce, verifier: verifier, binding: binding}, nil
}

func (s *Service) authorizationURL(endpoint string, secrets attemptSecrets) (string, error) {
	oauthConfig := oauth2.Config{
		ClientID: s.config.ClientID, RedirectURL: s.config.CallbackURL,
		Endpoint: oauth2.Endpoint{AuthURL: endpoint}, Scopes: []string{ScopeOpenID, ScopeRead},
	}
	authorizationURL := oauthConfig.AuthCodeURL(secrets.state, oauth2.S256ChallengeOption(secrets.verifier), coreoidc.Nonce(secrets.nonce))
	parsedAuthorizationURL, err := url.Parse(authorizationURL)
	if err != nil || parsedAuthorizationURL.Host == "" {
		return "", errors.New("invalid authorization endpoint")
	}
	return authorizationURL, nil
}

func (s *Service) persistAttempt(ctx context.Context, requestedReturn, authorizationURL string, secrets attemptSecrets) (BeginResult, error) {
	stateHash := s.hash(secrets.state)
	now := s.config.Clock().UTC()
	expires := now.Add(AttemptLifetime)
	encryptedVerifier, err := s.encryptVerifier(stateHash, secrets.verifier)
	if err != nil {
		return BeginResult{}, initializationFailure(StageGeneration, err)
	}
	attempt := Attempt{
		StateHash: stateHash, NonceHash: s.hash(secrets.nonce), BrowserBindingHash: s.hash(secrets.binding),
		EncryptedVerifier: encryptedVerifier, ReturnRoute: s.safeReturn(requestedReturn), ExpiresAt: expires,
	}
	if _, err := s.repository.Cleanup(ctx, now); err != nil {
		return BeginResult{}, initializationFailure(StageStorage, err)
	}
	if err := s.repository.Create(ctx, attempt); err != nil {
		return BeginResult{}, initializationFailure(StageStorage, err)
	}

	return BeginResult{AuthorizationURL: authorizationURL, BrowserBinding: secrets.binding, ExpiresAt: expires}, nil
}

// Claim atomically makes a matching, unexpired attempt single-use.
func (s *Service) Claim(ctx context.Context, state, binding string) error {
	if state == "" || binding == "" {
		return ErrNotClaimable
	}
	ctx, cancel := context.WithTimeout(ctx, s.config.OperationTimeout)
	defer cancel()
	if err := s.repository.Claim(ctx, s.hash(state), s.hash(binding), s.config.Clock().UTC()); err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return err
		}
		return ErrNotClaimable
	}
	return nil
}

// Cleanup removes expired or terminal authorization attempts.
func (s *Service) Cleanup(ctx context.Context) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, s.config.OperationTimeout)
	defer cancel()
	return s.repository.Cleanup(ctx, s.config.Clock().UTC())
}

func (s *Service) secret() (string, error) {
	value := make([]byte, randomBytes)
	if _, err := io.ReadFull(s.config.Random, value); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(value), nil
}

func (s *Service) hash(value string) []byte {
	digest := hmac.New(sha256.New, s.config.HashKey)
	_, _ = digest.Write([]byte(value))
	return digest.Sum(nil)
}

func (s *Service) encryptVerifier(stateHash []byte, verifier string) ([]byte, error) {
	nonce := make([]byte, s.aead.NonceSize())
	if _, err := io.ReadFull(s.config.Random, nonce); err != nil {
		return nil, err
	}
	return s.aead.Seal(nonce, nonce, []byte(verifier), stateHash), nil
}

func (s *Service) safeReturn(candidate string) string {
	if slices.Contains(s.config.AllowedReturns, candidate) {
		return candidate
	}
	return s.config.DefaultReturn
}
