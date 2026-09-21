package portfolio

import (
	"testing"
	"time"
)

func TestClassifyFreshnessBoundaries(t *testing.T) {
	now := time.Date(2026, 9, 21, 12, 0, 0, 0, time.UTC)
	for _, test := range []struct {
		name       string
		mode       SyncMode
		activities bool
		age        time.Duration
		want       Freshness
	}{
		{"realtime current at 15 minutes", SyncModeRealtime, false, 15 * time.Minute, FreshnessCurrent},
		{"realtime stale after 15 minutes", SyncModeRealtime, false, 15*time.Minute + time.Second, FreshnessStale},
		{"daily current at 36 hours", SyncModeDelayed, false, 36 * time.Hour, FreshnessCurrent},
		{"holdings expired after 72 hours", SyncModeDelayed, false, 72*time.Hour + time.Second, FreshnessExpired},
		{"activities current at two days", SyncModeRealtime, true, 48 * time.Hour, FreshnessCurrent},
		{"activities stale through seven days", SyncModeRealtime, true, 7 * 24 * time.Hour, FreshnessStale},
		{"activities expired on the eighth calendar day", SyncModeRealtime, true, 8 * 24 * time.Hour, FreshnessExpired},
	} {
		t.Run(test.name, func(t *testing.T) {
			at := now.Add(-test.age)
			if got := ClassifyFreshness(test.mode, test.activities, &at, nil, now); got != test.want {
				t.Fatalf("got %q want %q", got, test.want)
			}
		})
	}
	if got := ClassifyFreshness(SyncModeUnknown, false, &now, nil, now); got != FreshnessUnavailable {
		t.Fatalf("got %q", got)
	}
	lateTwoCalendarDaysAgo := time.Date(2026, 9, 19, 0, 1, 0, 0, time.UTC)
	lateSevenCalendarDaysAgo := time.Date(2026, 9, 14, 0, 1, 0, 0, time.UTC)
	if got := ClassifyFreshness(SyncModeRealtime, true, &lateTwoCalendarDaysAgo, nil, time.Date(2026, 9, 21, 23, 59, 0, 0, time.UTC)); got != FreshnessCurrent {
		t.Fatalf("two calendar days got %q", got)
	}
	if got := ClassifyFreshness(SyncModeRealtime, true, &lateSevenCalendarDaysAgo, nil, time.Date(2026, 9, 21, 23, 59, 0, 0, time.UTC)); got != FreshnessStale {
		t.Fatalf("seven calendar days got %q", got)
	}
}
