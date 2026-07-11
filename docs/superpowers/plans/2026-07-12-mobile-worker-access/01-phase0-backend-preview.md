# Phase 0 Backend 与 Preview

## Task 1: Pod Subscription Fail-Closed

**Files**

- Create: `backend/internal/api/connect/pod/connection_test.go`
- Modify: `backend/internal/api/connect/pod/connection.go`
- Modify: `backend/internal/api/connect/pod/server.go`

### RED

增加用例：

```go
func TestGetPodConnection_SubscribeFailureUnavailable(t *testing.T) {
    srv := newConnectionServer(t, errors.New("runner offline"))
    _, err := srv.GetPodConnection(userContext(),
        connect.NewRequest(&podv1.GetPodConnectionRequest{
            OrgSlug: "acme", PodKey: "pod-1",
        }))
    require.Equal(t, connect.CodeUnavailable, connectCodeOf(t, err))
}
```

另测 `commandSender == nil`、`RunnerID == 0`，均为 `CodeUnavailable`。

Run:

```bash
cd backend
go test ./internal/api/connect/pod -run 'TestGetPodConnection_' -count=1
```

Expected: 新增用例 FAIL，当前实现仍返回 connection info。

### GREEN

删除 warning-only 路径。依赖、Runner ID 或 `SendSubscribePod` 失败时立即返回
`CodeUnavailable`，只在 dispatch 成功后签发 Browser Token。

Run 同一命令，Expected: PASS。

Commit:

```bash
git add backend/internal/api/connect/pod/connection.go \
  backend/internal/api/connect/pod/connection_test.go \
  backend/internal/api/connect/pod/server.go
git commit -m "fix(relay): fail closed on pod subscription"
```

## Task 1b: Relay Publisher Ready Acknowledgement

详细步骤见 `01a-relay-subscription-ready.md`。Task 1 只保证 dispatch 失败时
阻断；Task 1b 进一步保证 Runner 已建立 publisher 后才签发 Browser Token。

## Task 2: Preview Session Fail-Closed

**Files**

- Modify: `backend/internal/api/rest/v1/pod_preview.go`
- Modify: `backend/internal/api/rest/v1/pod_preview_test.go`

### RED

```go
func TestGetPodPreview_CommandSenderMissingReturns503(t *testing.T) {
    h := newPreviewHandler(runningPreviewPod())
    h.commandSender = nil
    w := performPreviewGET(h)
    require.Equal(t, http.StatusServiceUnavailable, w.Code)
}

func TestGetPodPreview_ResponseOmitsToken(t *testing.T) {
    w := performPreviewGET(newPreviewHandler(runningPreviewPod()))
    var body map[string]any
    require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
    require.NotContains(t, body, "token")
}
```

Run:

```bash
cd backend
go test ./internal/api/rest/v1 -run 'TestGetPodPreview_' -count=1
```

Expected: 两个新增用例 FAIL。

### GREEN

- `commandSender == nil` 返回 `preview_unavailable` 503。
- tunnel dispatch 成功后才签发 Preview Token。
- JSON 只返回 `preview_base_url/session_url/expires_at`。

Run 同一命令，Expected: PASS。

## Task 3: Preview Path Contract

**Files**

- Modify: `backend/internal/service/relay/preview.go`
- Modify: `backend/internal/service/relay/preview_test.go`
- Modify: `backend/internal/service/relay/token.go`
- Modify: `backend/internal/service/relay/token_test.go`
- Modify: `relay/internal/auth/token.go`
- Modify: `relay/internal/auth/token_test.go`
- Modify: `relay/internal/server/handler_preview.go`
- Modify: `relay/internal/server/handler_preview_test.go`

### RED

新增服务测试：

```go
func TestResolvePreviewRoute_NormalizesPath(t *testing.T) {
    route, err := ResolvePreviewRoute(&agentpod.Pod{
        RunnerID: 7, PreviewPort: 3000, PreviewPath: "/app/",
        Status: agentpod.StatusRunning,
    })
    require.NoError(t, err)
    assert.Equal(t, "/app", route.Path)
}

func TestResolvePreviewRoute_RejectsTraversal(t *testing.T) {
    _, err := ResolvePreviewRoute(&agentpod.Pod{
        RunnerID: 7, PreviewPort: 3000, PreviewPath: "/../secret",
        Status: agentpod.StatusRunning,
    })
    require.ErrorIs(t, err, ErrInvalidPreviewPath)
}
```

Relay 测试固定 `/preview/pod-1/assets/app.js` 映射为 `/app/assets/app.js`。

Run:

```bash
cd backend && go test ./internal/service/relay -run Preview -count=1
cd ../relay && go test ./internal/server -run Preview -count=1
```

Expected: path 用例 FAIL。

### GREEN

使用 `path.Clean` 和结构化 join；禁止 `..`；Preview claim 绑定 normalized
path。不得做 HTML 字符串替换。

## Task 4: PreviewConfig Revision Contract

详细步骤见 `01b-preview-config-revision.md`。该任务必须在 Mobile descriptor
和页面实现前完成，保证前端展示的是可持久化、可审计的配置，而不是临时字段。
