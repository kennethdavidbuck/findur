package auth

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
)

const (
	testState          = "state"
	testNonce          = "nonce"
	testBinding        = "binding"
	testCode           = "code"
	testSubject        = "subject"
	testLoopbackOrigin = "http://127.0.0.1:8080"
)

type callbackRepoStub struct {
	claim          CallbackClaim
	claimErr       error
	claims         int
	user           uuid.UUID
	found, active  bool
	final          Finalization
	finalErr       error
	failed         int
	sessionChecked []byte
}

func (r *callbackRepoStub) ClaimCallback(context.Context, []byte, []byte, time.Time) (CallbackClaim, error) {
	r.claims++
	return r.claim, r.claimErr
}
func (r *callbackRepoStub) FindActiveUser(context.Context, string, string) (uuid.UUID, bool, error) {
	return r.user, r.found, nil
}
func (r *callbackRepoStub) FinalizeCallback(_ context.Context, f Finalization) error {
	r.final = f
	return r.finalErr
}
func (r *callbackRepoStub) FailCallback(context.Context, []byte, time.Time) error {
	r.failed++
	return nil
}
func (r *callbackRepoStub) SessionActive(_ context.Context, sessionHash []byte, _ time.Time) (bool, error) {
	r.sessionChecked = append([]byte(nil), sessionHash...)
	return r.active, nil
}

type oidcStub struct {
	identity               Identity
	exchangeErr, verifyErr error
	exchanges, revokes     int
	redirectURI            string
}

func (o *oidcStub) Exchange(_ context.Context, _, redirectURI, _ string) (TokenSet, error) {
	o.exchanges++
	o.redirectURI = redirectURI
	return TokenSet{IDToken: "id", AccessToken: "access", RefreshToken: "refresh", Expiry: time.Unix(2000, 0)}, o.exchangeErr
}
func (o *oidcStub) Verify(context.Context, string) (Identity, error) { return o.identity, o.verifyErr }
func (o *oidcStub) Revoke(context.Context, string) error             { o.revokes++; return nil }

func callbackFixture(t *testing.T) (*CallbackService, *callbackRepoStub, *oidcStub) {
	t.Helper()
	attemptKey, verifierKey := bytes.Repeat([]byte{1}, 32), bytes.Repeat([]byte{2}, 32)
	sessionKey, tokenKey := bytes.Repeat([]byte{3}, 32), bytes.Repeat([]byte{4}, 32)
	block, _ := aes.NewCipher(verifierKey)
	aead, _ := cipher.NewGCM(block)
	nonce := bytes.Repeat([]byte{9}, aead.NonceSize())
	stateHash := keyedHash(attemptKey, testState)
	envelope := aead.Seal(nonce, nonce, []byte("verifier"), stateHash)
	repo := &callbackRepoStub{claim: CallbackClaim{
		StateHash:         stateHash,
		NonceHash:         keyedHash(attemptKey, testNonce),
		EncryptedVerifier: envelope,
		ReturnRoute:       PortfolioReturnRoute,
	}}
	client := &oidcStub{identity: Identity{Subject: testSubject, Nonce: testNonce}}
	config := CallbackConfig{
		Provider:         SnapTradeProvider,
		CallbackURL:      testLoopbackOrigin + SnapTradeCallbackPath,
		AttemptHashKey:   attemptKey,
		SessionHashKey:   sessionKey,
		VerifierKey:      verifierKey,
		TokenKeys:        map[int][]byte{1: tokenKey},
		CurrentTokenKey:  1,
		Random:           bytes.NewReader(bytes.Repeat([]byte{7}, 256)),
		Clock:            func() time.Time { return time.Unix(1000, 0) },
		OperationTimeout: time.Second,
	}
	service, err := NewCallbackService(config, repo, client)
	if err != nil {
		t.Fatal(err)
	}
	return service, repo, client
}

