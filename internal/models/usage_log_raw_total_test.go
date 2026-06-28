package models

import "testing"

func ulRawPtr(v int64) *int64 { return &v }

func TestUsageLogRawTotal(t *testing.T) {
	t.Run("success: 四桶齐全 → 求和(忽略 TotalCost)", func(t *testing.T) {
		l := &UsageLog{
			RawInputCost: ulRawPtr(10), RawOutputCost: ulRawPtr(20),
			RawCacheReadCost: ulRawPtr(3), RawCacheWriteCost: ulRawPtr(4),
			TotalCost: 999,
		}
		if got := l.RawTotal(); got != 37 {
			t.Fatalf("RawTotal=%d want 37", got)
		}
	})

	t.Run("boundary: 四桶全 nil → 回退 TotalCost", func(t *testing.T) {
		l := &UsageLog{TotalCost: 50}
		if got := l.RawTotal(); got != 50 {
			t.Fatalf("RawTotal=%d want 50 (fallback)", got)
		}
	})

	t.Run("boundary: 部分 nil 当 0 计", func(t *testing.T) {
		l := &UsageLog{RawInputCost: ulRawPtr(15), TotalCost: 999}
		if got := l.RawTotal(); got != 15 {
			t.Fatalf("RawTotal=%d want 15 (nil 桶=0)", got)
		}
	})
}
