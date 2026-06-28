package models

import "testing"

func TestChannelLimitValidate_CostBasis(t *testing.T) {
	mk := func(basis string) ChannelLimit {
		return ChannelLimit{Rules: []LimitRule{{
			Metric: LimitMetricCost, Window: LimitWindowMonthly, Threshold: 100, CostBasis: basis,
		}}}
	}

	t.Run("success: 空口径合法(=折前)", func(t *testing.T) {
		if err := mk("").Validate(); err != nil {
			t.Fatalf("empty basis err=%v", err)
		}
	})

	t.Run("success: raw / billed 合法", func(t *testing.T) {
		if err := mk(CostBasisRaw).Validate(); err != nil {
			t.Fatalf("raw err=%v", err)
		}
		if err := mk(CostBasisBilled).Validate(); err != nil {
			t.Fatalf("billed err=%v", err)
		}
	})

	t.Run("failure: 非法口径被拒", func(t *testing.T) {
		if err := mk("net").Validate(); err == nil {
			t.Fatalf("want error for cost_basis=net")
		}
	})
}
