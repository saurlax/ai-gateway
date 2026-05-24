package inflight

import (
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestRegistry_TrackSnapshotDone(t *testing.T) {
	r := NewRegistry(nil, time.Hour)
	e := r.Track(Meta{ReqID: "r1", ChannelID: 7, ChannelName: "openai", Model: "gpt-4o", IsStream: true, StartTime: time.Now()})
	e.SetStage("upstream_dispatch")

	snaps := r.Snapshot()
	if len(snaps) != 1 {
		t.Fatalf("want 1 in-flight, got %d", len(snaps))
	}
	if snaps[0].ReqID != "r1" || snaps[0].ChannelName != "openai" || snaps[0].Stage != "upstream_dispatch" {
		t.Fatalf("snapshot mismatch: %+v", snaps[0])
	}
	if snaps[0].ElapsedMs < 0 {
		t.Fatalf("elapsed should be >= 0, got %d", snaps[0].ElapsedMs)
	}

	e.Done()
	if got := len(r.Snapshot()); got != 0 {
		t.Fatalf("want 0 after Done, got %d", got)
	}
}

func TestRegistry_WatchdogWarnsStuck(t *testing.T) {
	core, logs := observer.New(zap.WarnLevel)
	r := NewRegistry(zap.New(core), 20*time.Millisecond)

	stop := r.StartWatchdog(10 * time.Millisecond)
	defer stop()

	e := r.Track(Meta{ReqID: "stuck", StartTime: time.Now().Add(-time.Second)}) // 已"卡" 1s
	defer e.Done()

	time.Sleep(60 * time.Millisecond)
	if logs.FilterMessageSnippet("in-flight relay request stuck").Len() == 0 {
		t.Fatalf("expected watchdog WARN for stuck request")
	}
}
