package native

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/VaalaCat/ai-gateway/internal/agent/relay/codec"
	"github.com/VaalaCat/ai-gateway/internal/agent/relay/state"
	"github.com/VaalaCat/ai-gateway/internal/consts"
	"github.com/VaalaCat/ai-gateway/internal/models"
)

// Auto mode's server-side classifier is a paired protocol: Claude Code sends
// `anthropic-beta: dangerous-tool-use-2026-09-03` and a top-level `safeguards`
// field, and reads the verdict back from `safeguard_results`. A Claude→Claude
// native relay must forward both halves unchanged in each direction; these
// tests exercise the real Relay chain (decode → dataflow → encode → dispatch),
// not just the codecs in isolation.

const autoModeBeta = "dangerous-tool-use-2026-09-03"

const autoModeRequestBody = `{
	"model": "claude-opus-4-8",
	"max_tokens": 256,
	"messages": [{"role": "user", "content": "rm -rf build"}],
	"safeguards": [{
		"type": "dangerous_tool_use",
		"classifier_context": {"session_id": "s-1", "permission_mode": "auto"}
	}]
}`

const autoModeResponseBody = `{
	"id": "msg_1",
	"type": "message",
	"role": "assistant",
	"model": "claude-opus-4-8",
	"content": [{"type": "tool_use", "id": "toolu_01", "name": "Bash", "input": {"command": "rm -rf build"}}],
	"stop_reason": "tool_use",
	"stop_sequence": null,
	"stop_details": {"type": "refusal", "category": null, "explanation": "not applied"},
	"safeguard_results": [{
		"type": "dangerous_tool_use",
		"status": {"type": "available", "tool_uses": {
			"toolu_01": {"type": "evaluated", "outcome": "flagged"}
		}}
	}],
	"usage": {"input_tokens": 12, "output_tokens": 4}
}`

func makeClaudeChannel(baseURL string) *models.Channel {
	return &models.Channel{
		ChannelCore: models.ChannelCore{
			ID: 1, Type: consts.ChannelTypeAnthropic, BaseURL: baseURL, Status: 1, Weight: 1,
		},
		Key:    "upstream-key",
		Models: "claude-opus-4-8",
	}
}

func TestBackend_AutoModeClassifier_PassthroughNonStream(t *testing.T) {
	var upstreamHeader http.Header
	var upstreamBody map[string]any
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upstreamHeader = r.Header.Clone()
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &upstreamBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(autoModeResponseBody))
	}))
	defer upstream.Close()

	rctx, w := newNativeTestCtx(t, []byte(autoModeRequestBody), codec.ProtocolClaude, false)
	// The client half of the pairing: the beta header that gates the body field.
	rctx.Context.Request.Header.Set("anthropic-beta", autoModeBeta)
	rctx.Context.Request.Header.Set("anthropic-version", "2023-06-01")

	got := (&Backend{}).Relay(rctx, state.Attempt{Channel: makeClaudeChannel(upstream.URL), RealModel: "claude-opus-4-8"})
	if got.Err != nil {
		t.Fatalf("Relay: %v", got.Err)
	}

	// Request direction: header and body field both reach the upstream untouched.
	if beta := upstreamHeader.Get("anthropic-beta"); beta != autoModeBeta {
		t.Errorf("upstream anthropic-beta = %q, want %q (must not be dropped or rewritten)", beta, autoModeBeta)
	}
	safeguards, ok := upstreamBody["safeguards"].([]any)
	if !ok || len(safeguards) != 1 {
		t.Fatalf("upstream safeguards = %#v, want the client's one-element array (body keys: %v)", upstreamBody["safeguards"], keysOf(upstreamBody))
	}
	entry := safeguards[0].(map[string]any)
	ctx := entry["classifier_context"].(map[string]any)
	if entry["type"] != "dangerous_tool_use" || ctx["permission_mode"] != "auto" {
		t.Errorf("upstream safeguards altered: %#v", entry)
	}

	// Response direction: verdict and refusal detail survive the relay.
	var clientBody map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &clientBody); err != nil {
		t.Fatalf("client body is not JSON: %v (body: %s)", err, w.Body.String())
	}
	results, ok := clientBody["safeguard_results"].([]any)
	if !ok || len(results) != 1 {
		t.Fatalf("client safeguard_results = %#v, want the upstream array", clientBody["safeguard_results"])
	}
	status := results[0].(map[string]any)["status"].(map[string]any)
	toolUses := status["tool_uses"].(map[string]any)
	if toolUses["toolu_01"].(map[string]any)["outcome"] != "flagged" {
		t.Errorf("classifier verdict altered: %#v", toolUses)
	}
	if _, ok := clientBody["stop_details"].(map[string]any); !ok {
		t.Errorf("client stop_details was dropped: %#v", clientBody)
	}
}

