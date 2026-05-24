package agent

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/VaalaCat/ai-gateway/internal/consts"
	"github.com/VaalaCat/ai-gateway/internal/dao"
	"github.com/VaalaCat/ai-gateway/internal/master/api"
	"github.com/VaalaCat/ai-gateway/internal/pkg/app"
)

// AgentIDQuery uses ?id= query parameter (no dynamic route segment per project convention).
type AgentIDQuery struct {
	ID string `form:"id" binding:"required"`
}

// GetInflight fetches the in-flight request snapshot from a remote agent.
func (h *Handler) GetInflight(c *app.Context, req AgentIDQuery) (json.RawMessage, error) {
	if h.HubCall == nil {
		return nil, api.InternalError("hub not available", nil)
	}
	id, _ := strconv.ParseUint(req.ID, 10, 64)
	q := dao.NewAdminQuery(dao.NewContext(c.App))
	ag, err := q.Agent().GetByID(uint(id))
	if err != nil {
		return nil, api.NotFoundError("agent not found")
	}
	res, err := h.HubCall(ag.AgentID, consts.RPCAgentInflight, nil, 10*time.Second)
	if err != nil {
		return nil, api.BadRequestError(fmt.Sprintf("inflight query failed: %v", err), nil)
	}
	return res, nil
}

// GetGoroutines fetches a goroutine dump from a remote agent (admin only).
func (h *Handler) GetGoroutines(c *app.Context, req AgentIDQuery) (json.RawMessage, error) {
	if h.HubCall == nil {
		return nil, api.InternalError("hub not available", nil)
	}
	id, _ := strconv.ParseUint(req.ID, 10, 64)
	q := dao.NewAdminQuery(dao.NewContext(c.App))
	ag, err := q.Agent().GetByID(uint(id))
	if err != nil {
		return nil, api.NotFoundError("agent not found")
	}
	res, err := h.HubCall(ag.AgentID, consts.RPCAgentGoroutines, nil, 15*time.Second)
	if err != nil {
		return nil, api.BadRequestError(fmt.Sprintf("goroutine dump failed: %v", err), nil)
	}
	return res, nil
}
