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
	"time"

	"github.com/google/uuid"
)

const (
	SessionIdleLifetime     = 30 * time.Minute
	SessionAbsoluteLifetime = 12 * time.Hour
)

var (
	ErrRestartRequired = errors.New("authorization must be restarted")
	ErrTerminalization = errors.New("authorization recovery failed")
)

type CallbackClaim struct {
	StateHash         []byte
	NonceHash         []byte
	EncryptedVerifier []byte
	ReturnRoute       string
	TerminalOutcome   string
	TerminalRoute     string
	UserID            uuid.UUID
}

type CallbackRepository interface {
	ClaimCallback(context.Context, []byte, []byte, time.Time) (CallbackClaim, error)
	FindActiveUser(context.Context, string, string) (uuid.UUID, bool, error)
	FinalizeCallback(context.Context, Finalization) error
	FailCallback(context.Context, []byte, time.Time) error
	SessionActive(context.Context, []byte, time.Time) (bool, error)
}

type TokenSet struct {
	IDToken, AccessToken, RefreshToken string
	Expiry                             time.Time
}

type Identity struct{ Subject, Nonce string }

type OIDCClient interface {
	Exchange(context.Context, string, string, string) (TokenSet, error)
	Verify(context.Context, string) (Identity, error)
	Revoke(context.Context, string) error
}

type Finalization struct {
	StateHash, SessionHash, CSRFHash []byte
	Provider, Subject                string
	UserID                           uuid.UUID
	AccessToken, RefreshToken        []byte
	EnvelopeVersion                  int
	TokenExpiresAt                   time.Time
	SessionIdleExpiresAt             time.Time
	SessionAbsoluteExpiresAt         time.Time
	CompletedAt                      time.Time
	ReturnRoute                      string
}

type CallbackInput struct {
	State, Binding, Code, ProviderError, ExistingSession string
}

type CallbackResult struct {
	Route, Session, CSRF string
	Success              bool
}

type AuthorizationStatus struct {
	AuthorizationAvailable bool
	Authenticated          bool
}

type CallbackConfig struct {
	Provider, CallbackURL string
	AttemptHashKey        []byte
	SessionHashKey        []byte
	VerifierKey           []byte
	TokenKeys             map[int][]byte
	CurrentTokenKey       int
	Random                io.Reader
	Clock                 func() time.Time
	OperationTimeout      time.Duration
}

type CallbackService struct {
	config CallbackConfig
	repo   CallbackRepository
	oidc   OIDCClient
	verify cipher.AEAD
	tokens map[int]cipher.AEAD
}

func NewCallbackService(cfg CallbackConfig, repo CallbackRepository, client OIDCClient) (*CallbackService, error) {
	if repo == nil || client == nil || cfg.Provider == "" || cfg.CallbackURL == "" || cfg.Random == nil || cfg.Clock == nil || cfg.OperationTimeout <= 0 || len(cfg.AttemptHashKey) < 32 || len(cfg.SessionHashKey) < 32 || len(cfg.VerifierKey) != 32 {
		return nil, errors.New("incomplete callback service configuration")
	}
	if subtle.ConstantTimeCompare(cfg.AttemptHashKey, cfg.VerifierKey) == 1 || subtle.ConstantTimeCompare(cfg.SessionHashKey, cfg.AttemptHashKey) == 1 || subtle.ConstantTimeCompare(cfg.SessionHashKey, cfg.VerifierKey) == 1 {
		return nil, errors.New("callback keys must be independent")
	}
	verify, err := newAEAD(cfg.VerifierKey)
	if err != nil {
		return nil, err
	}
	tokens := make(map[int]cipher.AEAD, len(cfg.TokenKeys))
	for version, key := range cfg.TokenKeys {
		if version < 1 || len(key) != 32 {
			return nil, errors.New("invalid token key ring")
		}
		if subtle.ConstantTimeCompare(key, cfg.AttemptHashKey) == 1 || subtle.ConstantTimeCompare(key, cfg.SessionHashKey) == 1 || subtle.ConstantTimeCompare(key, cfg.VerifierKey) == 1 {
			return nil, errors.New("token encryption keys must be independent")
		}
		tokens[version], err = newAEAD(key)
		if err != nil {
			return nil, err
		}
	}
	if tokens[cfg.CurrentTokenKey] == nil {
		return nil, errors.New("current token key is unavailable")
	}
	return &CallbackService{config: cfg, repo: repo, oidc: client, verify: verify, tokens: tokens}, nil
}

func newAEAD(key []byte) (cipher.AEAD, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}

