// Package agents hosts per-agent adaptation plugins.
//
// Each subdirectory represents one AI agent (e.g., claude/, codex/, aider/).
// All agent-specific logic is self-contained within its directory.
//
// # Adding a New Agent
//
// 1. Create internal/agents/<name>/ directory.
//
// 2. Create register.go with an init() function that registers needed capabilities:
//
//	func init() {
//	    // Required for ACP agents with a custom protocol:
//	    acp.RegisterTransport("worker-adapter-id", factory)
//
//	    // Required — token usage collection from session files:
//	    tokenusage.RegisterParser([]string{"slug1", "slug2"}, &myParser{})
//
//	    // Required — process name identification for PTY state monitoring:
//	    agentkit.RegisterProcessNames("process-name")
//
//	    // Optional — terminal input sanitization for PTY TUI agents:
//	    agentkit.RegisterInputAdapter("slug", &myAdapter{})
//
//	    // Optional — per-pod home directory isolation + config merging:
//	    agentkit.RegisterAgentHome(agentkit.AgentHomeSpec{...})
//	}
//
// 3. Add a blank import in internal/runner/agents_import.go:
//
//	_ "github.com/l8ai-cn/agentcloud/runner/internal/agents/<name>"
//
// # Token Usage Fixture Contract
//
// Every agent that calls tokenusage.RegisterParser MUST also ship at least
// one fixture under <pkg>/testsupport/testdata/ and expose a
// BuildFixtureSandbox helper from a separate testsupport sub-package
// (see e.g. agents/codex/testsupport/fixture.go) so the production library
// does not link the testing package. The cross-agent contract test in
// runner/internal/tokenusage/parser_contract_test.go runs each parser
// against its own fixture and fails CI if it cannot produce non-zero tokens.
// Stub agents that have no on-disk session format MUST opt out via
// tokenusage.RegisterParserOptOut(...) instead of registering a no-op
// parser, so the contract test knows to skip them.
//
// # Extension Points
//
//   - acp.RegisterTransport: Wires a canonical Worker adapter ID to its protocol
//     implementation. Every ACP Worker, including standard JSON-RPC agents, must
//     register its exact adapter ID. The backend sends that ID in CreatePodCommand;
//     the Runner never infers a transport from the executable command or defaults
//     unknown adapters to ACP.
//
//   - tokenusage.RegisterParser: Parses agent session files for token usage data.
//     Called on pod exit to collect cost metrics.
//
//   - agentkit.RegisterProcessNames: Process names used by the monitor to identify
//     agent processes in the process tree (PTY state detection).
//
//   - agentkit.RegisterInputAdapter: Adapts raw terminal input before sending to
//     the agent's PTY. Used when a TUI needs input sanitization (e.g., newline handling).
//
//   - agentkit.RegisterAgentHome: Isolates the agent's config directory per-pod.
//     Copies user config, merges platform MCP settings in the agent's config format.
//
// # Conflict Detection
//
// All Register* functions panic on duplicate registration to catch misconfiguration
// early during init(). This ensures each command name, transport type, parser slug,
// and input adapter slug is owned by exactly one agent.
package agents
