package cache

import (
	"testing"
	"time"
)

func TestResolveNegativeTTL(t *testing.T) {
	cases := []struct {
		in   int
		want time.Duration
	}{
		{0, 600 * time.Second},   // 0 = 默认
		{-1, 0},                  // 负 = 禁用
		{120, 120 * time.Second}, // 正 = 原值
	}
	for _, c := range cases {
		if got := resolveNegativeTTL(c.in); got != c.want {
			t.Fatalf("resolveNegativeTTL(%d) = %v, want %v", c.in, got, c.want)
		}
	}
}
