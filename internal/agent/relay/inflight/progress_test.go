package inflight

import (
	"testing"

	"github.com/VaalaCat/ai-gateway/internal/models"
)

func TestEntry_SetCurrentAttempt_ShownInSnapshot(t *testing.T) {
	reg := NewRegistry(nil, 0)
	e := reg.Track(Meta{ReqID: "r1"})
	e.SetCurrentAttempt(&AttemptInProgress{Seq: 1, ChannelID: 7, ChannelName: "c7", Source: "admin", RealModel: "gpt"})

	s := reg.Snapshot()
	if len(s) != 1 || s[0].CurrentAttempt == nil {
		t.Fatalf("current attempt missing from snapshot: %+v", s)
	}
	if s[0].CurrentAttempt.ChannelName != "c7" || s[0].CurrentAttempt.ChannelID != 7 {
		t.Fatalf("current attempt wrong: %+v", s[0].CurrentAttempt)
	}
}

func TestEntry_UpdateFallbackChain_SetsChainAndClearsCurrent(t *testing.T) {
	reg := NewRegistry(nil, 0)
	e := reg.Track(Meta{ReqID: "r1"})
	e.SetCurrentAttempt(&AttemptInProgress{Seq: 1, ChannelName: "c1"})
	e.UpdateFallbackChain([]models.AttemptRecord{{Seq: 1, ChannelName: "c1", Status: "fail"}})

	s := reg.Snapshot()[0]
	if s.CurrentAttempt != nil {
		t.Fatalf("current attempt should be cleared after UpdateFallbackChain, got %+v", s.CurrentAttempt)
	}
	if len(s.View.FallbackChain) != 1 || s.View.FallbackChain[0].ChannelName != "c1" {
		t.Fatalf("fallback chain not written: %+v", s.View.FallbackChain)
	}
}

func TestEntry_ClearCurrentAttempt(t *testing.T) {
	reg := NewRegistry(nil, 0)
	e := reg.Track(Meta{ReqID: "r1"})
	e.SetCurrentAttempt(&AttemptInProgress{Seq: 1, ChannelName: "c1"})
	e.ClearCurrentAttempt()
	if reg.Snapshot()[0].CurrentAttempt != nil {
		t.Fatalf("current attempt should be nil after ClearCurrentAttempt")
	}
}

func TestEntry_ProgressMethods_NilReceiverSafe(t *testing.T) {
	var e *Entry
	e.SetCurrentAttempt(&AttemptInProgress{})
	e.UpdateFallbackChain(nil)
	e.ClearCurrentAttempt() // 必须不 panic
}
