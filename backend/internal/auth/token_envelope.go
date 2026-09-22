package auth

import (
	"crypto/cipher"
	"errors"
	"fmt"
	"io"

	"github.com/google/uuid"
)

// TokenCipher preserves the versioned AES-GCM envelope contract used by OAuth
// callback persistence while allowing later provider operations to decrypt only
// the access token they need.
type TokenCipher struct {
	provider string
	current  int
	random   io.Reader
	keys     map[int]cipher.AEAD
}

// NewTokenCipher validates and constructs the versioned provider token envelope.
func NewTokenCipher(provider string, keyRing map[int][]byte, current int, random io.Reader) (*TokenCipher, error) {
	if provider == "" || random == nil || current < 1 {
		return nil, errors.New("incomplete token cipher configuration")
	}
	keys := make(map[int]cipher.AEAD, len(keyRing))
	for version, key := range keyRing {
		if version < 1 || len(key) != 32 {
			return nil, errors.New("invalid token key ring")
		}
		aead, err := newAEAD(key)
		if err != nil {
			return nil, err
		}
		keys[version] = aead
	}
	if keys[current] == nil {
		return nil, errors.New("current token key is unavailable")
	}
	return &TokenCipher{provider: provider, current: current, random: random, keys: keys}, nil
}

// CurrentVersion is the envelope version used for new ciphertext.
func (c *TokenCipher) CurrentVersion() int { return c.current }

// EncryptAccess seals one access token with the established owner/provider/kind/version AAD.
func (c *TokenCipher) EncryptAccess(owner uuid.UUID, token string) ([]byte, error) {
	return c.encrypt(owner, tokenKindAccess, token)
}

// EncryptRefresh seals one refresh token with the established owner/provider/kind/version AAD.
func (c *TokenCipher) EncryptRefresh(owner uuid.UUID, token string) ([]byte, error) {
	return c.encrypt(owner, tokenKindRefresh, token)
}

// DecryptAccess opens a persisted access-token envelope without changing its AAD contract.
func (c *TokenCipher) DecryptAccess(owner uuid.UUID, version int, envelope []byte) (string, error) {
	return c.decrypt(owner, tokenKindAccess, version, envelope)
}

// DecryptRefresh opens a persisted refresh-token envelope without exposing it
// outside the server-side credential source.
func (c *TokenCipher) DecryptRefresh(owner uuid.UUID, version int, envelope []byte) (string, error) {
	return c.decrypt(owner, tokenKindRefresh, version, envelope)
}

func (c *TokenCipher) decrypt(owner uuid.UUID, kind string, version int, envelope []byte) (string, error) {
	aead := c.keys[version]
	if aead == nil || len(envelope) < aead.NonceSize() {
		return "", errors.New("invalid token envelope")
	}
	aad := []byte(fmt.Sprintf("%s|%s|%s|%d", owner.String(), c.provider, kind, version))
	plain, err := aead.Open(nil, envelope[:aead.NonceSize()], envelope[aead.NonceSize():], aad)
	if err != nil {
		return "", errors.New("invalid token envelope")
	}
	return string(plain), nil
}

func (c *TokenCipher) encrypt(owner uuid.UUID, kind, token string) ([]byte, error) {
	aead := c.keys[c.current]
	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(c.random, nonce); err != nil {
		return nil, err
	}
	aad := []byte(fmt.Sprintf("%s|%s|%s|%d", owner.String(), c.provider, kind, c.current))
	return aead.Seal(nonce, nonce, []byte(token), aad), nil
}
