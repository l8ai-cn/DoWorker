# Agent Cloud Runner

[![Release](https://img.shields.io/github/v/release/l8ai-cn/AgentCloud?style=flat-square)](https://github.com/l8ai-cn/AgentCloud/releases/latest)
[![Go Report Card](https://goreportcard.com/badge/github.com/l8ai-cn/agentcloud/runner?style=flat-square)](https://goreportcard.com/report/github.com/l8ai-cn/agentcloud/runner)
[![License](https://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)](LICENSE)

Agent Cloud Runner is a lightweight agent that connects to the Agent Cloud server and executes AI agent tasks in isolated terminal environments.

## Features

- 🚀 **Multi-mode operation**: CLI or system service
- 🔒 **Secure execution**: Isolated terminal environments for each task
- 🌐 **Cross-platform**: macOS, Linux, Windows support
- 📊 **Web console**: Built-in status monitoring and log viewer
- 🔄 **Auto-reconnect**: Resilient connection to Agent Cloud server

## Installation

### macOS / Linux (one-line)

```bash
curl -fsSL https://agentcloud.ai/install.sh | sh
```

### Windows (PowerShell)

```powershell
irm https://agentcloud.ai/install.ps1 | iex
```

All binaries are available on the [Releases](https://github.com/l8ai-cn/AgentCloud/releases/latest) page (tar.gz, deb, rpm, zip).

## Quick Start

### Optional document previews

DOCX, XLS/XLSX, and PPT/PPTX previews require the `soffice` executable from
LibreOffice. The official Runner container includes it. Native Runner installs
must install LibreOffice separately; without it, the source artifact remains
downloadable and the preview operation reports an explicit failure.

### 1. Register the runner

Get a registration token from your Agent Cloud dashboard, then:

```bash
# Global: https://agentcloud.ai (or your own server address)
agent-cloud-runner register --server <SERVER_URL> --token YOUR_TOKEN
```

### 2. Start the runner

**CLI mode (foreground):**

```bash
agent-cloud-runner run
```

**System service:**

```bash
# Install as service
sudo agent-cloud-runner service install

# Start service
sudo agent-cloud-runner service start

# Check status
agent-cloud-runner service status
```

## Usage

```
Agent Cloud Runner

Usage:
  agent-cloud-runner <command> [options]

Commands:
  register    Register this runner with the Agent Cloud server
  run         Start the runner in CLI mode
  webconsole  Open the web console in browser
  service     Manage runner as a system service
  version     Show version information
  help        Show this help message

Use "agent-cloud-runner <command> --help" for more information about a command.
```

## Configuration

Configuration is stored in `~/.agentcloud/config.yaml` after registration:

```yaml
server_url: <SERVER_URL>  # Your Agent Cloud server address
node_id: my-runner
max_concurrent_pods: 5
workspace_root: /tmp/agentcloud-workspace
default_agent: claude-code
log_level: info
```

## Web Console

When using the web console command, a local web UI is available at:

```
http://127.0.0.1:19080
```

Features:
- Real-time status monitoring
- Active pods and uptime tracking
- Configuration viewer
- Live log streaming

## Building from Source

```bash
# From the repository root
cd runner
go build -o ../bin/agent-cloud-runner ./cmd/runner
go test ./...

# Back at the repository root, build the Runner image
cd ..
docker build -f runner/Dockerfile .
```

Cross-platform binaries are produced by the release pipeline; see
`.github/workflows/release.yml`.

## Release

Releases are published to [l8ai-cn/AgentCloud](https://github.com/l8ai-cn/AgentCloud).

To create a new release:

```bash
# Tag a new version
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

The CI pipeline will automatically:
- Build CLI binaries for all platforms (via GoReleaser)
- Publish to GitHub Releases
- Update Homebrew formula

## License

MIT License - see [LICENSE](LICENSE) for details.

## Links

- [Agent Cloud](https://agentcloud.ai) - Main product website
- [Documentation](https://agentcloud.ai/docs/runner) - Full documentation
- [Releases](https://github.com/l8ai-cn/AgentCloud/releases) - Download binaries
