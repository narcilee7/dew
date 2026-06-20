# dew

> Go-native Agent Harness

**dew** is a Go-native harness for coding agents. It is not a graph workflow
framework and it is not a single chatbot CLI. It is a *runtime* that provides
isolated boundaries, a plugin hook system, and a minimal default agent loop.

Upper layers — skills, custom agent loops, multi-agent patterns, memory, plans,
and personality — are built *on top of* the harness, not wired into it.

---

## Vision

Build a **small core, rich plugins** agent harness:

- **Harness, not framework**: core provides boundaries, events, safety, and
  hooks. It does not own your agent's thought structure.
- **Plugin hooks over graph nodes**: extensibility comes from lifecycle hooks,
  not from a DAG or state-graph DSL.
- **Skills as the unit of capability**: behavior is packaged as self-contained
  skills (prompts + tools + conventions) that the harness loads.
- **Interface-first**: every isolatable resource is an interface. Callers do not
  know whether a capability runs in-process, cross-process, or remote.
- **Event-sourced**: the agent loop emits a stream of lifecycle events. Sessions,
  audit logs, UI updates, and replay all consume the same event stream.
- **Multi-surface**: one runtime, many surfaces (CLI / TUI / gRPC / MCP / A2A).
- **Terminal-first, batteries included**: the default CLI ships with useful
  built-in skills, but everything can be replaced.

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                         Surfaces                                    │
│   CLI (cobra)   TUI (bubbletea)   gRPC   MCP Server   Library       │
└─────────────────────────────┬───────────────────────────────────────┘
                              │
┌─────────────────────────────▼───────────────────────────────────────┐
│                      dew-core (Harness)                             │
│   Boundaries │ Event Bus │ Safety │ Plugin Registry │ Default Loop │
└─────────────────────────────┬───────────────────────────────────────┘
                              │
        ┌─────────────────────┼─────────────────────┐
        ▼                     ▼                     ▼
   ┌─────────┐          ┌──────────┐          ┌──────────┐
   │  Skills │          │ Plugins  │          │  Agents  │
   │ .dew/   │          │ (hooks)  │          │ (custom  │
   │ skills/ │          │          │          │  loops)  │
   └─────────┘          └──────────┘          └──────────┘
                              │
        ┌─────────────────────┼─────────────────────┐
        ▼                     ▼                     ▼
   ┌─────────┐          ┌──────────┐          ┌──────────┐
   │  dew-fs │          │dew-session│          │dew-sandbox│
   │ dew-llm │          │dew-tools  │          │dew-storage│
   └─────────┘          └──────────┘          └──────────┘
```

Cross-cutting layers:

- **dew-event**: agent lifecycle events (`TurnStart`, `ToolResult`, `Error`, …).
- **dew-storage**: generic file-backed stores on top of `fs.FileSystem`.
- **dew-log**: structured audit logging.
- **dew-trajectory**: model calls, agent traces, and replay data.

Upper-layer capabilities (skills/plugins):

- **dew-memory**: cross-session knowledge.
- **dew-plan**: task plans and checkpoints.
- **dew-soul**: persistent agent self-model.

See [docs/design/](docs/design/) for detailed design documents, especially
[harness.md](docs/design/harness.md) and [blueprint.md](docs/design/blueprint.md).

---

## Design Philosophy

1. **Harness, not framework**  
   dew core provides boundaries, events, safety, and hooks. It does not own the
   agent's decision structure or force a graph/workflow model.

2. **Plugin hooks over graph nodes**  
   Extensibility comes from lifecycle hooks (`BeforeTurn`, `AfterToolUse`, …),
   not from a DAG or state-graph DSL.

3. **Skills as the unit of capability**  
   A skill is a self-contained package of prompts, tools, and conventions. Most
   user-facing behavior is a skill.

4. **Isolation as Interface**  
   FileSystem, Session, Sandbox, Provider are explicit interfaces. Upper layers
   access resources only through boundaries.

5. **Event-driven + Channel-native**  
   Go `chan event.Event` and `context.Context` replace async iterables and
   AbortControllers.

6. **MCP-native**  
   Built-in tools and MCP tools are treated uniformly as `tools.Tool`.

7. **Terminal-first, batteries included**  
   The default experience ships with useful built-in skills, but every part can
   be replaced or disabled.

---

## Current Status

Phase 1 skeleton is done:

- [x] `pkg/core` — Harness + Plugin Hook system + DefaultLoop
- [x] `pkg/tools` — `bash`, `read`, `task` tools
- [x] `pkg/event` — agent lifecycle events
- [x] `pkg/agent` — `Agent`/`Pool`/`Registry` interfaces + `LocalAgent`
- [x] `pkg/session` — `Session`/`Store` interfaces + in-memory implementation
- [x] `pkg/storage` — generic `FileStore` + `JSONLSessionStore`
- [x] `pkg/fs` — isolated filesystem interface
- [x] `pkg/sandbox` — sandbox interface + local-process implementation
- [x] `pkg/llm` — normalized LLM provider + mock provider (to be unified under `llm.Model` capabilities)
- [x] `pkg/config`, `pkg/log` — config and audit logging skeletons
- [x] `pkg/memory`, `pkg/plan`, `pkg/soul` — storage skeletons (to become plugins/skills)
- [x] `cmd/dew` — cobra CLI with `run`, `chat`, `session`, `agent`, `version`

Next (revised priorities):

- [x] Refactor `pkg/core` into a Harness with plugin hooks
- [ ] Define and load the Skill package spec (`SKILL.md` + `manifest.toml`)
- [ ] Simplify `pkg/plan`, `pkg/memory`, `pkg/soul` into storage-backed plugins
- [ ] TUI (`dew tui`) with componentized bubbletea models
- [ ] Unified `llm.Model` + real LLM providers (OpenAI, Anthropic, etc.)
- [ ] Persistent session resume via event replay
- [ ] Trajectory storage (`pkg/trajectory`)
- [ ] gRPC / MCP server surfaces (worker mode)
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
