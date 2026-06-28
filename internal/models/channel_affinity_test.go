package models

import "testing"

func ptrInt(v int) *int { return &v }

func TestChannelAffinityValidate(t *testing.T) {
	cases := []struct {
		name    string
		a       ChannelAffinity
		wantErr bool
	}{
		{"nil ttl ok", ChannelAffinity{}, false},
		{"ttl in range", ChannelAffinity{TTLSec: ptrInt(600)}, false},
		{"ttl zero ok", ChannelAffinity{TTLSec: ptrInt(0)}, false},
		{"ttl max ok", ChannelAffinity{TTLSec: ptrInt(86400)}, false},
		{"ttl negative", ChannelAffinity{TTLSec: ptrInt(-1)}, true},
		{"ttl too large", ChannelAffinity{TTLSec: ptrInt(86401)}, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := c.a.Validate()
			if (err != nil) != c.wantErr {
				t.Fatalf("Validate() err=%v wantErr=%v", err, c.wantErr)
			}
		})
	}
}