func TestCallbackSuccessCreatesOpaqueEncryptedSessionMaterial(t *testing.T) {
	service, repo, client := callbackFixture(t)
	result, err := service.Complete(context.Background(), CallbackInput{State: testState, Binding: testBinding, Code: testCode})
	if err != nil || !result.Success || result.Session == "" || result.CSRF == "" {
		t.Fatalf("result=%+v err=%v", result, err)
	}
	if client.exchanges != 1 || client.redirectURI != "http://127.0.0.1:8080/api/auth/snaptrade/callback" || repo.final.Subject != "subject" || len(repo.final.SessionHash) != 32 || bytes.Contains(repo.final.AccessToken, []byte("access")) {
		t.Fatalf("unsafe finalization: %+v", repo.final)
	}
	block, _ := aes.NewCipher(bytes.Repeat([]byte{4}, 32))
	envelope, _ := cipher.NewGCM(block)
	aad := []byte(fmt.Sprintf("%s|snaptrade|access|1", repo.final.UserID))
	plain, err := envelope.Open(nil, repo.final.AccessToken[:envelope.NonceSize()], repo.final.AccessToken[envelope.NonceSize():], aad)
	if err != nil || string(plain) != "access" || bytes.Equal(repo.final.AccessToken, repo.final.RefreshToken) {
		t.Fatalf("token envelope invalid: %q %v", plain, err)
	}
	if result.Route != "/connect/result" {
		t.Fatalf("route=%q", result.Route)
	}
}

func TestCallbackInputFailuresNeverExchange(t *testing.T) {
	tests := map[string]struct {
		input                  CallbackInput
		claimErr               error
		wantClaims, wantFailed int
	}{
		"missing state":      {input: CallbackInput{Binding: testBinding, Code: testCode}},
		"missing code":       {input: CallbackInput{State: testState, Binding: testBinding}, wantClaims: 1, wantFailed: 1},
		"mismatched binding": {input: CallbackInput{State: testState, Binding: "wrong", Code: testCode}, claimErr: ErrNotClaimable, wantClaims: 1},
		"expired attempt":    {input: CallbackInput{State: testState, Binding: testBinding, Code: testCode}, claimErr: ErrNotClaimable, wantClaims: 1},
		"provider denial":    {input: CallbackInput{State: testState, Binding: testBinding, ProviderError: "access_denied"}, wantClaims: 1, wantFailed: 1},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			service, repo, client := callbackFixture(t)
			repo.claimErr = test.claimErr
			result, err := service.Complete(context.Background(), test.input)
			if !errors.Is(err, ErrRestartRequired) || result.Success || client.exchanges != 0 || repo.claims != test.wantClaims || repo.failed != test.wantFailed || repo.final.UserID != uuid.Nil {
				t.Fatalf("result=%+v err=%v exchanges=%d claims=%d failed=%d final=%+v", result, err, client.exchanges, repo.claims, repo.failed, repo.final)
			}
		})
	}
}

func TestIssuedTokenValidationFailuresCreateNoStateAndCompensateOnce(t *testing.T) {
	for _, name := range []string{"invalid signature", "invalid issuer", "invalid audience", "expired token", "implausible issued-at"} {
		t.Run(name, func(t *testing.T) {
			service, repo, client := callbackFixture(t)
			client.verifyErr = errors.New(name)
			result, err := service.Complete(context.Background(), CallbackInput{State: testState, Binding: testBinding, Code: testCode})
			if !errors.Is(err, ErrRestartRequired) || result.Success || client.exchanges != 1 || client.revokes != 1 || repo.failed != 1 || repo.final.UserID != uuid.Nil {
				t.Fatalf("result=%+v err=%v client=%+v final=%+v", result, err, client, repo.final)
			}
		})
	}
	t.Run("invalid nonce", func(t *testing.T) {
		service, repo, client := callbackFixture(t)
		client.identity.Nonce = "wrong"
		_, err := service.Complete(context.Background(), CallbackInput{State: testState, Binding: testBinding, Code: testCode})
		if !errors.Is(err, ErrRestartRequired) || client.revokes != 1 || repo.failed != 1 || repo.final.UserID != uuid.Nil {
			t.Fatalf("err=%v client=%+v final=%+v", err, client, repo.final)
		}
	})
}

