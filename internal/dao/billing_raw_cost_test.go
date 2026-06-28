package dao

import (
	"testing"

	"github.com/VaalaCat/ai-gateway/internal/models"
)

func TestBatchUpsertChannelDaily_RawCost(t *testing.T) {
	ctx, db := setupAdminContext(t)
	m := NewAdminMutation(ctx).Billing()

	rows := []ChannelDailyRow{
		{Date: "2026-06-01", ChannelID: 5, OwnerType: "admin", RequestCount: 1, TotalCost: 0, RawCost: 100},
	}
	if err := m.BatchUpsertChannelDaily(rows); err != nil {
		t.Fatalf("first upsert: %v", err)
	}
	// 同 key 再 upsert → raw_cost 累加
	if err := m.BatchUpsertChannelDaily(rows); err != nil {
		t.Fatalf("second upsert: %v", err)
	}

	var got models.ChannelDailyBilling
	if err := db.Where("date = ? AND channel_id = ? AND private_channel_id = 0", "2026-06-01", 5).
		First(&got).Error; err != nil {
		t.Fatalf("read back: %v", err)
	}
	if got.RawCost != 200 {
		t.Fatalf("RawCost=%d want 200 (累加)", got.RawCost)
	}
	if got.TotalCost != 0 {
		t.Fatalf("TotalCost=%d want 0 (免费行折后为 0)", got.TotalCost)
	}
}
