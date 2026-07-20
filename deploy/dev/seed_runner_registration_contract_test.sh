#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")"

grep -Fq "jsonb_build_object('runners', 1000)" seed/seed.sql
grep -Fq 'ON CONFLICT (token_hash) DO UPDATE' seed/seed.sql
grep -Fq "('dev-runner-minimax', 'Development Docker Runner (MiniMax CLI)')" seed/seed.sql
grep -Fq "('dev-runner-video-studio', 'Development Docker Runner (Video Studio)')" seed/seed.sql
grep -Fq "'Admin Workspace', 'admin-workspace'" seed/seed.sql
grep -Fq "admin-workspace-runner" seed/seed.sql
grep -Fq "SET http_clone_url = 'http://gitea:3000/dev-org/demo-webapp.git'" seed/seed.sql
grep -Fq "SET http_clone_url = 'http://gitea:3000/dev-org/demo-api.git'" seed/seed.sql
grep -Fq '重放幂等基础 seed，修复开发运行时配置' lib/bootstrap.sh
