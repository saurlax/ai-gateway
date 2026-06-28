package plan

import (
	"testing"

	"github.com/VaalaCat/ai-gateway/internal/agent/relay/state"
	"github.com/VaalaCat/ai-gateway/internal/models"
	"github.com/VaalaCat/ai-gateway/internal/pkg/app"
)

func byokAdminCand() ScoredCandidate {
	return ScoredCandidate{Channel: &models.Channel{}, Source: state.SourceAdmin}
}
func byokPrivateCand() ScoredCandidate {
	return ScoredCandidate{Channel: &models.Channel{}, Source: state.SourcePrivate}
}

func byokFctx(ui *app.UserInfo) *FilterContext {
	return &FilterContext{Rctx: withRequest(newTestRelayContext(&stubAgentCache{}, "m", ui, 0)), RealModel: "m"}
}

// 1. byok_only=false → 原样放行（行为不变）。
func TestBYOKOnlyFilter_Disabled_AllKept(t *testing.T) {
	in := []ScoredCandidate{byokPrivateCand(), byokAdminCand()}
	got, code := byokOnlyFilter{}.Apply(byokFctx(&app.UserInfo{UserID: 7, BYOKOnly: false}), in)
	if len(got) != 2 || code != DropNone {
		t.Fatalf("got len=%d code=%v, want len=2 DropNone", len(got), code)
	}
}

// 2. byok_only=true + 私有+共享混合 → 只留私有。
func TestBYOKOnlyFilter_Enabled_KeepsOnlyPrivate(t *testing.T) {
	in := []ScoredCandidate{byokPrivateCand(), byokAdminCand(), byokPrivateCand()}
	got, code := byokOnlyFilter{}.Apply(byokFctx(&app.UserInfo{UserID: 7, BYOKOnly: true}), in)
	if code != DropNone {
		t.Fatalf("code = %v, want DropNone", code)
	}
	if len(got) != 2 {
		t.Fatalf("kept len = %d, want 2 (only private)", len(got))
	}
	for _, c := range got {
		if c.Source != state.SourcePrivate {
			t.Errorf("kept candidate Source=%v, want SourcePrivate", c.Source)
		}
	}
}

// 3. byok_only=true + 全共享（无私有）→ 空 + DropBYOKOnly。
func TestBYOKOnlyFilter_Enabled_NoPrivate_Dropped(t *testing.T) {
	in := []ScoredCandidate{byokAdminCand(), byokAdminCand()}
	got, code := byokOnlyFilter{}.Apply(byokFctx(&app.UserInfo{UserID: 7, BYOKOnly: true}), in)
	if len(got) != 0 || code != DropBYOKOnly {
		t.Fatalf("got len=%d code=%v, want len=0 DropBYOKOnly", len(got), code)
	}
}

// 4. 边界：系统 token（UserInfo nil）→ 放行（即使候选全共享也不拦）。
func TestBYOKOnlyFilter_SystemToken_AllKept(t *testing.T) {
	in := []ScoredCandidate{byokAdminCand(), byokAdminCand()}
	got, code := byokOnlyFilter{}.Apply(byokFctx(nil), in)
	if len(got) != 2 || code != DropNone {
		t.Fatalf("got len=%d code=%v, want len=2 DropNone", len(got), code)
	}
}
