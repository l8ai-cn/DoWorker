# Gateway 模块 实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将 relay 演进为统一 Gateway：内网 Runner 主动出站隧道 + 所有 worker 的 HTTP/WS/媒体（HTML、图片、视频）经 Gateway 提供给前端。

**Architecture:** 在现有 `relay/` 进程内新增 HTTP 数据面：每个 Runner 一条 `/runner/tunnel` WebSocket，内部用 `[type][stream_id][payload]` 二进制帧多路复用任意 HTTP/WS 请求；Gateway 对外暴露 `/preview/{podKey}/*`，鉴权靠 backend 签发的 `token_type=preview` JWT（内含 `runner_id + target`），credit 窗口做流控。终端/ACP 现状不动。

**Tech Stack:** Go 1.x（gorilla/websocket、golang-jwt/jwt/v5、viper、gin、GORM）、Protobuf（Connect/gRPC）、React（web-user）、Bazel 构建、Traefik/Ingress 反代。

**设计文档:** `docs/superpowers/specs/2026-07-08-gateway-module-design.md`

**默认决策:** 不改模块名（在 `relay/` 内演进）；preview 走 path 前缀 `/preview/{podKey}/`。

---

## 构建与测试约定

- 构建：`bazel build //relay/... //backend/... //runner/...`
- Go 单测：`bazel test //relay/internal/<pkg>:<target>`；无 Bazel 目标时用 `go test ./relay/internal/<pkg>/...`（在 `doworker/` 下）。
- 新增 Go 文件后需同步更新对应目录 `BUILD.bazel` 的 `srcs`/`deps`（本仓库用 gazelle 风格；每个 Task 的 commit 前跑 `bazel run //:gazelle` 若存在，否则手动补）。
- Proto 改动后：`bazel run //proto:generate`（或仓库既有的 proto 生成命令），提交生成物。
- 每个 Task 结束提交一次；提交信息用 `feat(gateway): ...` / `test(gateway): ...` / `chore(gateway): ...`。

---

## 文件结构总览

新增/修改（详见设计文档第 3 节）：

- `relay/internal/protocol/tunnelframe/frame.go` — 帧编解码（新）
- `relay/internal/tunnel/{registry,tunnel,stream,limits}.go` — 隧道与流控（新）
- `relay/internal/proxy/{http,websocket,headers}.go` — HTTP/WS 代理（新）
- `relay/internal/auth/{token.go,origin.go}` — token_type + Origin（改/新）
- `relay/internal/server/{server,handler,handler_tunnel,handler_preview,handler_preview_session}.go`（改/新）
- `relay/internal/config/config.go` — 隧道/代理/origin 配置（改）
- `proto/runner/v1/runner.proto` — `ConnectTunnelCommand connect_tunnel = 14`（改）
- `runner/internal/tunnel/{client,dispatcher,local_http}.go`、`runner/internal/runner/message_handler_tunnel.go`（新）
- `backend/internal/service/relay/{token.go,preview.go}`、`backend/internal/api/rest/v1/pod_preview.go`、`backend/internal/domain/agentpod/pod.go`、迁移（改/新）
- `clients/web-user/src/hooks/usePodPreview.ts`、`clients/web-user/src/components/PreviewPanel.tsx`（新）
- `deploy/dev/traefik/dynamic/http.yml`、`deploy/kubernetes/cluster-oilan/40-ingress.yaml`（改）

---

# Phase 1 — 加固现状（与现网完全兼容，可独立发布）

## Task 1.1: token_type claim（Gateway 侧校验器）

**Files:**
- Modify: `relay/internal/auth/token.go`
- Test: `relay/internal/auth/token_test.go`

- [ ] **Step 1: Write the failing test**

在 `relay/internal/auth/token_test.go` 追加：

```go
func TestRelayClaims_TokenType(t *testing.T) {
	secret := "s3cret"
	// 旧 token（无 token_type）：runner=UserID 0，browser=UserID!=0 仍成立
	legacyRunner, err := GenerateToken(secret, "iss", "pod1", 7, 0, 3, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	v := NewTokenValidator(secret, "iss")
	c, err := v.ValidateToken(legacyRunner)
	if err != nil {
		t.Fatal(err)
	}
	if !c.IsRunnerToken() {
		t.Fatalf("legacy runner token should be runner")
	}
	if c.ResolvedType() != TokenTypeRunner {
		t.Fatalf("legacy runner should resolve to runner, got %q", c.ResolvedType())
	}

	// 新 token：显式 tunnel 类型
	tunnel, err := GenerateTypedToken(secret, "iss", TokenTypeTunnel, "", 7, 0, 3, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	tc, err := v.ValidateToken(tunnel)
	if err != nil {
		t.Fatal(err)
	}
	if tc.ResolvedType() != TokenTypeTunnel {
		t.Fatalf("expected tunnel, got %q", tc.ResolvedType())
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./relay/internal/auth/ -run TestRelayClaims_TokenType -v`
Expected: 编译失败（`TokenType`、`GenerateTypedToken`、`ResolvedType`、常量未定义）

- [ ] **Step 3: Write minimal implementation**

在 `relay/internal/auth/token.go` 修改/新增：

```go
type TokenType string

const (
	TokenTypeRunner  TokenType = "runner"
	TokenTypeBrowser TokenType = "browser"
	TokenTypeTunnel  TokenType = "tunnel"
	TokenTypePreview TokenType = "preview"
)

// RelayClaims: 新增字段（其余保持不变）
type RelayClaims struct {
	PodKey        string    `json:"pod_key"`
	RunnerID      int64     `json:"runner_id"`
	UserID        int64     `json:"user_id"`
	OrgID         int64     `json:"org_id"`
	TokenType     TokenType `json:"token_type,omitempty"`
	PreviewTarget string    `json:"preview_target,omitempty"` // e.g. 127.0.0.1:3000
	jwt.RegisteredClaims
}

// ResolvedType 回退：无显式 token_type 时按旧规则（UserID==0 → runner）。
func (c *RelayClaims) ResolvedType() TokenType {
	if c.TokenType != "" {
		return c.TokenType
	}
	if c.UserID == 0 {
		return TokenTypeRunner
	}
	return TokenTypeBrowser
}

// GenerateTypedToken 供 backend/测试使用（保留原 GenerateToken 不变）。
func GenerateTypedToken(secret, issuer string, tokenType TokenType, previewTarget string, runnerID, userID, orgID int64, expiry time.Duration) (string, error) {
	now := time.Now()
	claims := &RelayClaims{
		RunnerID:      runnerID,
		UserID:        userID,
		OrgID:         orgID,
		TokenType:     tokenType,
		PreviewTarget: previewTarget,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    issuer,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./relay/internal/auth/ -run TestRelayClaims_TokenType -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add relay/internal/auth/token.go relay/internal/auth/token_test.go
git commit -m "feat(gateway): add token_type claim with legacy fallback"
```

## Task 1.2: Origin 白名单

**Files:**
- Create: `relay/internal/auth/origin.go`
- Test: `relay/internal/auth/origin_test.go`

- [ ] **Step 1: Write the failing test**

```go
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
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./relay/internal/auth/ -run TestOriginChecker -v`
Expected: 编译失败（`NewOriginChecker` 未定义）

- [ ] **Step 3: Write minimal implementation**

```go
package auth

import "strings"

type OriginChecker struct {
	allowed  map[string]struct{}
	allowAll bool
}

func NewOriginChecker(origins []string) *OriginChecker {
	oc := &OriginChecker{allowed: make(map[string]struct{})}
	n := 0
	for _, o := range origins {
		o = strings.TrimSpace(strings.ToLower(o))
		if o == "" {
			continue
		}
		oc.allowed[o] = struct{}{}
		n++
	}
	if n == 0 {
		oc.allowAll = true
	}
	return oc
}

// Allowed: 空 origin（无浏览器 Origin 头，如 runner）放行；否则须命中白名单。
func (oc *OriginChecker) Allowed(origin string) bool {
	if origin == "" || oc.allowAll {
		return true
	}
	_, ok := oc.allowed[strings.ToLower(strings.TrimSpace(origin))]
	return ok
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./relay/internal/auth/ -run TestOriginChecker -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add relay/internal/auth/origin.go relay/internal/auth/origin_test.go
git commit -m "feat(gateway): add origin allowlist checker"
```

## Task 1.3: 配置增加 origin 与隧道段

**Files:**
- Modify: `relay/internal/config/config.go`
- Test: `relay/internal/config/config_test.go`

- [ ] **Step 1: Write the failing test**

在 `config_test.go` 追加：