func TestBackend_AutoModeClassifier_PassthroughStream(t *testing.T) {
	upstreamSSE := strings.Join([]string{
		`event: message_start`,
		`data: {"type":"message_start","message":{"id":"msg_1","type":"message","role":"assistant","model":"claude-opus-4-8","content":[],"stop_reason":null,"stop_sequence":null,"usage":{"input_tokens":12,"output_tokens":0}}}`,
		``,
		`event: content_block_start`,
		`data: {"type":"content_block_start","index":0,"content_block":{"type":"tool_use","id":"toolu_01","name":"Bash","input":{}}}`,
		``,
		`event: content_block_delta`,
		`data: {"type":"content_block_delta","index":0,"delta":{"type":"input_json_delta","partial_json":"{\"command\":\"rm -rf build\"}"}}`,
		``,
		`event: content_block_stop`,
		`data: {"type":"content_block_stop","index":0}`,
		``,
		`event: message_delta`,
		`data: {"type":"message_delta","delta":{"stop_reason":"tool_use","stop_sequence":null,"safeguard_results":[{"type":"dangerous_tool_use","status":{"type":"available","tool_uses":{"toolu_01":{"type":"evaluated","outcome":"flagged"}}}}]},"usage":{"output_tokens":4}}`,
		``,
		`event: message_stop`,
		`data: {"type":"message_stop"}`,
		``,
	}, "\n")

	var upstreamHeader http.Header
	var upstreamBody map[string]any
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upstreamHeader = r.Header.Clone()
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &upstreamBody)
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(upstreamSSE))
	}))
	defer upstream.Close()

	body := strings.Replace(autoModeRequestBody, `"max_tokens": 256,`, `"max_tokens": 256, "stream": true,`, 1)
	rctx, w := newNativeTestCtx(t, []byte(body), codec.ProtocolClaude, true)
	rctx.Context.Request.Header.Set("anthropic-beta", autoModeBeta)

	got := (&Backend{}).Relay(rctx, state.Attempt{Channel: makeClaudeChannel(upstream.URL), RealModel: "claude-opus-4-8"})
	if got.Err != nil {
		t.Fatalf("Relay: %v", got.Err)
	}

	if beta := upstreamHeader.Get("anthropic-beta"); beta != autoModeBeta {
		t.Errorf("upstream anthropic-beta = %q, want %q", beta, autoModeBeta)
	}
	if _, ok := upstreamBody["safeguards"]; !ok {
		t.Fatalf("upstream safeguards dropped (body keys: %v)", keysOf(upstreamBody))
	}

	clientSSE := w.Body.String()
	idx := strings.Index(clientSSE, "event: message_delta")
	if idx < 0 {
		t.Fatalf("client stream has no message_delta event:\n%s", clientSSE)
	}
	var delta map[string]any
	for _, line := range strings.Split(clientSSE[idx:], "\n") {
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		if err := json.Unmarshal([]byte(strings.TrimSpace(strings.TrimPrefix(line, "data:"))), &delta); err != nil {
			t.Fatalf("parse message_delta data: %v", err)
		}
		break
	}
	inner, ok := delta["delta"].(map[string]any)
	if !ok {
		t.Fatalf("message_delta.delta = %#v", delta["delta"])
	}
	results, ok := inner["safeguard_results"].([]any)
	if !ok || len(results) != 1 {
		t.Fatalf("client message_delta.delta.safeguard_results = %#v, want the upstream array", inner["safeguard_results"])
	}
	status := results[0].(map[string]any)["status"].(map[string]any)
	if status["tool_uses"].(map[string]any)["toolu_01"].(map[string]any)["outcome"] != "flagged" {
		t.Errorf("streamed classifier verdict altered: %#v", status)
	}
}

func keysOf(m map[string]any) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
