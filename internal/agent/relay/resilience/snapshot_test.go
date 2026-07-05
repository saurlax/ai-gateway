package resilience

import (
	"testing"

	"github.com/VaalaCat/ai-gateway/internal/agent/relay/state"
)

func TestSnapshotBreakers_StatesAndSort(t *testing.T) {
	r := NewRegistry()
	r.Get(adminKey(3), testCfg()) // closed
	cbOpen := r.Get(adminKey(1), testCfg())
	r.Get(adminKey(2), testCfg()) // closed
	cbOpen.RecordFailure()
	cbOpen.RecordFailure() // threshold=2 → open
	if !cbOpen.IsOpen() {
		t.Fatal("setup: breaker 1 should be open")
	}

	snap := r.SnapshotBreakers()
	if len(snap) != 3 {
		t.Fatalf("want 3 snapshots, got %d", len(snap))
	}
	// 按 (Source, ChannelID) 稳定排序
	if snap[0].ChannelID != 1 || snap[1].ChannelID != 2 || snap[2].ChannelID != 3 {
		t.Fatalf("not sorted by ChannelID: %+v", snap)
	}
	if snap[0].Source != "admin" {
		t.Fatalf("source = %q, want admin", snap[0].Source)
	}
	if snap[0].State != "open" {
		t.Fatalf("breaker 1 state = %q, want open", snap[0].State)
	}
	if snap[0].Failures < 2 {
		t.Fatalf("breaker 1 failures = %d, want >=2", snap[0].Failures)
	}
	if snap[0].RemainingMs <= 0 {
		t.Fatalf("open breaker should have remaining cooldown, got %d", snap[0].RemainingMs)
	}
	if snap[1].State != "closed" || snap[2].State != "closed" {
		t.Fatalf("breakers 2,3 should be closed: %+v", snap)
	}
}

func TestSnapshotBreakers_HalfOpen(t *testing.T) {
	r := NewRegistry()
	cb := r.Get(BreakerKey{Source: state.SourcePrivate, ID: 9}, testCfg())
	cb.HalfOpen()

	snap := r.SnapshotBreakers()
	if len(snap) != 1 || snap[0].State != "half-open" {
		t.Fatalf("want single half-open snapshot, got %+v", snap)
	}
	if snap[0].Source != "private" || snap[0].ChannelID != 9 {
		t.Fatalf("key wrong: %+v", snap[0])
	}
}

func TestSnapshotBreakers_Empty(t *testing.T) {
	if s := NewRegistry().SnapshotBreakers(); len(s) != 0 {
		t.Fatalf("empty registry → empty snapshot, got %+v", s)
	}
}
