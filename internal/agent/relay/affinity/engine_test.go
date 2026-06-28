package affinity

import (
	"testing"
	"time"

	"github.com/VaalaCat/ai-gateway/internal/agent/relay/state"
	"github.com/VaalaCat/ai-gateway/internal/settings"
)

type stubCfg struct{ s settings.AgentSettings }

func (c stubCfg) Settings() settings.AgentSettings { return c.s }

func TestEngine_DisabledDecide(t *testing.T) {
	e := New(stubCfg{settings.AgentSettings{AffinityEnabled: 0, AffinityTTLSec: 300}})
	d := e.Decide(Subject{UserID: 1, RealModel: "m"})
	if d.Apply || d.Record {
		t.Fatalf("disabled engine should not apply/record, got %+v", d)
	}
}

func TestEngine_EnabledRememberLookup(t *testing.T) {
	e := New(stubCfg{settings.AgentSettings{AffinityEnabled: 1, AffinityTTLSec: 300}})
	if d := e.Decide(Subject{UserID: 1, RealModel: "m"}); !d.Apply || !d.Record {
		t.Fatalf("enabled engine should apply+record, got %+v", d)
	}
	k := Key{UserID: 1, RealModel: "m"}
	e.Remember(k, state.SourceAdmin, 9, nil)
	got, ok := e.Lookup(k)
	if !ok || got.SourceID != 9 || got.Source != state.SourceAdmin {
		t.Fatalf("lookup = (%+v,%v), want admin/9/true", got, ok)
	}
	if !got.ExpiresAt.After(time.Now()) {
		t.Fatal("Remember should set future ExpiresAt from TTL")
	}
}

func TestEngine_RememberZeroTTLNoop(t *testing.T) {
	e := New(stubCfg{settings.AgentSettings{AffinityEnabled: 1, AffinityTTLSec: 0}})
	k := Key{UserID: 1, RealModel: "m"}
	e.Remember(k, state.SourceAdmin, 9, nil)
	if _, ok := e.Lookup(k); ok {
		t.Fatal("TTL<=0 should not record")
	}
}

func TestResolveEnabled(t *testing.T) {
	tru, fls := true, false
	cases := []struct {
		name     string
		flag     *bool
		globalOn bool
		want     bool
	}{
		{"nil inherits global on", nil, true, true},
		{"nil inherits global off", nil, false, false},
		{"true overrides global off", &tru, false, true},
		{"false overrides global on", &fls, true, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := resolveEnabled(c.flag, c.globalOn); got != c.want {
				t.Fatalf("resolveEnabled=%v want %v", got, c.want)
			}
		})
	}
}

func TestDecideChannelOverride(t *testing.T) {
	tru := true
	engOff := New(stubCfg{settings.AgentSettings{AffinityEnabled: 0, AffinityTTLSec: 300}}) // 全局关
	if d := engOff.Decide(Subject{UserID: 1, RealModel: "m", ChannelEnabled: &tru}); !d.Apply || !d.Record {
		t.Fatalf("channel force-on should apply+record even when global off, got %+v", d)
	}

	fls := false
	engOn := New(stubCfg{settings.AgentSettings{AffinityEnabled: 1, AffinityTTLSec: 300}}) // 全局开
	if d := engOn.Decide(Subject{UserID: 1, RealModel: "m", ChannelEnabled: &fls}); d.Apply || d.Record {
		t.Fatalf("channel force-off should not apply/record even when global on, got %+v", d)
	}
}
