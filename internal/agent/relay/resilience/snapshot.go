package resilience

import "sort"

// BreakerSnapshot 是某一时刻一个渠道熔断器的只读视图,供全局熔断看板跨 agent 聚合。
type BreakerSnapshot struct {
	Source      string  `json:"source"`       // "admin" | "private"
	ChannelID   uint    `json:"channel_id"`   // admin→Channel.ID / private→PrivateChannel.ID
	State       string  `json:"state"`        // "closed" | "open" | "half-open"
	RemainingMs int64   `json:"remaining_ms"` // open 时距 half-open 剩余;否则 0
	Failures    int     `json:"failures"`
	Successes   int     `json:"successes"`
	FailureRate float64 `json:"failure_rate"`
}

// SnapshotBreakers 快照当前所有 breaker,按 (Source, ChannelID) 确定性排序
// (避免 SyncMap 迭代序导致看板每次刷新行乱跳)。只读 failsafe breaker 状态/指标,
// 不解析渠道名——名字由上层 RPC handler 用 agent cache 补(解耦)。
func (r *Registry) SnapshotBreakers() []BreakerSnapshot {
	out := []BreakerSnapshot{}
	r.m.Range(func(k BreakerKey, e *breakerEntry) bool {
		cb := e.cb
		m := cb.Metrics()
		out = append(out, BreakerSnapshot{
			Source:      string(k.Source),
			ChannelID:   k.ID,
			State:       cb.State().String(),
			RemainingMs: cb.RemainingDelay().Milliseconds(),
			Failures:    int(m.Failures()),
			Successes:   int(m.Successes()),
			FailureRate: m.FailureRate(),
		})
		return true
	})
	sort.Slice(out, func(i, j int) bool {
		if out[i].Source != out[j].Source {
			return out[i].Source < out[j].Source
		}
		return out[i].ChannelID < out[j].ChannelID
	})
	return out
}
