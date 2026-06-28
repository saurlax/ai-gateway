package dao

import (
	"testing"

	"github.com/VaalaCat/ai-gateway/internal/models"
)

func TestBulkSyncFromTemplate_FieldSelection(t *testing.T) {
	ctx, db := setupAdminContext(t)
	m := NewAdminMutation(ctx).Token()

	tpl := &models.TokenTemplate{Name: "t", Models: `["a"]`, BYOKOnly: true, Status: 1, ExpiryDays: -1}
	if err := db.Create(tpl).Error; err != nil {
		t.Fatalf("seed tpl: %v", err)
	}
	tok := &models.Token{Name: "tk", TemplateID: &tpl.ID, Models: `["b"]`, BYOKOnly: false, Status: 1, ExpiredAt: -1, Key: "sk-1"}
	if err := db.Create(tok).Error; err != nil {
		t.Fatalf("seed tok: %v", err)
	}

	// 只同步 models：models 被覆盖，byok_only 保持不变。
	changed, total, err := m.BulkSyncFromTemplate(tpl.ID, tpl, models.SyncFields{Models: true})
	if err != nil {
		t.Fatalf("sync: %v", err)
	}
	if total != 1 || len(changed) != 1 {
		t.Fatalf("total=%d changed=%d, want 1/1", total, len(changed))
	}
	var got models.Token
	db.First(&got, tok.ID)
	if got.Models != `["a"]` {
		t.Errorf("Models = %q, want synced to template", got.Models)
	}
	if got.BYOKOnly != false {
		t.Errorf("BYOKOnly = %v, want unchanged false (not selected)", got.BYOKOnly)
	}

	// 再同步 byok_only：现在 byok_only 被覆盖为 true。
	changed2, _, err := m.BulkSyncFromTemplate(tpl.ID, tpl, models.SyncFields{BYOKOnly: true})
	if err != nil {
		t.Fatalf("sync2: %v", err)
	}
	if len(changed2) != 1 {
		t.Fatalf("changed2 = %d, want 1", len(changed2))
	}
	db.First(&got, tok.ID)
	if got.BYOKOnly != true {
		t.Errorf("BYOKOnly = %v, want true after selecting byok_only", got.BYOKOnly)
	}
}
