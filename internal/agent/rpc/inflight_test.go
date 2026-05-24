package rpc

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/VaalaCat/ai-gateway/internal/agent/relay/inflight"
)

func TestHandleInflight_ReturnsSnapshot(t *testing.T) {
	reg := inflight.NewRegistry(nil, 0)
	e := reg.Track(inflight.Meta{ReqID: "x"})
	defer e.Done()
	res, err := HandleInflight(reg)
	if err != nil {
		t.Fatal(err)
	}
	b, _ := json.Marshal(res)
	if !json.Valid(b) || !strings.Contains(string(b), `"x"`) {
		t.Fatalf("bad inflight payload: %s", b)
	}
}

func TestHandleGoroutines_NonEmpty(t *testing.T) {
	res, err := HandleGoroutines()
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Dump) == 0 {
		t.Fatal("expected non-empty goroutine dump")
	}
}