func TestCallbackFinalizationRollbackCompensatesOnce(t *testing.T) {
	service, repo, client := callbackFixture(t)
	repo.finalErr = errors.New("rollback")
	result, err := service.Complete(context.Background(), CallbackInput{State: testState, Binding: testBinding, Code: testCode})
	if !errors.Is(err, ErrRestartRequired) || result.Success || repo.failed != 1 || client.revokes != 1 {
		t.Fatalf("result=%+v err=%v failed=%d revokes=%d", result, err, repo.failed, client.revokes)
	}
}

func TestCallbackFailuresAreTerminalAndCompensatedOnce(t *testing.T) {
	tests := map[string]struct {
		mutate      func(*callbackRepoStub, *oidcStub)
		wantRevokes int
	}{
		"invalid token": {func(_ *callbackRepoStub, o *oidcStub) { o.verifyErr = errors.New("bad signature") }, 1},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			service, repo, client := callbackFixture(t)
			test.mutate(repo, client)
			input := CallbackInput{State: testState, Binding: testBinding, Code: testCode}
			if name == "provider denial" {
				input.Code, input.ProviderError = "", "access_denied"
			}
			result, err := service.Complete(context.Background(), input)
			if !errors.Is(err, ErrRestartRequired) || result.Success || repo.failed != 1 || client.revokes != test.wantRevokes {
				t.Fatalf("result=%+v err=%v failed=%d revokes=%d", result, err, repo.failed, client.revokes)
			}
		})
	}
}

func TestCallbackReplayNeverExchangesAgain(t *testing.T) {
	service, repo, client := callbackFixture(t)
	repo.claim.TerminalOutcome = "succeeded"
	repo.claim.TerminalRoute = "/connect/result"
	repo.active = true
	result, err := service.Complete(context.Background(), CallbackInput{State: testState, Binding: testBinding, Code: testCode, ExistingSession: "session"})
	if err != nil || result.Route != "/connect/result" || client.exchanges != 0 {
		t.Fatalf("result=%+v err=%v exchanges=%d", result, err, client.exchanges)
	}
}

func TestCallbackReplayWithoutActiveSessionRequiresRestart(t *testing.T) {
	service, repo, client := callbackFixture(t)
	repo.claim.TerminalOutcome, repo.claim.TerminalRoute = "succeeded", "/connect/result"
	result, err := service.Complete(context.Background(), CallbackInput{State: testState, Binding: testBinding, Code: testCode})
	if !errors.Is(err, ErrRestartRequired) || result.Route != "/connect/result" || client.exchanges != 0 || repo.final.UserID != uuid.Nil {
		t.Fatalf("result=%+v err=%v exchanges=%d", result, err, client.exchanges)
	}
}

func TestAuthorizationStatusUsesOnlyHashedOpaqueSession(t *testing.T) {
	service, repo, _ := callbackFixture(t)
	repo.active = true
	status, err := service.Status(context.Background(), "opaque-session")
	if err != nil || !status.AuthorizationAvailable || !status.Authenticated || bytes.Equal(repo.sessionChecked, []byte("opaque-session")) || len(repo.sessionChecked) != 32 {
		t.Fatalf("status=%+v err=%v hash=%x", status, err, repo.sessionChecked)
	}
	status, err = service.Status(context.Background(), "")
	if err != nil || !status.AuthorizationAvailable || status.Authenticated {
		t.Fatalf("empty status=%+v err=%v", status, err)
	}
}

func TestCallbackCreatesNewIdentityWhenNoneRemains(t *testing.T) {
	service, repo, _ := callbackFixture(t)
	if _, err := service.Complete(context.Background(), CallbackInput{State: testState, Binding: testBinding, Code: testCode}); err != nil {
		t.Fatal(err)
	}
	if repo.final.UserID == uuid.Nil || repo.final.Subject != "subject" {
		t.Fatalf("final=%+v", repo.final)
	}
}

func TestCallbackResumesExistingActiveIdentity(t *testing.T) {
	service, repo, _ := callbackFixture(t)
	repo.user, repo.found = uuid.New(), true
	if _, err := service.Complete(context.Background(), CallbackInput{State: testState, Binding: testBinding, Code: testCode}); err != nil {
		t.Fatal(err)
	}
	if repo.final.UserID != repo.user {
		t.Fatalf("created %s instead of resuming %s", repo.final.UserID, repo.user)
	}
}
