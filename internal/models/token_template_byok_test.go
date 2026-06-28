package models

import "testing"

func TestTokenFieldsEqualForFields(t *testing.T) {
	tpl := &TokenTemplate{Models: `["a"]`, BYOKOnly: true}
	tpl.AllowedChannelIDs = []uint{1, 2}

	// 只选 models：models 相同即相等，byok_only/channels 差异被忽略。
	t.Run("models-only ignores byok and channels diff", func(t *testing.T) {
		tok := &Token{Models: `["a"]`, BYOKOnly: false} // byok 不同
		tok.AllowedChannelIDs = []uint{9}               // channels 不同
		if !TokenFieldsEqualForFields(tpl, tok, SyncFields{Models: true}) {
			t.Errorf("want equal (only models compared, models match)")
		}
	})

	// 选 byok_only：byok 不同 → 不相等。
	t.Run("byok selected and differs -> not equal", func(t *testing.T) {
		tok := &Token{Models: `["a"]`, BYOKOnly: false}
		tok.AllowedChannelIDs = []uint{1, 2}
		if TokenFieldsEqualForFields(tpl, tok, SyncFields{BYOKOnly: true}) {
			t.Errorf("want not equal (byok differs)")
		}
	})

	// 全选且全相同 → 相等。
	t.Run("all selected all equal -> equal", func(t *testing.T) {
		tok := &Token{Models: `["a"]`, BYOKOnly: true}
		tok.AllowedChannelIDs = []uint{2, 1}
		if !TokenFieldsEqualForFields(tpl, tok, SyncFields{Models: true, Channels: true, BYOKOnly: true}) {
			t.Errorf("want equal")
		}
	})
}
