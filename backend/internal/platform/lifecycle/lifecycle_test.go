package lifecycle

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

type recordingReadiness struct{ events *[]string }

func (r recordingReadiness) SetReady(ready bool) {
	if ready {
		*r.events = append(*r.events, "ready")
		return
	}
	*r.events = append(*r.events, "not-ready")
}

type recordingServer struct {
	events *[]string
	err    error
}

func (s recordingServer) Shutdown(context.Context) error {
	*s.events = append(*s.events, "http-drained")
	return s.err
}

func TestDrainOrdersReadinessHTTPAndDatabase(t *testing.T) {
	events := []string{}
	err := Drain(
		context.Background(),
		recordingReadiness{events: &events},
		recordingServer{events: &events},
		func() { events = append(events, "database-closed") },
	)

	if err != nil {
		t.Fatalf("Drain() error = %v", err)
	}
	want := []string{"not-ready", "http-drained", "database-closed"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestDrainReturnsShutdownFailureAfterClosingDatabase(t *testing.T) {
	events := []string{}
	wantErr := context.DeadlineExceeded
	err := Drain(
		context.Background(),
		recordingReadiness{events: &events},
		recordingServer{events: &events, err: wantErr},
		func() { events = append(events, "database-closed") },
	)

	if !errors.Is(err, wantErr) {
		t.Fatalf("Drain() error = %v, want %v", err, wantErr)
	}
	if got := events[len(events)-1]; got != "database-closed" {
		t.Fatalf("last event = %q, want database-closed", got)
	}
}
