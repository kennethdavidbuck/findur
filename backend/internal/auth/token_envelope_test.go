package auth

import (
	"bytes"
	"testing"

	"github.com/google/uuid"
)

func TestTokenCipherPreservesOwnerProviderKindAndVersionAAD(t *testing.T) {
	owner := uuid.New()
	cipher, err := NewTokenCipher(SnapTradeProvider, map[int][]byte{1: bytes.Repeat([]byte{9}, 32)}, 1, bytes.NewReader(bytes.Repeat([]byte{2}, 64)))
	if err != nil {
		t.Fatal(err)
	}
	envelope, err := cipher.EncryptAccess(owner, "synthetic-access-token")
	if err != nil {
		t.Fatal(err)
	}
	plain, err := cipher.DecryptAccess(owner, 1, envelope)
	if err != nil || plain != "synthetic-access-token" {
		t.Fatalf("plain=%q err=%v", plain, err)
	}
	if _, err := cipher.DecryptAccess(uuid.New(), 1, envelope); err == nil {
		t.Fatal("token envelope opened for a different owner")
	}
	if _, err := cipher.DecryptAccess(owner, 2, envelope); err == nil {
		t.Fatal("token envelope opened with an unknown version")
	}
}
