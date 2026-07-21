# Edge Relay 与数据设计

## 1. Edge Relay 注册

Phase 2 增加 Relay 元数据：

```go
type RelayEndpoint struct {
    ID           string
    Kind         string
    PublicURL    string
    NetworkScope string
    Status       string
}
```

Edge Relay 通过现有 Relay 注册机制上报：

- 外部 HTTPS/WSS 地址
- 证书指纹和证书状态
- 网络 scope
- 已连接 Runner
- 健康和容量

手机不会通过 UDP 自动信任 Edge。是否可用由 Backend 注册状态和浏览器
显式 reachability probe 共同决定。

## 2. 连接候选

Backend 只返回满足以下条件的 Edge：

- Relay 注册有效
- 证书有效
- health check 通过
- 目标 Runner 已连接
- 用户组织允许使用

浏览器选择 Edge 后必须显式探测：

```text
GET https://edge.example/.well-known/agent-cloud-relay-health
```

响应不包含内部拓扑、Runner ID 或 Token。探测失败时展示错误，不自动切换
Cloud Relay。

## 3. Runner 连接

Runner 对 Cloud 和 Edge 使用相同 publisher/tunnel client。连接目标由
Backend 控制命令指定，不增加 Runner 本地 HTTP/WS 管理服务。

同一 Pod 在一个时刻只维持被需要的 publisher：

- 手机选择 Cloud：连接 Cloud Relay。
- 手机选择 Edge：连接指定 Edge Relay。
- 两种模式同时有观察者：允许两个标准 publisher，但共享同一 Pod output。

实现前必须评估多 publisher 对输出缓存和输入仲裁的影响，不能复制 PTY。

## 4. 数据库

Phase 1 尽量不新增持久化表。PreviewConfig 进入现有 Pod/config revision。

控制租约是短期 Relay 内存状态，不写数据库，只写审计事件。

Phase 2 如现有 Relay 注册表不能表达 Edge 元数据，再新增 migration：

```sql
ALTER TABLE relays ADD COLUMN kind VARCHAR(16) NOT NULL DEFAULT 'cloud';
ALTER TABLE relays ADD COLUMN network_scope VARCHAR(100);
ALTER TABLE relays ADD CONSTRAINT relays_kind_check
  CHECK (kind IN ('cloud', 'edge'));
```

必须新建当前序号 migration，不修改已发布历史 migration。

## 5. Edge 安全

- Edge Relay 只接受 Backend 信任链签发的 Token。
- Edge 管理 API 不暴露给手机。
- 公共地址必须 HTTPS/WSS。
- Origin allowlist 与 Cloud Relay 一致。
- Edge 不能信任源 IP 代替用户身份。
- 网络 scope 只用于候选筛选，不作为授权依据。
- Edge compromise 不应获得 Runner mTLS 私钥或 Backend 数据库凭证。

## 6. 运维

Edge Relay 必须进入 GitOps：

- Deployment/Service/Ingress
- TLS 和域名配置
- Relay 注册配置
- health/readiness probe
- 日志与指标
- 容量和告警
- 版本化回滚

不支持在用户电脑上临时运行未登记的 Relay 进程作为生产方案。
