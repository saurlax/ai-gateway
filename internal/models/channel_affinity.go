package models

import "fmt"

// ChannelAffinity 是每 channel 对全局粘性设置(affinity_enabled / affinity_ttl_sec)的覆盖。
// 字段全为指针:nil = 继承全局;非 nil = 显式覆盖。挂在共享的 ChannelCore 上,
// 让 admin Channel 与 BYOK PrivateChannel 用同一处声明、同一份语义。
type ChannelAffinity struct {
	Enabled *bool `json:"enabled,omitempty"` // 覆盖 affinity_enabled
	TTLSec  *int  `json:"ttl_sec,omitempty"` // 覆盖 affinity_ttl_sec(秒)
}

// Validate 校验覆盖值边界,与全局 Settings 的 affinity_ttl_sec(0..86400) 一致。
// nil 字段表示不覆盖,跳过。Enabled 为布尔无需范围校验。
func (a ChannelAffinity) Validate() error {
	if a.TTLSec != nil && (*a.TTLSec < 0 || *a.TTLSec > 86400) {
		return fmt.Errorf("ttl_sec must be between 0 and 86400, got %d", *a.TTLSec)
	}
	return nil
}