```go
func TestLoad_TunnelAndOriginDefaults(t *testing.T) {
	t.Setenv("JWT_SECRET", "x")
	t.Setenv("INTERNAL_API_SECRET", "y")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.Tunnel.Enabled {
		t.Fatalf("tunnel should default enabled")
	}
	if cfg.Tunnel.MaxStreamsPerPod != 32 {
		t.Fatalf("MaxStreamsPerPod default=32, got %d", cfg.Tunnel.MaxStreamsPerPod)
	}
	if cfg.Tunnel.StreamWindowBytes != 1<<20 {
		t.Fatalf("StreamWindowBytes default=1MiB, got %d", cfg.Tunnel.StreamWindowBytes)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./relay/internal/config/ -run TestLoad_TunnelAndOriginDefaults -v`
Expected: 编译失败（`cfg.Tunnel` 未定义）

- [ ] **Step 3: Write minimal implementation**

在 `config.go` 的 `Config` 结构追加字段与默认值：

```go
// Config 追加：
	Tunnel        TunnelConfig `mapstructure:"tunnel"`
	AllowedOrigins string      `mapstructure:"allowed_origins"` // 逗号分隔

// 新增类型：
type TunnelConfig struct {
	Enabled            bool          `mapstructure:"enabled"`
	MaxStreamsPerPod   int           `mapstructure:"max_streams_per_pod"`
	QueuePerPod        int           `mapstructure:"queue_per_pod"`
	QueueTimeout       time.Duration `mapstructure:"queue_timeout"`
	ReconnectGrace     time.Duration `mapstructure:"reconnect_grace"`
	StreamTimeout      time.Duration `mapstructure:"stream_timeout"`
	StreamWindowBytes  int           `mapstructure:"stream_window_bytes"`
}
```

在 `Load()` 的默认值区追加：

```go
	v.SetDefault("tunnel.enabled", true)
	v.SetDefault("tunnel.max_streams_per_pod", 32)
	v.SetDefault("tunnel.queue_per_pod", 16)
	v.SetDefault("tunnel.queue_timeout", 5*time.Second)
	v.SetDefault("tunnel.reconnect_grace", 5*time.Second)
	v.SetDefault("tunnel.stream_timeout", 300*time.Second)
	v.SetDefault("tunnel.stream_window_bytes", 1<<20)
```

在 `envMappings` map 追加：

```go
		"ALLOWED_ORIGINS":            "allowed_origins",
		"TUNNEL_ENABLED":             "tunnel.enabled",
		"TUNNEL_MAX_STREAMS_PER_POD": "tunnel.max_streams_per_pod",
		"TUNNEL_QUEUE_PER_POD":       "tunnel.queue_per_pod",
		"TUNNEL_QUEUE_TIMEOUT":       "tunnel.queue_timeout",
		"TUNNEL_RECONNECT_GRACE":     "tunnel.reconnect_grace",
		"TUNNEL_STREAM_TIMEOUT":      "tunnel.stream_timeout",
		"TUNNEL_STREAM_WINDOW":       "tunnel.stream_window_bytes",
```

新增辅助方法（供 server 使用）：

```go
// AllowedOriginList 拆分逗号分隔配置；PRIMARY_DOMAIN 存在时自动加入其 http/https 形式。
func (c *Config) AllowedOriginList() []string {
	var out []string
	for _, o := range strings.Split(c.AllowedOrigins, ",") {
		if s := strings.TrimSpace(o); s != "" {
			out = append(out, s)
		}
	}
	if c.PrimaryDomain != "" {
		out = append(out, "https://"+c.PrimaryDomain, "http://"+c.PrimaryDomain)
	}
	return out
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./relay/internal/config/ -run TestLoad_TunnelAndOriginDefaults -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add relay/internal/config/config.go relay/internal/config/config_test.go
git commit -m "feat(gateway): add tunnel and origin config"
```

## Task 1.4: 终端/ACP handler 接入 Origin 校验 + token_type

**Files:**
- Modify: `relay/internal/server/handler.go`
- Modify: `relay/internal/server/server.go`
- Test: `relay/internal/server/handler_test.go`

- [ ] **Step 1: Write the failing test**

在 `handler_test.go` 追加（复用现有测试的 upgrader/token 构造工具；若无则参考现有用例构造）：

```go
func TestHandler_RejectsDisallowedOrigin(t *testing.T) {
	h := NewHandlerWithOrigin(newTestChannelManager(t), newTestValidator(t),
		auth.NewOriginChecker([]string{"https://good.example"}))

	srv := httptest.NewServer(http.HandlerFunc(h.HandleBrowserWS))
	defer srv.Close()

	url := "ws" + strings.TrimPrefix(srv.URL, "http") + "/?token=" + validBrowserToken(t)
	hdr := http.Header{"Origin": {"https://evil.example"}}
	_, resp, err := websocket.DefaultDialer.Dial(url, hdr)
	if err == nil {
		t.Fatal("expected dial to fail on bad origin")
	}
	if resp == nil || resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403, got %v", resp)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./relay/internal/server/ -run TestHandler_RejectsDisallowedOrigin -v`
Expected: 编译失败（`NewHandlerWithOrigin` 未定义）

- [ ] **Step 3: Write minimal implementation**

在 `handler.go`：为 `Handler` 增加 `originChecker *auth.OriginChecker` 字段；新增构造器 `NewHandlerWithOrigin(cm, tv, oc)`，令 `NewHandler` 调用它并传入 `auth.NewOriginChecker(nil)`（allowAll，兼容旧测试）。将全局 `upgrader` 的 `CheckOrigin` 改为方法内闭包：

```go
func (h *Handler) upgrade(w http.ResponseWriter, r *http.Request) (*websocket.Conn, bool) {
	if !h.originChecker.Allowed(r.Header.Get("Origin")) {
		http.Error(w, "forbidden origin", http.StatusForbidden)
		return nil, false
	}
	up := websocket.Upgrader{
		ReadBufferSize:  1024 * 64,
		WriteBufferSize: 1024 * 64,
		CheckOrigin:     func(*http.Request) bool { return true }, // 已在上面校验
	}
	conn, err := up.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("upgrade failed", "error", err)
		return nil, false
	}
	return conn, true
}
```

