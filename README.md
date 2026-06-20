# dew

> Go-native Agent Harness

**dew** is a Go-native runtime for coding agents. It is not a single chatbot CLI;
it is a *harness* that wires together isolated capabilities—FileSystem, Session,
Sandbox, Tools, Memory, Plans, and Soul—so the same agent runtime can be exposed
as a CLI, a TUI, an RPC service, or an MCP server.

---

## Vision

Build a **thin harness, fat capabilities** agent runtime:

- **Interface-first**: every isolatable resource is an interface. Callers do not
  know whether a capability runs in-process, cross-process, or remote.
- **Event-sourced**: the agent loop emits a stream of lifecycle events. Sessions,
  audit logs, and replay all consume the same event stream.
- **Multi-surface**: one runtime, many surfaces (CLI / TUI / gRPC / MCP / A2A).
- **Monorepo → multi-repo**: each `pkg/X` is designed to become an independent
  Go module (`dew-core`, `dew-fs`, `dew-session`, `dew-tools`, `dew-memory`, …).
- **Human + team memory**: explicit project instructions (`AGENTS.md`) plus a
  learned **SOUL** layer that remembers how the agent and user like to work.

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                         Surfaces                                    │
│   CLI (cobra)   TUI (bubbletea)   RPC (gRPC)   MCP Server   A2A    │
└─────────────────────────────┬───────────────────────────────────────┘
                              │
┌─────────────────────────────▼───────────────────────────────────────┐
│                         dew-core                                    │
│   Runner │ Agent Loop │ Tool Registry │ Safety │ Context Manager   │
└─────────────────────────────┬───────────────────────────────────────┘
                              │
        ┌─────────────┬───────┴───────┬─────────────┐
        ▼             ▼               ▼             ▼
   ┌─────────┐  ┌──────────┐  ┌───────────┐  ┌──────────┐
   │  dew-fs │  │dew-session│  │ dew-sandbox│  │ dew-llm  │
   │FileSystem│  │  Session  │  │  Sandbox  │  │ Provider │
   └─────────┘  └──────────┘  └───────────┘  └──────────┘
        │             │               │             │
   Local/Mem/S3   JSONL/Memory   Docker/gVisor/…  OpenAI/Anthropic/…
```

Cross-cutting layers:

- **dew-event**: agent lifecycle events (`TurnStart`, `ToolResult`, `Error`, …).
- **dew-storage**: generic file-backed stores on top of `fs.FileSystem`.
- **dew-memory**: cross-session knowledge (rules, facts, preferences, lessons).
- **dew-plan**: task plans and checkpoints.
- **dew-soul**: persistent agent self-model and user preferences.
- **dew-log**: structured audit logging.

See [docs/design/](docs/design/) for detailed design documents.

---

## Design Philosophy

1. **Isolation as Interface**  
   FileSystem, Session, Sandbox are explicit interfaces, not implementation
   details. Upper layers only access resources through these boundaries.

2. **Thin Harness, Fat Capabilities**  
   The agent loop is minimal; context management, safety, tool registry,
   session, memory, and eval are pluggable capabilities.

3. **Event-driven + Channel-native**  
   Go `chan event.Event` and `context.Context` replace async iterables and
   AbortControllers.

4. **MCP-native**  
   Built-in tools and MCP tools are treated uniformly as `tools.Tool`.

5. **Learn, not just remember**  
   Static project instructions (`AGENTS.md`) keep team consistency.
   The SOUL layer learns per-user preferences from session transcripts.

---

## Current Status

Phase 1 of the harness is done:

- [x] `pkg/core` — Runner, Safety, ContextManager, ToolRegistry
- [x] `pkg/tools` — `bash`, `read`, `task` tools
- [x] `pkg/event` — agent lifecycle events
- [x] `pkg/agent` — `Agent`/`Pool`/`Registry` interfaces + `LocalAgent` implementation
- [x] `pkg/session` — `Session`/`Store` interfaces + in-memory implementation
- [x] `pkg/storage` — generic `FileStore` + `JSONLSessionStore`
- [x] `pkg/fs` — isolated filesystem interface
- [x] `pkg/sandbox` — sandbox interface + local-process implementation
- [x] `pkg/llm` — normalized LLM provider + mock provider
- [x] `pkg/memory`, `pkg/plan`, `pkg/soul`, `pkg/log` — storage interfaces and file implementations
- [x] `cmd/dew` — cobra CLI with `run`, `chat`, `session`, `agent`, `version`

In progress / next:

- [ ] TUI (`dew tui`)
- [ ] gRPC service (`dew server`)
- [ ] MCP server
- [ ] Real LLM providers (OpenAI, Anthropic, etc.)
- [ ] Persistent session resume
- [ ] SOUL observation/reflection loop
- [ ] Eval harness

---

## Quick Start

```bash
# Build
go build ./...

# Run tests
go test ./...

# Run a single prompt
dew run "Say hello"

# Interactive chat
dew chat

# Version
dew version
```

Example output:

```
$ dew run "Say hello"
[start] agent=session-...
[turn] 1
[text] I'll run a command for you.
[tool-start] bash(call-1)
[tool-result] bash: hello from dew
[/turn] 1
[turn] 2
[text] Done.
[/turn] 2
[end] agent=session-...
```

---

## Development

```bash
go build ./...
go test ./...
go vet ./...
go run ./cmd/dew --help
```

---

## License

MIT