func (s *CallbackService) Complete(ctx context.Context, in CallbackInput) (CallbackResult, error) {
	if in.State == "" || in.Binding == "" {
		return CallbackResult{Route: "/connect/result"}, ErrRestartRequired
	}
	now := s.config.Clock().UTC()
	claim, err := s.repo.ClaimCallback(ctx, s.attemptHash(in.State), s.attemptHash(in.Binding), now)
	if err != nil {
		return CallbackResult{Route: "/connect/result"}, ErrRestartRequired
	}
	if claim.TerminalOutcome != "" {
		if claim.TerminalOutcome == "succeeded" && in.ExistingSession != "" {
			active, checkErr := s.repo.SessionActive(ctx, s.sessionHash(in.ExistingSession), now)
			if checkErr == nil && active {
				return CallbackResult{Route: claim.TerminalRoute, Success: true}, nil
			}
		}
		return CallbackResult{Route: "/connect/result"}, ErrRestartRequired
	}
	if in.ProviderError != "" || in.Code == "" {
		return s.fail(ctx, claim.StateHash, now, "")
	}
	verifier, err := s.decryptVerifier(claim.StateHash, claim.EncryptedVerifier)
	if err != nil {
		return s.fail(ctx, claim.StateHash, now, "")
	}
	opCtx, cancel := context.WithTimeout(ctx, s.config.OperationTimeout)
	defer cancel()
	tokenSet, err := s.oidc.Exchange(opCtx, in.Code, s.config.CallbackURL, verifier)
	defer clearStrings(&tokenSet)
	if err != nil {
		return s.fail(ctx, claim.StateHash, now, tokenSet.AccessToken)
	}
	if tokenSet.IDToken == "" || tokenSet.AccessToken == "" {
		return s.fail(ctx, claim.StateHash, now, tokenSet.AccessToken)
	}
	identity, err := s.oidc.Verify(opCtx, tokenSet.IDToken)
	cancel()
	if err != nil || identity.Subject == "" || subtle.ConstantTimeCompare(claim.NonceHash, s.attemptHash(identity.Nonce)) != 1 {
		return s.fail(ctx, claim.StateHash, now, tokenSet.AccessToken)
	}
	userID, found, err := s.repo.FindActiveUser(ctx, s.config.Provider, identity.Subject)
	if err != nil {
		return s.fail(ctx, claim.StateHash, now, tokenSet.AccessToken)
	}
	if !found {
		userID = uuid.New()
	}
	session, err := s.secret()
	if err != nil {
		return s.failIssued(ctx, claim.StateHash, now, tokenSet.AccessToken)
	}
	csrf, err := s.secret()
	if err != nil {
		return s.failIssued(ctx, claim.StateHash, now, tokenSet.AccessToken)
	}
	access, err := s.encryptToken(userID, "access", tokenSet.AccessToken)
	if err != nil {
		return s.failIssued(ctx, claim.StateHash, now, tokenSet.AccessToken)
	}
	var refresh []byte
	if tokenSet.RefreshToken != "" {
		refresh, err = s.encryptToken(userID, "refresh", tokenSet.RefreshToken)
		if err != nil {
			return s.failIssued(ctx, claim.StateHash, now, tokenSet.AccessToken)
		}
	}
	finalNow := s.config.Clock().UTC()
	final := Finalization{StateHash: claim.StateHash, Provider: s.config.Provider, Subject: identity.Subject, UserID: userID,
		AccessToken: access, RefreshToken: refresh, EnvelopeVersion: s.config.CurrentTokenKey, TokenExpiresAt: tokenSet.Expiry,
		SessionHash: s.sessionHash(session), CSRFHash: s.sessionHash(csrf), SessionIdleExpiresAt: finalNow.Add(SessionIdleLifetime), SessionAbsoluteExpiresAt: finalNow.Add(SessionAbsoluteLifetime), CompletedAt: finalNow, ReturnRoute: claim.ReturnRoute}
	if err := s.repo.FinalizeCallback(ctx, final); err != nil {
		return s.fail(ctx, claim.StateHash, finalNow, tokenSet.AccessToken)
	}
	return CallbackResult{Route: "/connect/result", Session: session, CSRF: csrf, Success: true}, nil
}

// Status returns only categorical browser state. It never resolves or exposes an owner.
func (s *CallbackService) Status(ctx context.Context, session string) (AuthorizationStatus, error) {
	result := AuthorizationStatus{AuthorizationAvailable: true}
	if session == "" {
		return result, nil
	}
	active, err := s.repo.SessionActive(ctx, s.sessionHash(session), s.config.Clock().UTC())
	if err != nil {
		return result, err
	}
	result.Authenticated = active
	return result, nil
}

func (s *CallbackService) failIssued(ctx context.Context, stateHash []byte, now time.Time, token string) (CallbackResult, error) {
	return s.fail(ctx, stateHash, now, token)
}

func (s *CallbackService) fail(ctx context.Context, stateHash []byte, now time.Time, token string) (CallbackResult, error) {
	cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), s.config.OperationTimeout)
	defer cancel()
	failErr := s.repo.FailCallback(cleanupCtx, stateHash, now)
	if token != "" {
		_ = s.oidc.Revoke(cleanupCtx, token)
	}
	if failErr != nil {
		return CallbackResult{Route: "/connect/result"}, ErrTerminalization
	}
	return CallbackResult{Route: "/connect/result"}, ErrRestartRequired
}

func (s *CallbackService) attemptHash(value string) []byte {
	return keyedHash(s.config.AttemptHashKey, value)
}
func (s *CallbackService) sessionHash(value string) []byte {
	return keyedHash(s.config.SessionHashKey, value)
}
func keyedHash(key []byte, value string) []byte {
	d := hmac.New(sha256.New, key)
	_, _ = d.Write([]byte(value))
	return d.Sum(nil)
}

func (s *CallbackService) secret() (string, error) {
	b := make([]byte, randomBytes)
	if _, err := io.ReadFull(s.config.Random, b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func (s *CallbackService) decryptVerifier(stateHash, envelope []byte) (string, error) {
	if len(envelope) < s.verify.NonceSize() {
		return "", errors.New("invalid verifier envelope")
	}
	plain, err := s.verify.Open(nil, envelope[:s.verify.NonceSize()], envelope[s.verify.NonceSize():], stateHash)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}

func (s *CallbackService) encryptToken(owner uuid.UUID, kind, token string) ([]byte, error) {
	aead := s.tokens[s.config.CurrentTokenKey]
	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(s.config.Random, nonce); err != nil {
		return nil, err
	}
	aad := []byte(fmt.Sprintf("%s|%s|%s|%d", owner.String(), s.config.Provider, kind, s.config.CurrentTokenKey))
	return aead.Seal(nonce, nonce, []byte(token), aad), nil
}

func clearStrings(tokens *TokenSet) {
	tokens.IDToken, tokens.AccessToken, tokens.RefreshToken = "", "", ""
}
