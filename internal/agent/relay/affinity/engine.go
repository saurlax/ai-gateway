package affinity

import (
	"time"

	"github.com/VaalaCat/ai-gateway/internal/agent/relay/state"
	"github.com/VaalaCat/ai-gateway/internal/settings"
)

// ConfigReader 读当前同步配置快照。由 app.AgentCache（cache.Store）结构化满足，
// affinity 借此读 AffinityEnabled / AffinityTTLSec 而不 import app（避免成环）。
type ConfigReader interface {
	Settings() settings.AgentSettings
}

// AffinityStatus values reported in usage logs.
const (
	StatusHit      = "hit"
	StatusFallback = "fallback"
	StatusNone     = "none"
)

// Subject 是 Policy 决策的身份载体。ChannelEnabled/ChannelTTLSec 是命中渠道的每渠道覆盖,
// 零值(nil)=继承全局。解析在 Decide/ttl 内统一完成,调用方只负责把渠道覆盖原样传入。
type Subject struct {
	UserID         uint
	RealModel      string
	ChannelEnabled *bool
	ChannelTTLSec  *int
}

// Decision 拆 Record/Apply 两个布尔，为"只记录不应用 / 只应用不记录"等反向配置留口。
type Decision struct {
	Record bool // 是否记录粘性
	Apply  bool // 是否应用粘性（重排置顶）
}

// Policy 决定对某 Subject 是否记录/应用粘性。
type Policy interface {
	Decide(s Subject) Decision
}

// globalPolicy 看全局开关，渠道覆盖由 resolveEnabled 合并。
type globalPolicy struct{ cfg ConfigReader }

// resolveEnabled 把"渠道三态覆盖 + 全局开关"解析成最终是否启用。
// nil = 继承全局;非 nil = 强制覆盖(true 即使全局关也参与,false 即使全局开也不参与)。
func resolveEnabled(channelFlag *bool, globalOn bool) bool {
	if channelFlag != nil {
		return *channelFlag
	}
	return globalOn
}

func (p globalPolicy) Decide(s Subject) Decision {
	globalOn := p.cfg != nil && p.cfg.Settings().AffinityEnabled != 0
	on := resolveEnabled(s.ChannelEnabled, globalOn)
	return Decision{Record: on, Apply: on}
}

// Engine 是注入到 plan/publish/exec 的门面，组合 Store + Policy + ConfigReader。
type Engine struct {
	store  Store
	policy Policy
	cfg    ConfigReader
}

// New 用全局策略装配 Engine。cfg 提供开关与 TTL。
func New(cfg ConfigReader) *Engine {
	return &Engine{store: newTTLStore(), policy: globalPolicy{cfg: cfg}, cfg: cfg}
}

// Decide 转交策略。
func (e *Engine) Decide(s Subject) Decision { return e.policy.Decide(s) }

// Lookup 查粘性记录。
func (e *Engine) Lookup(k Key) (Entry, bool) { return e.store.Lookup(k) }

// Remember 记录/续期;解析后 TTL<=0 视为关闭,不写。ttlSecOverride 非 nil 时覆盖全局 TTL。
func (e *Engine) Remember(k Key, src state.ChannelSource, sourceID uint, ttlSecOverride *int) {
	ttl := e.ttl(ttlSecOverride)
	if ttl <= 0 {
		return
	}
	e.store.Remember(k, Entry{Source: src, SourceID: sourceID, ExpiresAt: time.Now().Add(ttl)})
}

// Forget 剔除（粘性 channel 硬失败时调用）。
func (e *Engine) Forget(k Key) { e.store.Forget(k) }

func (e *Engine) ttl(override *int) time.Duration {
	if e.cfg == nil {
		return 0
	}
	sec := e.cfg.Settings().AffinityTTLSec
	if override != nil {
		sec = *override
	}
	return time.Duration(sec) * time.Second
}
