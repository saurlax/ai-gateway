package billing

import (
	"testing"

	"github.com/VaalaCat/ai-gateway/internal/dao"
	"github.com/VaalaCat/ai-gateway/internal/models"
)

func TestMetricValue(t *testing.T) {
	// 免费渠道窗口用量:折后(billed)=0,折前(raw)=500
	u := dao.ChannelUsage{Calls: 7, BilledCost: 0, RawCost: 500}

	t.Run("success: cost + raw 口径 → RawCost(免费渠道也非 0)", func(t *testing.T) {
		r := models.LimitRule{Metric: models.LimitMetricCost, CostBasis: models.CostBasisRaw}
		if got := metricValue(r, u); got != 500 {
			t.Fatalf("got=%d want 500", got)
		}
	})
	t.Run("success: cost + 空口径 → RawCost(默认折前)", func(t *testing.T) {
		r := models.LimitRule{Metric: models.LimitMetricCost}
		if got := metricValue(r, u); got != 500 {
			t.Fatalf("got=%d want 500", got)
		}
	})
	t.Run("boundary: cost + billed 口径(免费渠道) → 0(永不触发)", func(t *testing.T) {
		r := models.LimitRule{Metric: models.LimitMetricCost, CostBasis: models.CostBasisBilled}
		if got := metricValue(r, u); got != 0 {
			t.Fatalf("got=%d want 0", got)
		}
	})
	t.Run("success: calls 指标 → Calls", func(t *testing.T) {
		r := models.LimitRule{Metric: models.LimitMetricCalls}
		if got := metricValue(r, u); got != 7 {
			t.Fatalf("got=%d want 7", got)
		}
	})
}
