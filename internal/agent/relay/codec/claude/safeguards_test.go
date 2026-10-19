package claude

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/VaalaCat/ai-gateway/internal/agent/relay/codec"
)

// The auto-mode classifier protocol pairs an `anthropic-beta` value
// (dangerous-tool-use-2026-09-03) with a top-level `safeguards` request field,
// and returns verdicts in `safeguard_results` (plus `stop_details` on a
// refusal). The header half already travels through ForwardClientHeaders; a
// same-protocol native relay must not drop either body half.

func decodeClaudeRequest(t *testing.T, body string) *codec.Request {
	t.Helper()
	r := httptest.NewRequest(http.MethodPost, "/v1/messages", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	req, err := (&ClaudeCodec{}).DecodeRequest(r)
	if err != nil {
		t.Fatalf("DecodeRequest: %v", err)
	}
	return req
}

func encodeClaudeRequest(t *testing.T, req *codec.Request) map[string]any {
	t.Helper()
	httpReq, err := (&ClaudeCodec{}).EncodeRequest(req, &codec.ChannelConfig{
		BaseURL: "https://api.anthropic.com",
		APIKey:  "test-key",
		Model:   req.Model,
	})
	if err != nil {
		t.Fatalf("EncodeRequest: %v", err)
	}
	body, err := io.ReadAll(httpReq.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	var out map[string]any
	if err := json.Unmarshal(body, &out); err != nil {
		t.Fatalf("unmarshal outbound body: %v", err)
	}
	return out
}

const safeguardsRequestBody = `{
	"model": "claude-opus-4-8",
	"max_tokens": 256,
	"stream": true,
	"messages": [{"role": "user", "content": "hi"}],
	"safeguards": [{
		"type": "dangerous_tool_use",
		"classifier_context": {"v": 1, "permission_mode": "auto"}
	}],
	"context_management": {"edits": []},
	"metadata": {"user_id": "u-1"}
}`

func TestDecodeRequest_CapturesUnknownTopLevelFields(t *testing.T) {
	req := decodeClaudeRequest(t, safeguardsRequestBody)

	if req.PassthroughFields["safeguards"] == nil {
		t.Fatal("safeguards was not captured as a passthrough field")
	}
	if req.PassthroughFields["context_management"] == nil {
		t.Fatal("context_management was not captured as a passthrough field")
	}
	if req.PassthroughFields["metadata"] == nil {
		t.Fatal("metadata was not captured as a passthrough field")
	}
	if _, ok := req.PassthroughFields["model"]; ok {
		t.Fatal("modelled field model must not be duplicated in PassthroughFields")
	}
	if _, ok := req.PassthroughFields["messages"]; ok {
		t.Fatal("modelled field messages must not be duplicated in PassthroughFields")
	}
}

func TestEncodeRequest_ForwardsSafeguardsOnSameProtocol(t *testing.T) {
	req := decodeClaudeRequest(t, safeguardsRequestBody)
	out := encodeClaudeRequest(t, req)

	raw, ok := out["safeguards"].([]any)
	if !ok || len(raw) != 1 {
		t.Fatalf("safeguards = %#v, want the original one-element array", out["safeguards"])
	}
	entry, ok := raw[0].(map[string]any)
	if !ok || entry["type"] != "dangerous_tool_use" {
		t.Fatalf("safeguards[0] = %#v, want type dangerous_tool_use", raw[0])
	}
	ctx, ok := entry["classifier_context"].(map[string]any)
	if !ok || ctx["permission_mode"] != "auto" {
		t.Fatalf("classifier_context was not forwarded unchanged: %#v", entry["classifier_context"])
	}
	if out["context_management"] == nil || out["metadata"] == nil {
		t.Fatalf("other unknown fields were dropped: %#v", out)
	}
}

func TestEncodeRequest_DropsUnknownFieldsOnCrossProtocol(t *testing.T) {
	req := decodeClaudeRequest(t, safeguardsRequestBody)
	// Cross-protocol relay: the upstream speaks a different schema, and the
	// paired anthropic-beta header is stripped for the same reason.
	req.InboundProtocol = codec.ProtocolOpenAIChat

	out := encodeClaudeRequest(t, req)
	if _, ok := out["safeguards"]; ok {
		t.Fatal("safeguards must not be forwarded when the inbound protocol is not Claude")
	}
	if _, ok := out["context_management"]; ok {
		t.Fatal("context_management must not be forwarded cross-protocol")
	}
}

func TestEncodeRequest_TypedFieldsWinOverPassthrough(t *testing.T) {
	req := decodeClaudeRequest(t, safeguardsRequestBody)
	// A channel-level model mapping rewrites Model; the raw inbound copy must
	// never resurrect the client's model.
	cfg := &codec.ChannelConfig{BaseURL: "https://api.anthropic.com", APIKey: "k", Model: "mapped-model"}
	httpReq, err := (&ClaudeCodec{}).EncodeRequest(req, cfg)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(httpReq.Body)
	var out map[string]any
	if err := json.Unmarshal(body, &out); err != nil {
		t.Fatal(err)
	}
	if out["model"] != "mapped-model" {
		t.Fatalf("model = %v, want mapped-model", out["model"])
	}
}

const safeguardsStream = `event: message_start
data: {"type":"message_start","message":{"id":"msg_1","type":"message","role":"assistant","model":"claude-opus-4-8","content":[],"stop_reason":null,"stop_sequence":null,"usage":{"input_tokens":10,"output_tokens":0}}}

event: content_block_start
data: {"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"ok"}}

event: content_block_stop
data: {"type":"content_block_stop","index":0}

event: message_delta
data: {"type":"message_delta","delta":{"stop_reason":"end_turn","stop_sequence":"END","safeguard_results":[{"type":"dangerous_tool_use","status":{"type":"available","tool_uses":{"toolu_01":{"type":"evaluated","outcome":"not_flagged"}}}}]},"usage":{"output_tokens":5},"context_management":{"applied_edits":[]}}

event: message_stop
data: {"type":"message_stop"}`

func TestStreamRelay_PreservesSafeguardResults(t *testing.T) {
	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(safeguardsStream)),
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
	}
	c := &ClaudeCodec{}
	events, err := c.DecodeResponse(resp, true)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	if err := c.EncodeResponse(events, w, true); err != nil {
		t.Fatalf("EncodeResponse: %v", err)
	}

	delta := findSSEData(t, w.Body.String(), "message_delta")
	inner, ok := delta["delta"].(map[string]any)
	if !ok {
		t.Fatalf("message_delta.delta = %#v, want object", delta["delta"])
	}
	results, ok := inner["safeguard_results"].([]any)
	if !ok || len(results) != 1 {
		t.Fatalf("safeguard_results = %#v, want the upstream array", inner["safeguard_results"])
	}
	result := results[0].(map[string]any)
	if result["type"] != "dangerous_tool_use" {
		t.Fatalf("result type = %v", result["type"])
	}
	status := result["status"].(map[string]any)
	verdicts := status["tool_uses"].(map[string]any)
	if verdicts["toolu_01"].(map[string]any)["outcome"] != "not_flagged" {
		t.Fatalf("verdict was altered: %#v", verdicts)
	}
	if inner["stop_sequence"] != "END" {
		t.Fatalf("stop_sequence from the upstream delta was dropped or altered: %#v", inner["stop_sequence"])
	}
	if inner["stop_reason"] != "end_turn" {
		t.Fatalf("stop_reason = %v, want end_turn", inner["stop_reason"])
	}
	if delta["context_management"] == nil {
		t.Fatal("top-level context_management of message_delta was dropped")
	}
	if _, ok := delta["usage"].(map[string]any); !ok {
		t.Fatalf("usage missing from message_delta: %#v", delta)
	}
}

