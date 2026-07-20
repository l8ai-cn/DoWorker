# Oilan Endpoints

- App: https://dowork.l8ai.cn (`/api`, `/proto.`, `/relay`, `/health`)
- Isolated Pod preview: `https://<pod-key>.l8ai.cn` (`/preview` only)
- Mobile Worker entry: https://mobile.l8ai.cn
- Marketplace Storefront: https://market.l8ai.cn
- Marketplace API: https://market.l8ai.cn/api/marketplace/v1
- Organization marketplace: https://dowork.l8ai.cn/dev-org/marketplace
- Admin console: https://admin.l8ai.cn (separate host, no `/admin` basePath)
- Object storage (presigned URLs): https://minio.dowork.l8ai.cn
- Test account: `admin@agentsmesh.local / Ab123456`

DNS for `dowork.l8ai.cn`, `market.l8ai.cn`, `mobile.l8ai.cn`,
`admin.l8ai.cn`, `*.l8ai.cn`, and `minio.dowork.l8ai.cn` must point at the
Oilan node.

Each Pod preview uses `<pod-key>.l8ai.cn`, covered by the existing
`l8ai-wildcard-tls` Secret. Relay accepts `/preview` only when the request Host
matches the Pod key in the path, so the wildcard Ingress does not create a
shared preview origin.

All public URLs share one domain family so relay/WebSocket URLs from
`GetPodConnection` match the page origin. Mixed `l8an.cn` / `l8ai.cn` hosts
caused terminal attach 403s.

Ingress-nginx may expose NodePort `10007` for HTTP so external `:10007` reaches
the controller:

```bash
kubectl -n ingress-nginx patch svc ingress-nginx-controller -p \
  '{"spec":{"ports":[{"name":"http","port":80,"targetPort":"http","nodePort":10007}]}}'
```