在 `HandleRunnerWS`/`HandleBrowserWS` 用 `h.upgrade(...)` 替换 `upgrader.Upgrade(...)`；并把原 `claims.UserID != 0` / `== 0` 判断替换为 `claims.ResolvedType() != auth.TokenTypeRunner` / `!= auth.TokenTypeBrowser`。在 `server.go` 的 `New` 中改为 `NewHandlerWithOrigin(..., auth.NewOriginChecker(cfg.AllowedOriginList()))`。

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./relay/internal/server/ -v`
Expected: PASS（含既有用例）

- [ ] **Step 5: Commit**

```bash
git add relay/internal/server/handler.go relay/internal/server/server.go relay/internal/server/handler_test.go
git commit -m "feat(gateway): enforce origin allowlist and token_type on terminal ws"
```

## Task 1.5: backend preview 下发失败不发 token（先修 A3 的通用工具）

**Files:**
- Modify: `backend/internal/service/relay/token.go`
- Test: `backend/internal/service/relay/token_test.go`

- [ ] **Step 1: Write the failing test**

```go
func TestGenerateTypedToken_Preview(t *testing.T) {
	g := NewTokenGenerator("secret", "iss")
	tok, err := g.GenerateTypedToken("pod1", 7, 42, 3, "preview", "127.0.0.1:3000", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if tok == "" {
		t.Fatal("empty token")
	}
	// 空 preview target 应报错（preview 必须绑定 target）
	if _, err := g.GenerateTypedToken("pod1", 7, 42, 3, "preview", "", time.Hour); err == nil {
		t.Fatal("preview token without target must error")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./backend/internal/service/relay/ -run TestGenerateTypedToken_Preview -v`
Expected: 编译失败（`GenerateTypedToken` 未定义）

- [ ] **Step 3: Write minimal implementation**

在 `backend/internal/service/relay/token.go` 的 `TokenClaims` 增加 `TokenType string` 与 `PreviewTarget string`（json `token_type`/`preview_target`），并新增：

```go
func (g *TokenGenerator) GenerateTypedToken(podKey string, runnerID, userID, orgID int64, tokenType, previewTarget string, expiry time.Duration) (string, error) {
	if tokenType == "preview" && previewTarget == "" {
		return "", fmt.Errorf("preview token requires target")
	}
	if expiry <= 0 {
		return "", fmt.Errorf("expiry must be positive")
	}
	now := time.Now()
	claims := &TokenClaims{
		PodKey:        podKey,
		RunnerID:      runnerID,
		UserID:        userID,
		OrgID:         orgID,
		TokenType:     tokenType,
		PreviewTarget: previewTarget,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    g.issuer,
			Subject:   podKey,
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(g.secretKey)
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./backend/internal/service/relay/ -run TestGenerateTypedToken_Preview -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add backend/internal/service/relay/token.go backend/internal/service/relay/token_test.go
git commit -m "feat(gateway): backend typed token generator with preview target"
```

---

# Phase 2 — 隧道骨架（帧协议 + 注册表 + runner 客户端 + connect_tunnel）

## Task 2.1: 隧道帧编解码

**Files:**
- Create: `relay/internal/protocol/tunnelframe/frame.go`
- Test: `relay/internal/protocol/tunnelframe/frame_test.go`

- [ ] **Step 1: Write the failing test**

```go
package tunnelframe

import (
	"bytes"
	"testing"
)

func TestEncodeDecode(t *testing.T) {
	f := Frame{Type: TypeReqBody, StreamID: 0x01020304, Payload: []byte("hello")}
	raw := Encode(f)
	if len(raw) != 1+4+5 {
		t.Fatalf("bad length %d", len(raw))
	}
	got, err := Decode(raw)
	if err != nil {
		t.Fatal(err)
	}
	if got.Type != TypeReqBody || got.StreamID != 0x01020304 || !bytes.Equal(got.Payload, []byte("hello")) {
		t.Fatalf("roundtrip mismatch: %+v", got)
	}
}

func TestDecode_TooShort(t *testing.T) {
	if _, err := Decode([]byte{0x10, 0x00}); err == nil {
		t.Fatal("expected error on short frame")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./relay/internal/protocol/tunnelframe/ -v`
Expected: 编译失败（包不存在）

- [ ] **Step 3: Write minimal implementation**

```go
package tunnelframe

import (
	"encoding/binary"
	"errors"
)

type FrameType byte

const (
	TypeHello       FrameType = 0x01
	TypePing        FrameType = 0x02
	TypePong        FrameType = 0x03
	TypeReqStart    FrameType = 0x10
	TypeReqBody     FrameType = 0x11
	TypeReqEnd      FrameType = 0x12
	TypeStreamCancel FrameType = 0x13
	TypeRespStart   FrameType = 0x20
	TypeRespBody    FrameType = 0x21
	TypeRespEnd     FrameType = 0x22
	TypeRespError   FrameType = 0x23
	TypeWSData      FrameType = 0x30
	TypeWSClose     FrameType = 0x31
	TypeCredit      FrameType = 0x40
)

const HeaderSize = 5 // 1B type + 4B stream_id
const MaxChunk = 256 << 10

var ErrShortFrame = errors.New("tunnelframe: short frame")

type Frame struct {
	Type     FrameType
	StreamID uint32
	Payload  []byte
}

func Encode(f Frame) []byte {
	buf := make([]byte, HeaderSize+len(f.Payload))
	buf[0] = byte(f.Type)
	binary.BigEndian.PutUint32(buf[1:5], f.StreamID)
	copy(buf[5:], f.Payload)
	return buf
}

func Decode(raw []byte) (Frame, error) {
	if len(raw) < HeaderSize {
		return Frame{}, ErrShortFrame
	}
	return Frame{
		Type:     FrameType(raw[0]),
		StreamID: binary.BigEndian.Uint32(raw[1:5]),
		Payload:  raw[HeaderSize:],
	}, nil
}
```

新增 JSON 载荷类型（同文件或 `payloads.go`）：`HelloPayload{RunnerID, OrgID, Version string, Capabilities []string}`、`ReqStartPayload{Method, Path, RawQuery string, Header map[string][]string, PodKey, Target string, ContentLength int64, IsWebSocket bool}`、`RespStartPayload{Status int, Header map[string][]string}`、`RespErrorPayload{Code, Message string}`、`WSClosePayload{Code int, Reason string}`、`StreamCancelPayload{Code int, Reason string}`。用 `encoding/json` 序列化到对应帧的 `Payload`。

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./relay/internal/protocol/tunnelframe/ -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add relay/internal/protocol/tunnelframe/
git commit -m "feat(gateway): tunnel frame codec and payload types"
```

## Task 2.2: Stream 与 credit 流控

**Files:**
- Create: `relay/internal/tunnel/stream.go`
- Test: `relay/internal/tunnel/stream_test.go`

- [ ] **Step 1: Write the failing test**

```go
package tunnel

import (
	"context"
	"testing"
	"time"
)

func TestCreditWindow_BlocksAndResumes(t *testing.T) {
	w := newCreditWindow(4)
	if err := w.acquire(context.Background(), 4); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- w.acquire(context.Background(), 3) }() // 窗口耗尽，阻塞
	select {
	case <-done:
		t.Fatal("acquire should block when window empty")
	case <-time.After(50 * time.Millisecond):
	}
	w.add(5) // 补窗
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("acquire should resume after credit added")
	}
}

func TestCreditWindow_CtxCancel(t *testing.T) {
	w := newCreditWindow(0)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := w.acquire(ctx, 1); err == nil {
		t.Fatal("expected ctx error")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./relay/internal/tunnel/ -run TestCreditWindow -v`
Expected: 编译失败（`newCreditWindow` 未定义）

- [ ] **Step 3: Write minimal implementation**

在 `stream.go` 实现 credit 窗口（条件变量）：

```go
package tunnel

import (
	"context"
	"sync"
)

type creditWindow struct {
	mu     sync.Mutex
	cond   *sync.Cond
	avail  int
	closed bool
}

func newCreditWindow(initial int) *creditWindow {
	w := &creditWindow{avail: initial}
	w.cond = sync.NewCond(&w.mu)
	return w
}

func (w *creditWindow) acquire(ctx context.Context, n int) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	for w.avail < n && !w.closed {
		if err := ctx.Err(); err != nil {
			return err
		}
		// 用 goroutine 桥接 ctx 取消到 cond
		done := make(chan struct{})
		go func() {
			select {
			case <-ctx.Done():
				w.cond.Broadcast()
			case <-done:
			}
		}()
		w.cond.Wait()
		close(done)
		if err := ctx.Err(); err != nil {
			return err
		}
	}
	if w.closed {
		return context.Canceled
	}
	w.avail -= n
	return nil
}

func (w *creditWindow) add(n int) {
	w.mu.Lock()
	w.avail += n
	w.mu.Unlock()
	w.cond.Broadcast()
}

func (w *creditWindow) close() {
	w.mu.Lock()
	w.closed = true
	w.mu.Unlock()
	w.cond.Broadcast()
}
```

同文件定义 `Stream` 结构（`ID uint32`、`sendWin/recvWin *creditWindow`、`respCh chan tunnelframe.Frame`、`cancel func()`），供 Task 2.3/3.x 使用。

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./relay/internal/tunnel/ -run TestCreditWindow -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add relay/internal/tunnel/stream.go relay/internal/tunnel/stream_test.go
git commit -m "feat(gateway): credit-based flow control window"
```

## Task 2.3: 隧道连接（读写循环 + stream 表 + 心跳）

**Files:**
- Create: `relay/internal/tunnel/tunnel.go`
- Test: `relay/internal/tunnel/tunnel_test.go`

- [ ] **Step 1: Write the failing test**

```go
func TestTunnel_OpenStreamRoutesResponse(t *testing.T) {
	c1, c2 := net.Pipe() // 用 websocket over pipe 或 mock conn
	_ = c2
	tun := newTunnelForTest(c1) // helper 包装成带 writeMu 的连接

	st := tun.OpenStream()
	// 模拟对端回一个 RESP_START 到该 stream
	tun.dispatch(tunnelframe.Frame{Type: tunnelframe.TypeRespStart, StreamID: st.ID,
		Payload: mustJSON(tunnelframe.RespStartPayload{Status: 200})})

	select {
	case f := <-st.respCh:
		if f.Type != tunnelframe.TypeRespStart {
			t.Fatalf("unexpected %v", f.Type)
		}
	case <-time.After(time.Second):
		t.Fatal("response not routed to stream")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./relay/internal/tunnel/ -run TestTunnel_OpenStreamRoutesResponse -v`
Expected: 编译失败（`newTunnelForTest`/`OpenStream`/`dispatch` 未定义）

- [ ] **Step 3: Write minimal implementation**

在 `tunnel.go` 实现 `Tunnel`：持有 `conn *websocket.Conn`（含 `writeMu`）、`streams map[uint32]*Stream`（`sync.Mutex`）、`nextID uint32`（`atomic`，从 1 起）、`RunnerID int64`、`closed chan struct{}`。方法：

- `OpenStream() *Stream`：分配 ID、建 `respCh`（缓冲 16）、注册进表。
- `WriteFrame(f) error`：`writeMu` 下 `conn.WriteMessage(Binary, tunnelframe.Encode(f))`。
- `readLoop()`：循环 `ReadMessage` → `Decode` → `dispatch`；出错关闭 tunnel 并 drain 所有 stream（向 respCh 投递合成的 RESP_ERROR，然后 close）。
- `dispatch(f)`：`stream_id==0` 处理 PING/PONG/HELLO；否则查表投递到 `st.respCh`（非阻塞，满则丢弃并记 warn —— 因为 respCh 只承载控制帧 START/END/ERROR/CREDIT，body 帧在 proxy 层直接消费，见 Task 3.1 说明）。
- `pingLoop()`：每 10s 发 PING；`closeStream(id)`、`Close()`。

为测试提供 `newTunnelForTest(conn)`：跳过真实 WS，注入可写 mock。

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./relay/internal/tunnel/ -run TestTunnel -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add relay/internal/tunnel/tunnel.go relay/internal/tunnel/tunnel_test.go
git commit -m "feat(gateway): tunnel connection with stream table and heartbeat"
```

## Task 2.4: 隧道注册表 + WaitForTunnel 宽限

**Files:**
- Create: `relay/internal/tunnel/registry.go`
- Test: `relay/internal/tunnel/registry_test.go`

- [ ] **Step 1: Write the failing test**

```go
func TestRegistry_WaitForTunnel_Grace(t *testing.T) {
	r := NewRegistry()
	go func() {
		time.Sleep(80 * time.Millisecond)
		r.Register(&Tunnel{RunnerID: 5})
	}()
	tun := r.WaitForTunnel(context.Background(), 5, 500*time.Millisecond)
	if tun == nil {
		t.Fatal("expected tunnel within grace")
	}
}

func TestRegistry_WaitForTunnel_Timeout(t *testing.T) {
	r := NewRegistry()
	if tun := r.WaitForTunnel(context.Background(), 9, 50*time.Millisecond); tun != nil {
		t.Fatal("expected nil after grace timeout")
	}
}

func TestRegistry_ReconnectReplaces(t *testing.T) {
	r := NewRegistry()
	old := &Tunnel{RunnerID: 1}
	r.Register(old)
	newT := &Tunnel{RunnerID: 1}
	r.Register(newT)
	if r.Get(1) != newT {
		t.Fatal("new tunnel should replace old")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./relay/internal/tunnel/ -run TestRegistry -v`
Expected: 编译失败（`NewRegistry` 未定义）

- [ ] **Step 3: Write minimal implementation**

实现 `Registry`：`map[int64]*Tunnel` + `RWMutex`。`Register` 若已有同 ID，调用旧 `Close()` 后替换。`Get`。`WaitForTunnel`：先 `Get`，未命中则 100ms `time.Ticker` 轮询到 `grace` 或 `ctx.Done()`。`Unregister`（仅当当前登记的就是该实例时删除，避免误删新连接）。`Stats()`。

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./relay/internal/tunnel/ -run TestRegistry -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add relay/internal/tunnel/registry.go relay/internal/tunnel/registry_test.go
git commit -m "feat(gateway): tunnel registry with reconnect grace"
```

## Task 2.5: Gateway 隧道接入端点 `/runner/tunnel`

**Files:**
- Create: `relay/internal/server/handler_tunnel.go`
- Modify: `relay/internal/server/server.go`
- Test: `relay/internal/server/handler_tunnel_test.go`

- [ ] **Step 1: Write the failing test**

```go
func TestHandleTunnelWS_RejectsNonTunnelToken(t *testing.T) {
	h := newTestTunnelHandler(t) // 注入 validator + registry + origin(allowAll)
	srv := httptest.NewServer(http.HandlerFunc(h.HandleTunnelWS))
	defer srv.Close()

	// browser token（非 tunnel 类型）应被拒
	url := "ws" + strings.TrimPrefix(srv.URL, "http") + "/?token=" + validBrowserToken(t)
	_, resp, err := websocket.DefaultDialer.Dial(url, nil)
	if err == nil {
		t.Fatal("expected reject")
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestHandleTunnelWS_RegistersTunnel(t *testing.T) {
	h := newTestTunnelHandler(t)
	srv := httptest.NewServer(http.HandlerFunc(h.HandleTunnelWS))
	defer srv.Close()
	url := "ws" + strings.TrimPrefix(srv.URL, "http") + "/?token=" + validTunnelToken(t, 7)
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	// 发 HELLO
	hello, _ := json.Marshal(tunnelframe.HelloPayload{RunnerID: "7"})
	conn.WriteMessage(websocket.BinaryMessage, tunnelframe.Encode(tunnelframe.Frame{Type: tunnelframe.TypeHello, Payload: hello}))
	waitFor(t, func() bool { return h.registry.Get(7) != nil })
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./relay/internal/server/ -run TestHandleTunnelWS -v`
Expected: 编译失败（`HandleTunnelWS` 未定义）

- [ ] **Step 3: Write minimal implementation**

在 `handler_tunnel.go`：`TunnelHandler{validator, registry, originChecker, cfg, logger}`。`HandleTunnelWS`：Origin 校验 → 取 `?token=` → `ValidateToken` → `claims.ResolvedType()==TokenTypeTunnel` 否则 401 → upgrade → 读首帧 HELLO 校验 `RunnerID==claims.RunnerID` → 构造 `tunnel.Tunnel` 启动 `readLoop/pingLoop` → `registry.Register` → 阻塞至 `tun.closed` → `registry.Unregister`。在 `server.go` 注册路由 `mux.HandleFunc("/runner/tunnel", tunnelHandler.HandleTunnelWS)`，并把 `registry` 存到 `Server`（供 preview handler 用）。仅当 `cfg.Tunnel.Enabled` 时注册。

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./relay/internal/server/ -run TestHandleTunnelWS -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add relay/internal/server/handler_tunnel.go relay/internal/server/server.go relay/internal/server/handler_tunnel_test.go
git commit -m "feat(gateway): runner tunnel websocket endpoint"
```

## Task 2.6: proto 新增 ConnectTunnelCommand

**Files:**
- Modify: `proto/runner/v1/runner.proto`
- Regenerate: `proto/gen/...`

- [ ] **Step 1: Write the change**

在 `ServerMessage` oneof 追加（编号 14，接续现有 13）：

```proto
    ConnectTunnelCommand connect_tunnel = 14;
```

新增消息（放在 SubscribePodCommand 附近）：

```proto
// ConnectTunnelCommand 通知 Runner 建立到 Gateway 的 HTTP 隧道长连接
message ConnectTunnelCommand {
  string gateway_url = 1;   // e.g. wss://domain/relay（与 relay_url 同源）
  string tunnel_token = 2;  // token_type=tunnel 的 JWT（不绑定单一 pod）
}
```

- [ ] **Step 2: Regenerate**

Run: `bazel run //proto:generate`（或仓库既有 proto 生成命令）
Expected: `proto/gen/go/runner/v1/*.pb.go` 出现 `ConnectTunnelCommand` 与 `ServerMessage_ConnectTunnel`

- [ ] **Step 3: Verify build**

Run: `bazel build //proto/...`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add proto/runner/v1/runner.proto proto/gen/
git commit -m "feat(gateway): add connect_tunnel runner command"
```

## Task 2.7: Runner 隧道客户端

**Files:**
- Create: `runner/internal/tunnel/client.go`
- Test: `runner/internal/tunnel/client_test.go`

- [ ] **Step 1: Write the failing test**

```go
func TestClient_ConnectAndHello(t *testing.T) {
	// 起一个 httptest WS server 充当 gateway，接收 HELLO
	got := make(chan tunnelframe.HelloPayload, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, _ := (&websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}).Upgrade(w, r, nil)
		_, data, _ := c.ReadMessage()
		f, _ := tunnelframe.Decode(data)
		var hp tunnelframe.HelloPayload
		json.Unmarshal(f.Payload, &hp)
		got <- hp
	}))
	defer srv.Close()

	url := "ws" + strings.TrimPrefix(srv.URL, "http")
	cl := NewClient(context.Background(), url, "tok", 7, 3, nil)
	if err := cl.Connect(); err != nil {
		t.Fatal(err)
	}
	defer cl.Stop()
	select {
	case hp := <-got:
		if hp.RunnerID != "7" {
			t.Fatalf("bad runner id %q", hp.RunnerID)
		}
	case <-time.After(time.Second):
		t.Fatal("no hello received")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./runner/internal/tunnel/ -run TestClient_ConnectAndHello -v`
Expected: 编译失败（包不存在）

- [ ] **Step 3: Write minimal implementation**

`client.go` 参考 `runner/internal/relay/client*.go` 的骨架（重连 backoff、`safego`、`stopCh`、token 刷新回调）实现 `Client`：`Connect()` 拨号 `{gatewayURL}/runner/tunnel?token=`（复用 `runner/relay/client_connection.go` 的 scheme 转换与 `path.Join`），成功后发 HELLO；`readLoop` 把收到的帧交给注入的 `Dispatcher`（Task 2.8）；`Send(frame)`；`Stop()`。先用 `nil` dispatcher 让本 Task 只验证连接+HELLO。

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./runner/internal/tunnel/ -run TestClient_ConnectAndHello -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add runner/internal/tunnel/client.go runner/internal/tunnel/client_test.go
git commit -m "feat(gateway): runner tunnel client connect and hello"
```

## Task 2.8: Runner OnConnectTunnel 命令处理

**Files:**
- Create: `runner/internal/runner/message_handler_tunnel.go`
- Modify: 命令分发处（`runner/internal/runner/` 中处理 `ServerMessage` oneof 的位置，对照 `message_handler_relay.go` 的接线）
- Test: `runner/internal/runner/message_handler_tunnel_test.go`

- [ ] **Step 1: Write the failing test**

```go
func TestOnConnectTunnel_StartsClient(t *testing.T) {
	h := newTestRunnerHandler(t) // 复用现有测试脚手架
	factory := &fakeTunnelFactory{}
	h.tunnelClientFactory = factory.New

	err := h.OnConnectTunnel(client.ConnectTunnelRequest{
		GatewayURL:  "ws://127.0.0.1:1/relay",
		TunnelToken: "tok",
	})
	if err != nil {
		t.Fatal(err)
	}
	if factory.created != 1 {
		t.Fatalf("expected 1 client, got %d", factory.created)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./runner/internal/runner/ -run TestOnConnectTunnel -v`
Expected: 编译失败

- [ ] **Step 3: Write minimal implementation**

对照 `message_handler_relay.go` 的锁策略实现 `OnConnectTunnel`：已连同一 `gateway_url` → `UpdateToken`；否则停旧、建新、`Connect()`+`Start()`、原子替换。`gateway_url` 走 `config.RewriteRelayURL`（dev 环境重写）。在 oneof 分发处新增 `case *runnerv1.ServerMessage_ConnectTunnel:` 调用 `OnConnectTunnel`。定义 `client.ConnectTunnelRequest{GatewayURL, TunnelToken string}` 并在 gRPC 消息转换处填充。

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./runner/internal/runner/ -run TestOnConnectTunnel -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add runner/internal/runner/message_handler_tunnel.go runner/internal/runner/*.go runner/internal/client/*.go runner/internal/runner/message_handler_tunnel_test.go
git commit -m "feat(gateway): runner handles connect_tunnel command"
```

## Task 2.9: Backend 下发 connect_tunnel

**Files:**
- Modify: `backend/internal/service/runner/command_sender.go`
- Modify: `backend/internal/api/grpc/runner_adapter_send.go`
- Modify: `backend/internal/api/grpc/command_sender_adapter.go`
- Modify: `backend/cmd/server/services_init.go`
- Test: `backend/internal/api/grpc/runner_adapter_send_extra_test.go`

- [ ] **Step 1: Write the failing test**

```go
func TestGRPCRunnerAdapter_SendConnectTunnel(t *testing.T) {
	logger := newTestLogger()
	connMgr := runner.NewRunnerConnectionManager(logger)
	defer connMgr.Close()
	adapter := NewGRPCRunnerAdapter(connMgr, nil, nil, logger)
	mockStream := &mockRunnerStream{}
	conn := connMgr.AddConnection(1, "n", "o", mockStream)
	_ = conn
	if err := adapter.SendConnectTunnel(1, "wss://d/relay", "tok"); err != nil {
		t.Fatal(err)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./backend/internal/api/grpc/ -run TestGRPCRunnerAdapter_SendConnectTunnel -v`
Expected: 编译失败

- [ ] **Step 3: Write minimal implementation**

- `runner_adapter_send.go` 新增 `SendConnectTunnel(runnerID int64, gatewayURL, tunnelToken string) error`：构造 `&runnerv1.ServerMessage{Payload:&runnerv1.ServerMessage_ConnectTunnel{ConnectTunnel:&runnerv1.ConnectTunnelCommand{GatewayUrl:gatewayURL, TunnelToken:tunnelToken}}}` 走 `conn.SendMessage`（对照现有 `SendSubscribePod`）。
- `RunnerCommandSender` 接口与 `GRPCCommandSender`、`NoOpCommandSender`、各 mock 增加同名方法（`command_sender.go`、`command_sender_adapter.go`、`test_helper_test.go`、`pod_commands_test.go`）。
- `services_init.go`：`SetInitializedCallback` 现有回调尾部追加：若 `cfg.Tunnel` 下发开关开启，`token,_ := relayTokens.GenerateTypedToken("", runnerID, 0, orgID, "tunnel", "", time.Hour)`，`commandSender.SendConnectTunnel(runnerID, cfg.RelayURL(), token)`（失败仅记 warn，隧道靠 runner 重连兜底）。

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./backend/internal/api/grpc/ -run TestGRPCRunnerAdapter_SendConnectTunnel -v` 及 `go test ./backend/internal/service/runner/...`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add backend/internal/api/grpc/ backend/internal/service/runner/ backend/cmd/server/services_init.go
git commit -m "feat(gateway): backend dispatches connect_tunnel on runner init"
```

---

# Phase 3 — HTTP preview（HTML/JS/CSS/图片）

## Task 3.1: proxy header 卫生

**Files:**
- Create: `relay/internal/proxy/headers.go`
- Test: `relay/internal/proxy/headers_test.go`

- [ ] **Step 1: Write the failing test**

```go
func TestSanitizeRequestHeaders(t *testing.T) {
	in := http.Header{
		"Connection":       {"keep-alive"},
		"Upgrade":          {"websocket"},
		"Proxy-Connection": {"x"},
		"X-Forwarded-For":  {"1.2.3.4"},
		"Content-Type":     {"text/html"},
	}
	out := SanitizeRequestHeaders(in, "9.9.9.9", "https", "host")
	if _, ok := out["Connection"]; ok {
		t.Fatal("hop-by-hop must be stripped")
	}
	if out.Get("Content-Type") != "text/html" {
		t.Fatal("passthrough header lost")
	}
	if out.Get("X-Forwarded-For") != "9.9.9.9" {
		t.Fatal("XFF must be rewritten, not trusted from client")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./relay/internal/proxy/ -run TestSanitizeRequestHeaders -v`
Expected: 编译失败

- [ ] **Step 3: Write minimal implementation**

实现 `SanitizeRequestHeaders(in http.Header, clientIP, proto, host string) http.Header`：克隆后删除 hop-by-hop（`Connection, Proxy-Connection, Keep-Alive, Transfer-Encoding, TE, Trailer, Upgrade` 以及 `Connection` 列出的 header）与入站 `X-Forwarded-*`、`Forwarded`，再写入受控 `X-Forwarded-For/Proto/Host`。同文件加 `SanitizeResponseHeaders(in http.Header) http.Header`（剥 hop-by-hop，保留 `Content-Type/Content-Length/Content-Range/Accept-Ranges/Content-Encoding/Cache-Control` 等）。

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./relay/internal/proxy/ -run TestSanitizeRequestHeaders -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add relay/internal/proxy/headers.go relay/internal/proxy/headers_test.go
git commit -m "feat(gateway): proxy header sanitation"
```

## Task 3.2: HTTP 代理（stream ↔ HTTP，含流控写回）

**Files:**
- Create: `relay/internal/proxy/http.go`
- Test: `relay/internal/proxy/http_test.go`

- [ ] **Step 1: Write the failing test**

用一个 fake tunnel（内存双工，实现 `WriteFrame` 并可注入对端帧）驱动：

```go
func TestProxyHTTP_StreamsResponse(t *testing.T) {
	ft := newFakeTunnel()
	// 对端行为：收到 REQ_START 后回 RESP_START(200)+RESP_BODY("hi")+RESP_END
	ft.onReqStart = func(st *tunnel.Stream, p tunnelframe.ReqStartPayload) {
		ft.inject(st.ID, tunnelframe.TypeRespStart, mustJSON(tunnelframe.RespStartPayload{Status: 200, Header: http.Header{"Content-Type": {"text/plain"}}}))
		ft.inject(st.ID, tunnelframe.TypeRespBody, []byte("hi"))
		ft.inject(st.ID, tunnelframe.TypeRespEnd, nil)
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/preview/pod1/index.html", nil)

	err := ProxyHTTP(context.Background(), ft, rec, req, ProxyParams{
		PodKey: "pod1", Target: "127.0.0.1:3000", Path: "/index.html", WindowBytes: 1 << 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	if rec.Code != 200 || rec.Body.String() != "hi" {
		t.Fatalf("bad response: %d %q", rec.Code, rec.Body.String())
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./relay/internal/proxy/ -run TestProxyHTTP -v`
Expected: 编译失败

- [ ] **Step 3: Write minimal implementation**

`ProxyHTTP(ctx, tun tunnelIface, w http.ResponseWriter, r *http.Request, p ProxyParams) error`：
1. `st := tun.OpenStream()`，`defer tun.CloseStream(st.ID)`。
2. 发 `REQ_START`（`SanitizeRequestHeaders` 后的 header + target + method + path + rawquery + is_websocket=false）。
3. 若有 body：分块 ≤`MaxChunk` 读 `r.Body`，每块前 `st.sendWin.acquire`，发 `REQ_BODY`；末尾 `REQ_END`。
4. 从 `st.respCh` 读控制帧：首个必须 `RESP_START`（写 status+header 到 `w`）或 `RESP_ERROR`（映射 502 + code）。`http.Flusher` flush。
5. body：从 `st.respCh` 读 `RESP_BODY` 写入 `w` 并 flush，每写一块回发 `CREDIT`；`RESP_END` 结束。`ctx`/超时取消发 `STREAM_CANCEL`。

> 说明：为让 body 帧也能到达 proxy，`Tunnel.dispatch` 对 `RESP_BODY/START/END/ERROR/CREDIT` 均投递到 `st.respCh`（缓冲足够，配合 credit 背压保证不会无界堆积）。

定义 `tunnelIface`（`OpenStream/WriteFrame/CloseStream`）便于 fake 注入。

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./relay/internal/proxy/ -run TestProxyHTTP -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add relay/internal/proxy/http.go relay/internal/proxy/http_test.go
git commit -m "feat(gateway): http reverse proxy over tunnel stream"
```

## Task 3.3: Runner 本地 HTTP 转发（含 loopback 校验）

**Files:**
- Create: `runner/internal/tunnel/local_http.go`
- Create: `runner/internal/tunnel/dispatcher.go`
- Test: `runner/internal/tunnel/local_http_test.go`

- [ ] **Step 1: Write the failing test**

```go
func TestServeLocalHTTP_RejectsNonLoopback(t *testing.T) {
	err := validateTarget("10.0.0.5:80")
	if err == nil {
		t.Fatal("non-loopback must be rejected")
	}
	if err := validateTarget("127.0.0.1:3000"); err != nil {
		t.Fatalf("loopback should pass: %v", err)
	}
}

func TestServeLocalHTTP_StreamsUpstream(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		io.WriteString(w, "pong")
	}))
	defer upstream.Close()
	target := strings.TrimPrefix(upstream.URL, "http://") // 127.0.0.1:port

	fc := newFakeFrameSink()
	serveLocalHTTP(context.Background(), fc, streamID(1), tunnelframe.ReqStartPayload{
		Method: "GET", Path: "/", Target: target,
	}, nil, 1<<20)

	if fc.status() != 200 || fc.body() != "pong" {
		t.Fatalf("bad: %d %q", fc.status(), fc.body())
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./runner/internal/tunnel/ -run TestServeLocalHTTP -v`
Expected: 编译失败

- [ ] **Step 3: Write minimal implementation**

- `validateTarget(target)`：`net.SplitHostPort` → `net.ParseIP` 必须 `IsLoopback()`（或 host==`localhost`）；否则 error。
- `serveLocalHTTP(ctx, sink frameSink, id, reqStart, body io.Reader, window int)`：构造 `http.Request` 到 `http://`+target+path，透传 header 与 `Range`；`http.Client{CheckRedirect: 不跟随}` 请求；回 `RESP_START`（status+`SanitizeResponseHeaders`），分块读 body（≤`MaxChunk`，credit acquire）发 `RESP_BODY`，末 `RESP_END`；错误发 `RESP_ERROR{target_unreachable}`。
- `dispatcher.go`：`Dispatcher` 收帧按 `stream_id` 建/查本地 stream，`REQ_START` 起 goroutine 调 `serveLocalHTTP`；`REQ_BODY` 写入该 stream 的 body pipe；`CREDIT` 补 send 窗；`STREAM_CANCEL` 取消。把 `Dispatcher` 接到 Task 2.7 的 `Client.readLoop`。

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./runner/internal/tunnel/ -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add runner/internal/tunnel/local_http.go runner/internal/tunnel/dispatcher.go runner/internal/tunnel/local_http_test.go
git commit -m "feat(gateway): runner local http forwarding with loopback guard"
```

## Task 3.4: 每 pod 并发/排队限制

**Files:**
- Create: `relay/internal/tunnel/limits.go`
- Test: `relay/internal/tunnel/limits_test.go`

- [ ] **Step 1: Write the failing test**

```go
func TestPodLimiter_AcquireRelease(t *testing.T) {
	l := NewPodLimiter(1, 0, 10*time.Millisecond) // 并发1，队列0
	rel, err := l.Acquire(context.Background(), "pod1")
	if err != nil {
		t.Fatal(err)
	}
	// 队列0，第二个立即 busy
	if _, err := l.Acquire(context.Background(), "pod1"); err == nil {
		t.Fatal("expected busy")
	}
	rel()
	if r2, err := l.Acquire(context.Background(), "pod1"); err != nil {
		t.Fatal(err)
	} else {
		r2()
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./relay/internal/tunnel/ -run TestPodLimiter -v`
Expected: 编译失败

- [ ] **Step 3: Write minimal implementation**

`PodLimiter`：`map[podKey]*podSlot`，每 pod 一个带缓冲 chan（容量=maxConcurrent）+ 排队计数。`Acquire(ctx, podKey)`：尝试非阻塞入 slot；满则若 `queueLen<maxQueue` 入队等待 `queueTimeout`，否则返回 `ErrTargetBusy`。返回 release 闭包。借鉴 doops `acquire` 的 select 结构但按 pod 维度、无 opSlot 串行（preview 允许并发）。

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./relay/internal/tunnel/ -run TestPodLimiter -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add relay/internal/tunnel/limits.go relay/internal/tunnel/limits_test.go
git commit -m "feat(gateway): per-pod concurrency and queue limiter"
```

## Task 3.5: Backend ResolvePreviewRoute + pod 元数据

**Files:**
- Modify: `backend/internal/domain/agentpod/pod.go`
- Create: `backend/migrations/0001XX_pod_preview.up.sql` + `.down.sql`
- Create: `backend/internal/service/relay/preview.go`
- Test: `backend/internal/service/relay/preview_test.go`

- [ ] **Step 1: Write the failing test**

```go
func TestResolvePreviewRoute(t *testing.T) {
	pod := &agentpod.Pod{RunnerID: 7, PreviewPort: 3000}
	pod.SetActiveForTest() // 或构造 active 状态
	r, err := ResolvePreviewRoute(pod)
	if err != nil {
		t.Fatal(err)
	}
	if r.Target != "127.0.0.1:3000" || r.RunnerID != 7 {
		t.Fatalf("bad route %+v", r)
	}

	if _, err := ResolvePreviewRoute(&agentpod.Pod{RunnerID: 7, PreviewPort: 0}); err == nil {
		t.Fatal("port 0 must error preview_disabled")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./backend/internal/service/relay/ -run TestResolvePreviewRoute -v`
Expected: 编译失败（`PreviewPort`/`ResolvePreviewRoute` 未定义）

- [ ] **Step 3: Write minimal implementation**

- `pod.go`：结构增加 `PreviewPort int` 与 `PreviewPath string`（GORM 字段 + json tag）。
- 迁移 up：`ALTER TABLE pods ADD COLUMN preview_port INT NOT NULL DEFAULT 0; ALTER TABLE pods ADD COLUMN preview_path VARCHAR(255) NOT NULL DEFAULT '';`；down 反向 drop。
- `preview.go`：
```go
type PreviewRoute struct {
	RunnerID int64
	Target   string
	Path     string
}
var (
	ErrPreviewDisabled = errors.New("preview_disabled")
	ErrPodNotActive    = errors.New("pod_not_active")
)
func ResolvePreviewRoute(pod *agentpod.Pod) (PreviewRoute, error) {
	if pod == nil || !pod.IsActive() || pod.RunnerID == 0 {
		return PreviewRoute{}, ErrPodNotActive
	}
	if pod.PreviewPort <= 0 {
		return PreviewRoute{}, ErrPreviewDisabled
	}
	return PreviewRoute{RunnerID: pod.RunnerID, Target: fmt.Sprintf("127.0.0.1:%d", pod.PreviewPort), Path: pod.PreviewPath}, nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./backend/internal/service/relay/ -run TestResolvePreviewRoute -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add backend/internal/domain/agentpod/pod.go backend/migrations/0001XX_pod_preview.* backend/internal/service/relay/preview.go backend/internal/service/relay/preview_test.go
git commit -m "feat(gateway): pod preview metadata and route resolver"
```

## Task 3.6: Backend preview REST API

**Files:**
- Create: `backend/internal/api/rest/v1/pod_preview.go`
- Modify: `backend/internal/api/rest/router.go`（注册路由）
- Test: `backend/internal/api/rest/v1/pod_preview_test.go`

- [ ] **Step 1: Write the failing test**

```go
func TestGetPodPreview_ReturnsTokenAndURL(t *testing.T) {
	d := newTestDeps(t) // active pod PreviewPort=3000, 有读权限, relay 健康
	w := performGET(t, d, "/api/v1/orgs/acme/pods/pod1/preview")
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
	var resp struct {
		PreviewBaseURL string `json:"preview_base_url"`
		Token          string `json:"token"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Token == "" || !strings.Contains(resp.PreviewBaseURL, "/preview/pod1/") {
		t.Fatalf("bad resp %+v", resp)
	}
}

func TestGetPodPreview_DisabledReturns404(t *testing.T) {
	d := newTestDeps(t)
	d.setPodPreviewPort("pod1", 0)
	w := performGET(t, d, "/api/v1/orgs/acme/pods/pod1/preview")
	if w.Code != 404 {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./backend/internal/api/rest/v1/ -run TestGetPodPreview -v`
Expected: 编译失败

- [ ] **Step 3: Write minimal implementation**

`pod_preview.go` `handleGetPodPreview`：授权 pod 读 → `ResolvePreviewRoute` → relay 选择（`SelectRelayForPodGeo`）→ `SendConnectTunnel`（失败返回 503，**不发 token**，修复 A3）→ `GenerateTypedToken(podKey, runnerID, userID, orgID, "preview", route.Target, 30min)` → 返回：
```json
{"preview_base_url":"<relayHTTPBase>/preview/<podKey>/","session_url":"<...>/preview/<podKey>/__session?token=<jwt>","token":"<jwt>","expires_at":"..."}
```
错误映射：`ErrPreviewDisabled`→404、`ErrPodNotActive`→409、无权限→403、relay 不可用→503。在 `router.go` pod 组注册 `GET .../pods/:key/preview`。

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./backend/internal/api/rest/v1/ -run TestGetPodPreview -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add backend/internal/api/rest/v1/pod_preview.go backend/internal/api/rest/router.go backend/internal/api/rest/v1/pod_preview_test.go
git commit -m "feat(gateway): backend pod preview endpoint"
```

## Task 3.7: Gateway preview HTTP 入口 + session 交换

**Files:**
- Create: `relay/internal/server/handler_preview.go`
- Create: `relay/internal/server/handler_preview_session.go`
- Modify: `relay/internal/server/server.go`
- Test: `relay/internal/server/handler_preview_test.go`

- [ ] **Step 1: Write the failing test**

```go
func TestPreview_UnauthorizedWithoutToken(t *testing.T) {
	h := newTestPreviewHandler(t) // registry 空, validator
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/preview/pod1/index.html", nil)
	h.HandlePreview(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestPreview_OfflineReturns502(t *testing.T) {
	h := newTestPreviewHandler(t)
	tok := validPreviewToken(t, "pod1", 7, "127.0.0.1:3000")
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/preview/pod1/index.html?token="+tok, nil)
	h.HandlePreview(rec, req) // registry 无 runner 7
	if rec.Code != http.StatusBadGateway {
		t.Fatalf("expected 502, got %d", rec.Code)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./relay/internal/server/ -run TestPreview -v`
Expected: 编译失败

- [ ] **Step 3: Write minimal implementation**

`handler_preview.go` `HandlePreview`：
1. 解析 path `/preview/{podKey}/{rest...}`。
2. 取 token：`gw_preview` cookie 优先，回退 `?token=`。`ValidateToken` → `ResolvedType()==TokenTypePreview` 且 `claims.PodKey==podKey` 否则 401。
3. `tun := registry.WaitForTunnel(ctx, claims.RunnerID, cfg.Tunnel.ReconnectGrace)`；nil→502 `target_offline`。
4. `rel, err := limiter.Acquire(ctx, podKey)`；`ErrTargetBusy`→429；`defer rel()`。
5. `proxy.ProxyHTTP(ctx, tun, w, r, ProxyParams{PodKey, Target: claims.PreviewTarget, Path: "/"+rest, WindowBytes: cfg.Tunnel.StreamWindowBytes, Timeout: cfg.Tunnel.StreamTimeout})`。
6. 若 `RESP_START.status==101`→转 `proxy.ProxyWebSocket`（Phase 4，先返回 501 占位分支，Task 4.3 替换）。
7. 记录访问日志。

`handler_preview_session.go` `HandlePreviewSession`：`GET /preview/{podKey}/__session?token=` 校验后 `Set-Cookie gw_preview`（`HttpOnly;Secure;SameSite=Lax;Path=/preview/{podKey};Max-Age=token剩余`）→ 302 到 `/preview/{podKey}/`。

`server.go`：`mux.HandleFunc("/preview/", previewHandler.route)`，`route` 内分派 `__session` 与普通请求。仅 `cfg.Tunnel.Enabled` 注册。

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./relay/internal/server/ -run TestPreview -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add relay/internal/server/handler_preview.go relay/internal/server/handler_preview_session.go relay/internal/server/server.go relay/internal/server/handler_preview_test.go
git commit -m "feat(gateway): preview http entrypoint and session cookie"
```

## Task 3.8: 端到端集成测试（HTML/图片）

**Files:**
- Create: `relay/internal/server/preview_e2e_test.go`

- [ ] **Step 1: Write the failing test**

用真实 httptest Gateway + 内存 runner tunnel client（连回 Gateway 的 `/runner/tunnel`）+ 内存 local svc，验证浏览器侧 GET 能拿到 HTML 与图片字节：

```go
func TestPreviewE2E_HTMLAndImage(t *testing.T) {
	// 1. 起 Gateway server（含 tunnel + preview 路由）
	// 2. 起本地 svc：/ 返回 text/html "<h1>ok</h1>"，/logo.png 返回 image/png 字节
	// 3. runner tunnel client 连 Gateway，dispatcher target 指向本地 svc
	// 4. 浏览器 GET /preview/pod1/ 与 /preview/pod1/logo.png（带 preview token）
	// 断言：200、Content-Type 正确、body 完整
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./relay/internal/server/ -run TestPreviewE2E_HTMLAndImage -v`
Expected: FAIL（组件尚未接线完整）

- [ ] **Step 3: Make it pass**

补齐接线：确保 `Client.readLoop`→`Dispatcher`→`serveLocalHTTP`→回帧→`ProxyHTTP` 全链路 credit/close 正确。修复过程中发现的 bug 就地修。

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./relay/internal/server/ -run TestPreviewE2E -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add relay/internal/server/preview_e2e_test.go
git commit -m "test(gateway): e2e preview for html and image"
```

## Task 3.9: 反代路由

**Files:**
- Modify: `deploy/dev/traefik/dynamic/http.yml`
- Modify: `deploy/kubernetes/cluster-oilan/40-ingress.yaml`

- [ ] **Step 1: Add dev traefik route**

在 `http.yml` `routers` 增加（不 strip 前缀，Gateway 内部按 `/preview/` 路由）：

```yaml
    preview:
      entryPoints:
        - web
      rule: "PathPrefix(`/preview`)"
      service: relay
      priority: 40
```

- [ ] **Step 2: Add k8s ingress path**

在 `40-ingress.yaml` relay 服务下增加 `/preview` path，转发到 relay:8090（保留前缀，不 rewrite）。

- [ ] **Step 3: Verify dev routing**

Run: 本地起 relay + traefik 后 `curl -i http://localhost:$HTTP_PORT/preview/pod1/`（无 token）
Expected: 401（说明路由已到 Gateway preview handler）

- [ ] **Step 4: Commit**

```bash
git add deploy/dev/traefik/dynamic/http.yml deploy/kubernetes/cluster-oilan/40-ingress.yaml
git commit -m "chore(gateway): route /preview to gateway"
```

---

# Phase 4 — 媒体（视频/Range/SSE）与 WebSocket + 前端

## Task 4.1: 视频 Range 透传 e2e

**Files:**
- Create: `relay/internal/server/preview_range_test.go`

- [ ] **Step 1: Write the failing test**

```go
func TestPreviewE2E_VideoRange(t *testing.T) {
	// 本地 svc 用 http.ServeContent 提供 1MB 视频，支持 Range
	// 浏览器 GET /preview/pod1/movie.mp4 带 Range: bytes=0-1023
	// 断言：206、Content-Range: bytes 0-1023/1048576、body 长度 1024
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./relay/internal/server/ -run TestPreviewE2E_VideoRange -v`
Expected: FAIL if Range/206 未正确透传

- [ ] **Step 3: Make it pass**

确认 `serveLocalHTTP` 透传请求 `Range` 头、`SanitizeResponseHeaders` 保留 `Content-Range/Accept-Ranges`、`ProxyHTTP` 原样写 206 status。修复偏差。

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./relay/internal/server/ -run TestPreviewE2E_VideoRange -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add relay/internal/server/preview_range_test.go
git commit -m "test(gateway): video range request passthrough"
```

## Task 4.2: SSE 流式 flush

**Files:**
- Create: `relay/internal/server/preview_sse_test.go`
- Modify（如需）: `relay/internal/proxy/http.go`

- [ ] **Step 1: Write the failing test**

```go
func TestPreviewE2E_SSE(t *testing.T) {
	// 本地 svc 每 20ms flush 一条 "data: n\n\n" 共 3 条, Content-Type text/event-stream
	// 浏览器读取，断言在 svc 完成前就能逐条收到（验证不缓冲）
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./relay/internal/server/ -run TestPreviewE2E_SSE -v`
Expected: FAIL if buffered

- [ ] **Step 3: Make it pass**

`ProxyHTTP` 每 `RESP_BODY` 后调用 `http.Flusher.Flush()`；runner 侧 `serveLocalHTTP` 对 `text/event-stream` 不缓冲，读到即发帧（禁用 bufio 聚合）。

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./relay/internal/server/ -run TestPreviewE2E_SSE -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add relay/internal/server/preview_sse_test.go relay/internal/proxy/http.go
git commit -m "feat(gateway): streaming flush for sse"
```

## Task 4.3: WebSocket 透传

**Files:**
- Create: `relay/internal/proxy/websocket.go`
- Modify: `relay/internal/server/handler_preview.go`（替换 101 占位分支）
- Modify: `runner/internal/tunnel/local_http.go`（WS 拨号）
- Test: `relay/internal/server/preview_ws_test.go`

- [ ] **Step 1: Write the failing test**

```go
func TestPreviewE2E_WebSocket(t *testing.T) {
	// 本地 svc: WS echo server
	// 浏览器连 /preview/pod1/ws/echo (preview token via cookie/query)
	// 发 "ping" 收 "ping"
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./relay/internal/server/ -run TestPreviewE2E_WebSocket -v`
Expected: FAIL

- [ ] **Step 3: Write minimal implementation**

- `proxy/websocket.go` `ProxyWebSocket`：升级浏览器连接；发 `REQ_START{is_websocket:true}`；等 `RESP_START{status:101}`；之后两侧循环：浏览器帧→`WS_DATA`（带 opcode 标志字节前缀）→ tunnel；tunnel `WS_DATA`→浏览器；任一关闭发 `WS_CLOSE`。credit 复用。
- runner `local_http.go`：`is_websocket` 时用 `websocket.DefaultDialer` 拨 `ws://`+target+path，双向搬运，回 `RESP_START{101}` 后走 `WS_DATA`。
- `handler_preview.go`：`RESP_START.status==101` 分支调用 `ProxyWebSocket`（改为先升级再判断的顺序：WS 请求以 `Upgrade` 头识别，直接走 `ProxyWebSocket`）。

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./relay/internal/server/ -run TestPreviewE2E_WebSocket -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add relay/internal/proxy/websocket.go relay/internal/server/handler_preview.go runner/internal/tunnel/local_http.go relay/internal/server/preview_ws_test.go
git commit -m "feat(gateway): websocket passthrough over tunnel"
```

## Task 4.4: 前端 preview hook + 面板

**Files:**
- Create: `clients/web-user/src/hooks/usePodPreview.ts`
- Create: `clients/web-user/src/components/PreviewPanel.tsx`
- Test: `clients/web-user/src/hooks/usePodPreview.test.ts`

- [ ] **Step 1: Write the failing test**

```ts
import { describe, it, expect, vi } from "vitest";
import { buildPreviewSrc } from "./usePodPreview";

describe("buildPreviewSrc", () => {
  it("uses session url so iframe src has no raw token", () => {
    const src = buildPreviewSrc({
      previewBaseUrl: "https://d/preview/pod1/",
      sessionUrl: "https://d/preview/pod1/__session?token=JWT",
      token: "JWT",
      expiresAt: "",
    });
    expect(src).toContain("__session");
    // 建立 session 后 iframe 落到 base
    expect(src.startsWith("https://d/preview/pod1/")).toBe(true);
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd clients/web-user && pnpm vitest run src/hooks/usePodPreview.test.ts`
Expected: FAIL（模块不存在）

- [ ] **Step 3: Write minimal implementation**

- `usePodPreview.ts`：`fetch` backend `GET /api/v1/orgs/:slug/pods/:key/preview` 返回 `{previewBaseUrl, sessionUrl, token, expiresAt}`；导出 `buildPreviewSrc(info)`（返回 `sessionUrl`，浏览器访问后 302 到 base 并种 cookie）；到期前刷新。
- `PreviewPanel.tsx`：`<iframe src={buildPreviewSrc(info)} sandbox="allow-scripts allow-same-origin allow-forms" />` + 刷新/新窗口打开按钮 + 加载/离线/403 状态。

- [ ] **Step 4: Run test to verify it passes**

Run: `cd clients/web-user && pnpm vitest run src/hooks/usePodPreview.test.ts`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add clients/web-user/src/hooks/usePodPreview.ts clients/web-user/src/components/PreviewPanel.tsx clients/web-user/src/hooks/usePodPreview.test.ts
git commit -m "feat(gateway): web-user preview hook and panel"
```

---

# Phase 5 — 运维化

## Task 5.1: 隧道/代理 OTel 指标

**Files:**
- Modify: `relay/internal/otel/` （新增 gauge/counter 注册）
- Modify: `relay/internal/tunnel/registry.go`、`relay/internal/proxy/http.go`（打点）
- Test: `relay/internal/otel/*_test.go`（若有既有模式）

- [ ] **Step 1: Write the failing test**

对照现有 `RegisterRelayGauges` 的测试模式，新增 `RegisterTunnelGauges(activeTunnels, activeStreams func() int)` 的注册测试。

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./relay/internal/otel/ -v`
Expected: FAIL

- [ ] **Step 3: Write minimal implementation**

新增 gauge：`gateway.tunnels.active`、`gateway.streams.active`；counter：`gateway.preview.requests{status}`、`gateway.preview.bytes{dir}`。在 `server.Start` 调 `RegisterTunnelGauges(registry.Stats...)`；`ProxyHTTP` 结束时记录 status 与 bytes。

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./relay/internal/otel/ -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add relay/internal/otel/ relay/internal/tunnel/registry.go relay/internal/proxy/http.go
git commit -m "feat(gateway): tunnel and preview otel metrics"
```

## Task 5.2: 心跳上报隧道统计

**Files:**
- Modify: `relay/internal/backend/client.go`（`HeartbeatRequest` 加字段）
- Modify: `backend/internal/api/rest/internal/relay_types.go`、`relay_heartbeat.go`（接收）
- Test: `relay/internal/backend/client_test.go`、`backend/internal/api/rest/internal/relay_handler_test.go`

- [ ] **Step 1: Write the failing test**

在 relay client 心跳测试断言请求体包含 `active_tunnels`；backend handler 测试断言能解析该字段不报错。

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./relay/internal/backend/ ./backend/internal/api/rest/internal/ -run Heartbeat -v`
Expected: FAIL

- [ ] **Step 3: Write minimal implementation**

`HeartbeatRequest` 加 `ActiveTunnels int json:"active_tunnels,omitempty"`；relay 心跳回调填入 `registry.Stats()`；backend 侧结构体加对应字段并在 handler 记录（可先仅日志/metrics，不落库）。

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./relay/internal/backend/ ./backend/internal/api/rest/internal/ -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add relay/internal/backend/client.go backend/internal/api/rest/internal/relay_types.go backend/internal/api/rest/internal/relay_heartbeat.go
git commit -m "feat(gateway): report tunnel stats in heartbeat"
```

## Task 5.3: 全量回归 + 文档更新

**Files:**
- Modify: `docs/architecture/remediation-plan.md`（Phase 4 隧道落地标注）
- Modify: `relay/README.md`（若存在，补 Gateway 说明）

- [ ] **Step 1: Run full regression**

Run: `bazel test //relay/... //backend/... //runner/...`
Expected: 全绿（终端/ACP 既有测试零回归）

- [ ] **Step 2: Update docs**

在 remediation-plan Phase 4 勾选隧道 Gateway 已落地；relay README 增加 `/runner/tunnel` 与 `/preview` 端点、配置项说明。

- [ ] **Step 3: Commit**

```bash
git add docs/architecture/remediation-plan.md relay/README.md
git commit -m "docs(gateway): document tunnel gateway rollout"
```

---

## Self-Review 记录

- **Spec 覆盖**：内网隧道（Phase 2）、HTTP/HTML/图片（Phase 3）、视频/Range/SSE/WS（Phase 4）、鉴权/Origin/SSRF（Task 1.2/1.4/3.3）、限流排队（Task 3.4）、前端接入（Task 4.4）、运维（Phase 5）均有对应任务。设计文档第 1.2 节的 A1–A5 分别由 Task 1.4 / 1.1 / 3.6 / 2.5 / 2.2+3.2 覆盖。
- **类型一致**：`RelayClaims.ResolvedType()`、`TokenType*` 常量、`tunnelframe.Frame`/`FrameType`、`Tunnel.OpenStream/WriteFrame/CloseStream`、`creditWindow.acquire/add/close`、`ResolvePreviewRoute`/`PreviewRoute`、`SendConnectTunnel` 在跨任务引用中命名一致。
- **占位符**：无 TBD/TODO；每个代码步骤含可编译代码或精确修改点。
- **风险点**：`Tunnel.dispatch` 将 body 帧投递到 `st.respCh` 的缓冲与 credit 背压需在 Task 3.2/3.8 联调验证内存上界；proto 编号 14 需确认生成时无冲突（Task 2.6 build 校验）。