func TestStreamRelay_PreservesStopDetailsOnRefusal(t *testing.T) {
	sse := `event: message_start
data: {"type":"message_start","message":{"id":"msg_1","type":"message","role":"assistant","content":[],"stop_reason":null,"usage":{"input_tokens":1,"output_tokens":0}}}

event: message_delta
data: {"type":"message_delta","delta":{"stop_reason":"refusal","stop_details":{"type":"refusal","category":null,"explanation":"blocked"}},"usage":{"output_tokens":1}}

event: message_stop
data: {"type":"message_stop"}`

	resp := &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(sse))}
	c := &ClaudeCodec{}
	events, err := c.DecodeResponse(resp, true)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	if err := c.EncodeResponse(events, w, true); err != nil {
		t.Fatal(err)
	}

	inner := findSSEData(t, w.Body.String(), "message_delta")["delta"].(map[string]any)
	if inner["stop_reason"] != "refusal" {
		t.Fatalf("stop_reason = %v, want refusal", inner["stop_reason"])
	}
	details, ok := inner["stop_details"].(map[string]any)
	if !ok || details["explanation"] != "blocked" {
		t.Fatalf("stop_details = %#v, want upstream object", inner["stop_details"])
	}
}

func TestNonStreamRelay_PreservesSafeguardResults(t *testing.T) {
	body := `{
		"id":"msg_1","type":"message","role":"assistant","model":"claude-opus-4-8",
		"content":[{"type":"text","text":"ok"}],
		"stop_reason":"tool_use",
		"safeguard_results":[{"type":"dangerous_tool_use","status":{"type":"available","tool_uses":{"toolu_01":{"type":"evaluated","outcome":"flagged","explanation":"destructive"}}}}],
		"usage":{"input_tokens":3,"output_tokens":2}
	}`
	resp := &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body))}
	c := &ClaudeCodec{}
	events, err := c.DecodeResponse(resp, false)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	if err := c.EncodeResponse(events, w, false); err != nil {
		t.Fatal(err)
	}

	var out map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	results, ok := out["safeguard_results"].([]any)
	if !ok || len(results) != 1 {
		t.Fatalf("safeguard_results = %#v, want the upstream array", out["safeguard_results"])
	}
	if out["model"] != "claude-opus-4-8" {
		t.Fatalf("upstream model was dropped: %#v", out)
	}
	if out["stop_reason"] != "tool_use" {
		t.Fatalf("stop_reason = %v, want tool_use", out["stop_reason"])
	}
}

func findSSEData(t *testing.T, body, event string) map[string]any {
	t.Helper()
	lines := strings.Split(body, "\n")
	for i, line := range lines {
		if strings.TrimSpace(line) != "event: "+event {
			continue
		}
		for _, next := range lines[i+1:] {
			if !strings.HasPrefix(next, "data:") {
				continue
			}
			var payload map[string]any
			if err := json.Unmarshal([]byte(strings.TrimSpace(strings.TrimPrefix(next, "data:"))), &payload); err != nil {
				t.Fatalf("parse %s data: %v", event, err)
			}
			return payload
		}
	}
	t.Fatalf("event %q not found in stream:\n%s", event, body)
	return nil
}
