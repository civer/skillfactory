# SkillFactory

**Build real programs as Claude Code Skills - not just templates.**

SkillFactory is a Go binary generator with a Skill wrapper TUI. It lets you create **compiled CLI tools** that Claude Code can use as Skills, replacing complex MCP servers with lean, fast binaries.

## Features

- **Go Binary Generator** - Compile full Go programs into Skills
- **MCP Alternative** - Replace heavy MCP servers with lightweight binaries
- **~95% Token Reduction** - Lean JSON output vs. verbose MCP responses
- **TUI for Configuration** - Interactive setup of API keys, URLs, secrets
- **Auto-Generated SKILL.md** - Documentation Claude can discover
- **Community Skills Library** - Extensible `skills/` folder with ready-to-use integrations
- **Zero Runtime Dependencies** - Single binary, no servers to run

## Why SkillFactory?

### The Problem with MCP

MCP servers are powerful but come with overhead:
- Verbose JSON-RPC responses consume tokens
- Require running server processes
- Complex setup and configuration
- Limited to what the protocol supports

### The Problem with Template Skills

Traditional Skills are often just prompt templates:
- No real program logic
- Can only guide, not execute
- Still rely on Claude for everything

### The SkillFactory Solution

**Your Skill IS a program.** A compiled Go binary that:

```
┌─────────────────────────────────────────────────────────────┐
│                    Go Program (Your Skill)                  │
│                                                             │
│   • HTTP clients for any API                                │
│   • Complex business logic                                  │
│   • Data transformation & filtering                         │
│   • Any Go package from the ecosystem                       │
│   • Native speed execution                                  │
│                                                             │
└─────────────────────────────────────────────────────────────┘
                            │
                    SkillFactory TUI
              (Configure → Build → Deploy)
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│              Claude Code Skill Output                       │
│                                                             │
│   ~/.claude/skills/vikunja/                                 │
│   ├── bin/vikunja      ← Compiled binary                   │
│   ├── .env             ← Your configuration                │
│   └── SKILL.md         ← Auto-generated docs               │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### MCP vs. SkillFactory Comparison

| | MCP Server | SkillFactory Skill |
|---|---|---|
| **Token Usage** | High (full JSON-RPC) | ~95% less (lean JSON) |
| **Architecture** | Running server process | Single binary |
| **Logic Location** | Server + Claude context | Compiled in binary |
| **Setup** | Install, configure, run server | Deploy once, done |
| **Capabilities** | Protocol-limited | Unlimited (it's Go) |
| **Speed** | Network + JSON-RPC | Native execution |

## Quick Start

```bash
# Clone
git clone git@github.com:Civer/skillfactory.git
cd skillfactory

# Build
go build -ldflags "-X main.version=$(git describe --tags --always)" -o skillfactory ./cmd/skillfactory

# Run TUI
./skillfactory
```

The TUI guides you through:
1. **Select** a skill from the library
2. **Configure** environment variables (API keys, URLs)
3. **Deploy** to your Claude Code skills folder

## Skills Library

The `skills/` folder is a **community-extensible library**. Each skill is a complete Go CLI application.

### Example Skill: Vikunja

The repository includes **Vikunja** as the reference implementation - a full-featured task management skill:

```
skills/vikunja/
├── skill.yaml           # Manifest with variables
├── main.go              # Cobra CLI entry point
├── client/              # HTTP client with auth
├── tasks/               # Task CRUD + labels
├── labels/              # Label management
└── projects/            # Project management
```

This skill demonstrates all key concepts:
- **Lean JSON output** - Returns only essential fields (~95% token reduction)
- **Cobra CLI** - Professional command structure with subcommands
- **Environment config** - API URL and token via `.env`
- **Multiple entities** - Tasks, projects, labels with full CRUD

Use it as a template when building your own skills!

### Available Skills

| Skill | Description |
|-------|-------------|
| `vikunja` | Task management via Vikunja API (reference implementation) |

*Contributions welcome! See [SKILL_DEVELOPMENT.md](./SKILL_DEVELOPMENT.md)*

## How It Works

### 1. Write a Go CLI

Use Cobra or any CLI framework. Implement commands that output lean JSON:

```go
// Instead of returning 50 fields from the API
type TaskLean struct {
    ID    int64  `json:"id"`
    Title string `json:"title"`
    Done  bool   `json:"done"`
}
```

### 2. Create skill.yaml

```yaml
name: my-skill
description: What this skill does
version: 1.0.0

variables:
  - name: API_URL
    type: string
    required: true
  - name: API_TOKEN
    type: secret
    required: true

build:
  entry: "."
  binary: my-skill
```

### 3. Deploy with TUI

```bash
./skillfactory
# Select skill → Configure → Deploy
```

### 4. Claude Uses It

Claude discovers the skill via `SKILL.md` and calls the binary directly:

```bash
vikunja tasks list --project 2
# {"tasks":[{"id":1,"title":"Review PR","done":false}]}
```

## Build Commands

```bash
# Build TUI (with version)
go build -ldflags "-X main.version=$(git describe --tags --always)" -o skillfactory ./cmd/skillfactory

# Build skill directly (development)
cd skills/vikunja && go build -o vikunja .

# Check version
./skillfactory --version
```

## Documentation

- [SKILL_DEVELOPMENT.md](./SKILL_DEVELOPMENT.md) - Guide for creating your own Skills

## Contributing

The `skills/` library thrives on community contributions. Whether it's a Notion integration, a GitHub client, or something entirely new - if it's a Go CLI that outputs lean JSON, it can be a SkillFactory skill.

## License

GPL-3.0
