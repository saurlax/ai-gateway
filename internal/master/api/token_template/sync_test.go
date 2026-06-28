package token_template

import (
	"reflect"
	"testing"

	"github.com/VaalaCat/ai-gateway/internal/models"
)

func TestDiffToken(t *testing.T) {
	tpl := &models.TokenTemplate{
		ID:     12,
		Name:   "default",
		Models: `["gpt-4","gpt-5"]`,
	}
	tpl.AllowedChannelIDs = []uint{1, 2}

	t.Run("equal -> not changed", func(t *testing.T) {
		tok := &models.Token{ID: 100, Name: "tok", Models: `["gpt-5","gpt-4"]`}
		tok.AllowedChannelIDs = []uint{2, 1}
		changed, item := diffToken(tpl, tok, models.SyncFields{Models: true, Channels: true})
		if changed {
			t.Fatalf("expected not changed, got changed=true item=%+v", item)
		}
	})

	t.Run("models changed -> PreviewItem populated", func(t *testing.T) {
		tok := &models.Token{ID: 100, Name: "tok", Models: `["gpt-4"]`}
		tok.AllowedChannelIDs = []uint{1, 2}
		changed, item := diffToken(tpl, tok, models.SyncFields{Models: true, Channels: true})
		if !changed {
			t.Fatal("expected changed=true")
		}
		if item.TokenID != 100 || item.TokenName != "tok" {
			t.Fatalf("item id/name wrong: %+v", item)
		}
		if item.ModelsBefore != `["gpt-4"]` || item.ModelsAfter != `["gpt-4","gpt-5"]` {
			t.Fatalf("models before/after wrong: %+v", item)
		}
		if !reflect.DeepEqual(item.ChannelsBefore, []uint{1, 2}) {
			t.Fatalf("ChannelsBefore = %v", item.ChannelsBefore)
		}
		if !reflect.DeepEqual(item.ChannelsAfter, []uint{1, 2}) {
			t.Fatalf("ChannelsAfter = %v", item.ChannelsAfter)
		}
	})

	t.Run("channels changed -> PreviewItem populated", func(t *testing.T) {
		tok := &models.Token{ID: 101, Name: "tok2", Models: `["gpt-4","gpt-5"]`}
		tok.AllowedChannelIDs = []uint{1}
		changed, item := diffToken(tpl, tok, models.SyncFields{Models: true, Channels: true})
		if !changed {
			t.Fatal("expected changed=true")
		}
		if !reflect.DeepEqual(item.ChannelsBefore, []uint{1}) {
			t.Fatalf("ChannelsBefore = %v", item.ChannelsBefore)
		}
		if !reflect.DeepEqual(item.ChannelsAfter, []uint{1, 2}) {
			t.Fatalf("ChannelsAfter = %v", item.ChannelsAfter)
		}
	})
}

func TestDiffToken_FieldSelection(t *testing.T) {
	tpl := &models.TokenTemplate{ID: 12, Name: "t", Models: `["a"]`, BYOKOnly: true}
	tpl.AllowedChannelIDs = []uint{1}

	// 只选 models：channels/byok 差异不算 changed。
	tok := &models.Token{ID: 1, Name: "tk", Models: `["a"]`, BYOKOnly: false}
	tok.AllowedChannelIDs = []uint{9}
	if changed, _ := diffToken(tpl, tok, models.SyncFields{Models: true}); changed {
		t.Errorf("models-only: want not changed (only models compared, models equal)")
	}

	// 选 byok_only：byok 差异 → changed，且 before/after 填充。
	changed, item := diffToken(tpl, tok, models.SyncFields{BYOKOnly: true})
	if !changed {
		t.Fatal("byok selected & differs: want changed")
	}
	if item.BYOKOnlyBefore != false || item.BYOKOnlyAfter != true {
		t.Errorf("byok before/after = %v/%v, want false/true", item.BYOKOnlyBefore, item.BYOKOnlyAfter)
	}
}
