package codec

import "encoding/json"

// CaptureUnknownFields returns the top-level members of body whose key is not
// listed in handled. Values are kept as raw JSON so a same-protocol encoder can
// re-emit them byte-for-byte.
//
// handled names every key the decoder either models in the IR (and re-emits
// from it) or deliberately drops; anything else is an unknown request field and
// must survive a same-protocol relay. Claude Code and the Claude API grow new
// body fields over releases (for example the auto-mode `safeguards` field), and
// a gateway pinned to an observed list silently strips the next capability.
func CaptureUnknownFields(body []byte, handled map[string]bool) map[string]json.RawMessage {
	var all map[string]json.RawMessage
	if err := json.Unmarshal(body, &all); err != nil {
		return nil
	}
	var out map[string]json.RawMessage
	for k, v := range all {
		if handled[k] {
			continue
		}
		if out == nil {
			out = make(map[string]json.RawMessage)
		}
		out[k] = v
	}
	return out
}

// MergeRawFields adds the members of extra that are not already present in the
// JSON object base and returns the re-marshalled object. Members of extra keep
// their exact JSON encoding. A nil/empty extra returns base unchanged.
//
// Keys already set by the encoder win: typed IR fields are authoritative, and a
// passthrough value must never overwrite them.
func MergeRawFields(base []byte, extra map[string]json.RawMessage) ([]byte, error) {
	if len(extra) == 0 {
		return base, nil
	}
	var out map[string]json.RawMessage
	if err := json.Unmarshal(base, &out); err != nil {
		return nil, err
	}
	for k, v := range extra {
		if _, exists := out[k]; exists {
			continue
		}
		out[k] = v
	}
	return json.Marshal(out)
}

// MergeAnyFields adds the members of extra that are not already present in the
// JSON object base and returns the re-marshalled object. Used on response paths,
// where values have already been decoded into the IR's any-typed containers.
// A nil/empty extra returns base unchanged.
func MergeAnyFields(base []byte, extra map[string]any) ([]byte, error) {
	if len(extra) == 0 {
		return base, nil
	}
	var out map[string]any
	if err := json.Unmarshal(base, &out); err != nil {
		return nil, err
	}
	for k, v := range extra {
		if _, exists := out[k]; exists {
			continue
		}
		out[k] = v
	}
	return json.Marshal(out)
}

// MergeRawObjectAddMissing decodes raw as a JSON object and copies its members
// into dst for keys dst does not already hold. Values keep their exact JSON
// encoding when dst is re-marshalled. Non-object or malformed raw is ignored,
// and so are null members: an explicit null carries no information and the
// gateway has always omitted empty optional members of its own deltas.
func MergeRawObjectAddMissing(dst map[string]any, raw json.RawMessage) {
	if len(raw) == 0 {
		return
	}
	var members map[string]json.RawMessage
	if err := json.Unmarshal(raw, &members); err != nil {
		return
	}
	for k, v := range members {
		if _, exists := dst[k]; exists {
			continue
		}
		if string(v) == "null" {
			continue
		}
		dst[k] = v
	}
}

// RawObjectMembers decodes raw as a JSON object and returns its members, or nil
// when raw is not a JSON object.
func RawObjectMembers(raw json.RawMessage) map[string]json.RawMessage {
	if len(raw) == 0 {
		return nil
	}
	var members map[string]json.RawMessage
	if err := json.Unmarshal(raw, &members); err != nil {
		return nil
	}
	return members
}
