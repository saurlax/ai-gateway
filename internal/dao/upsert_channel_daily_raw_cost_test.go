package dao

import (
	"testing"

	"github.com/VaalaCat/ai-gateway/internal/models"
)

func ucdRawPtr(v int64) *int64 { return &v }

func TestUpsertChannelDaily_RawCost(t *testing.T) {
	ctx, db := setupAdminContext(t)
	m := NewAdminMutation(ctx).Billing()

	// 免费行:TotalCost=0,Raw* 桶保留原价
	log := &models.UsageLog{
		ChannelID: 5, PrivateChannelID: 0, OwnerType: "admin",
		ChannelName: "c", ChannelType: 1, Status: 1,
		CreatedAt:    1717200000, // 2026-06-01 UTC 区间内
		TotalCost:    0,
		RawInputCost: ucdRawPtr(70), RawOutputCost: ucdRawPtr(30),
	}
	if err := m.UpsertChannelDaily(log); err != nil {
		t.Fatalf("first: %v", err)
	}
	if err := m.UpsertChannelDaily(log); err != nil { // 同 key 累加
		t.Fatalf("second: %v", err)
	}

	var got models.ChannelDailyBilling
	if err := db.Where("channel_id = ? AND private_channel_id = 0", 5).First(&got).Error; err != nil {
		t.Fatalf("read back: %v", err)
	}
	if got.RawCost != 200 {
		t.Fatalf("RawCost=%d want 200 ((70+30)*2)", got.RawCost)
	}
	if got.TotalCost != 0 {
		t.Fatalf("TotalCost=%d want 0", got.TotalCost)
	}
}
