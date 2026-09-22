package postgres_test

import (
	"sync"
	"testing"
	"time"

	"github.com/kennethdavidbuck/findur/backend/internal/auth"
	postgresadapter "github.com/kennethdavidbuck/findur/backend/internal/platform/postgres"
)

func TestCredentialRepositoryCoordinatesRotationAndFailsClosedOnExpiredLease(t *testing.T) {
	fixture := newRepositoryFixture(t)
	fixture.reset(t)
	owner := inclusionOwnerWithInventory(t, fixture, 249, "credential-owner")
	first := postgresadapter.NewCredentialRepository(fixture.pool)
	second := postgresadapter.NewCredentialRepository(fixture.pool)
	current, found, err := first.ReadCredential(fixture.ctx, owner)
	if err != nil || !found || current.Status != "active" {
		t.Fatalf("initial credential=%+v found=%v err=%v", current, found, err)
	}

	start := make(chan struct{})
	results := make(chan auth.Credential, 2)
	var group sync.WaitGroup
	for _, repository := range []*postgresadapter.CredentialRepository{first, second} {
		group.Add(1)
		go func(repo *postgresadapter.CredentialRepository) {
			defer group.Done()
			<-start
			claim, claimed, claimErr := repo.ClaimRefresh(fixture.ctx, owner, current.Version, fixture.now, time.Minute)
			if claimErr != nil {
				t.Errorf("claim: %v", claimErr)
			}
			if claimed {
				results <- claim
			}
		}(repository)
	}
	close(start)
	group.Wait()
	close(results)
	var winner auth.Credential
	count := 0
	for claim := range results {
		winner = claim
		count++
	}
	if count != 1 || winner.LeaseID == nil {
		t.Fatalf("refresh claims=%d, want one", count)
	}
	installed, err := second.InstallRefresh(fixture.ctx, winner, []byte("replacement-access-envelope"), []byte("replacement-refresh-envelope"), 1, fixture.now.Add(time.Hour), fixture.now)
	if err != nil || !installed {
		t.Fatalf("install accepted=%v err=%v", installed, err)
	}
	if installed, err = first.InstallRefresh(fixture.ctx, winner, []byte("stale"), []byte("stale"), 1, fixture.now.Add(time.Hour), fixture.now); err != nil || installed {
		t.Fatalf("stale install accepted=%v err=%v", installed, err)
	}
	rotated, found, err := first.ReadCredential(fixture.ctx, owner)
	if err != nil || !found || rotated.Version != current.Version+1 || string(rotated.AccessEnvelope) != "replacement-access-envelope" {
		t.Fatalf("rotated credential=%+v found=%v err=%v", rotated, found, err)
	}
	preSendClaim, claimed, err := first.ClaimRefresh(fixture.ctx, owner, rotated.Version, fixture.now, time.Minute)
	if err != nil || !claimed {
		t.Fatalf("expiry setup claimed=%v err=%v", claimed, err)
	}
	if released, err := second.ReleaseRefresh(fixture.ctx, preSendClaim, fixture.now); err != nil || !released {
		t.Fatalf("safe release accepted=%v err=%v", released, err)
	}
	if released, err := first.ReleaseRefresh(fixture.ctx, preSendClaim, fixture.now); err != nil || released {
		t.Fatalf("stale release accepted=%v err=%v", released, err)
	}
	if _, claimed, err := first.ClaimRefresh(fixture.ctx, owner, rotated.Version, fixture.now, time.Minute); err != nil || !claimed {
		t.Fatalf("claim after safe release=%v err=%v", claimed, err)
	}
	if _, err := fixture.pool.Exec(fixture.ctx, `UPDATE provider_authorizations SET refresh_lease_expires_at=$2 WHERE user_id=$1`, owner, fixture.now.Add(-time.Second)); err != nil {
		t.Fatal(err)
	}
	if _, claimed, err := second.ClaimRefresh(fixture.ctx, owner, rotated.Version, fixture.now, time.Minute); err != nil || claimed {
		t.Fatalf("expired lease replay claimed=%v err=%v", claimed, err)
	}
	failed, found, err := first.ReadCredential(fixture.ctx, owner)
	if err != nil || !found || failed.Status != "reauthorization-required" || failed.LeaseID != nil {
		t.Fatalf("failed credential=%+v found=%v err=%v", failed, found, err)
	}
}
