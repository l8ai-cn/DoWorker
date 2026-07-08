package auth

import "testing"

func TestOriginChecker(t *testing.T) {
	oc := NewOriginChecker([]string{"https://app.example.com", "http://localhost:10000"})

	cases := []struct {
		origin string
		ok     bool
	}{
		{"https://app.example.com", true},
		{"http://localhost:10000", true},
		{"https://evil.com", false},
		{"", true}, // 非浏览器客户端（无 Origin 头）放行
	}
	for _, tc := range cases {
		if got := oc.Allowed(tc.origin); got != tc.ok {
			t.Fatalf("Allowed(%q)=%v want %v", tc.origin, got, tc.ok)
		}
	}

	// 空白名单：allowAll（保持向后兼容，配置未设置时不破坏现网）
	open := NewOriginChecker(nil)
	if !open.Allowed("https://anything.com") {
		t.Fatalf("empty allowlist should allow all")
	}
}
