package rpc

import (
	"runtime"

	"github.com/VaalaCat/ai-gateway/internal/agent/relay/inflight"
)

// HandleInflight 返回当前在途请求快照(供 master 节点管理远程查看)。
func HandleInflight(reg *inflight.Registry) (any, error) {
	if reg == nil {
		return []inflight.Snapshot{}, nil
	}
	return reg.Snapshot(), nil
}

// GoroutineDump 是 goroutine 栈快照载体。
type GoroutineDump struct {
	Count int    `json:"count"`
	Dump  string `json:"dump"`
}

// HandleGoroutines 抓取全部 goroutine 栈(诊断卡死阻塞点用)。
func HandleGoroutines() (GoroutineDump, error) {
	buf := make([]byte, 1<<20) // 1MB
	n := runtime.Stack(buf, true)
	return GoroutineDump{Count: runtime.NumGoroutine(), Dump: string(buf[:n])}, nil
}
