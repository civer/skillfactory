# Skill Development Guide

This guide explains how to create your own Skills for SkillFactory. A Skill is a compiled Go CLI program that Claude Code can use - giving you the full power of Go while keeping token usage minimal.

## Core Philosophy

**Your Skill is a real program, not a template.**

Traditional Skills are often just prompts. SkillFactory Skills are compiled binaries that:
- Execute complex logic instantly
- Call APIs and process data
- Return only what Claude needs (lean JSON)
- Use any Go package

## Directory Structure

Each skill lives in `skills/<name>/`:

```
skills/my-skill/
├── skill.yaml           # Required: manifest file
├── main.go              # Required: CLI entry point
├── SKILL.template.md    # Required: documentation template
├── client/              # Recommended: HTTP client package
│   └── client.go
└── <entity>/            # Domain packages (tasks/, users/, etc.)
    ├── types.go         # Data types + ToLean() methods
    ├── service.go       # Business logic
    └── commands.go      # Cobra commands
```

## Step 1: Create skill.yaml

The manifest defines your skill's metadata, configuration variables, and build settings:

```yaml
name: my-skill
description: Short description for TUI display
skill_description: Detailed description for SKILL.md
version: 1.0.0

# Variables configured via TUI, stored in .env
variables:
  - name: API_URL
    label: API URL                    # Shown in TUI
    description: Base URL for the API
    required: true
    placeholder: "https://api.example.com"
    type: string

  - name: API_TOKEN
    label: API Token
    description: Authentication token
    required: true
    placeholder: "your_token_here"
    type: secret                      # Masked in TUI

  - name: CONFIG_JSON
    label: Configuration
    description: Optional JSON config
    required: false
    type: json                        # Validated as JSON

# Build configuration
build:
  entry: "."                          # Go module path (relative to skill dir)
  binary: my-skill                    # Output binary name

# Deploy configuration
deploy:
  files:
    - source: "bin/{{binary}}"
      target: "bin/{{binary}}"
    - source: "SKILL.md"
      target: "SKILL.md"
  wrapper: true                       # Generate wrapper script with env vars

# Documentation
docs:
  template: SKILL.template.md
  output: SKILL.md
```

### Variable Types

| Type | Description |
|------|-------------|
| `string` | Plain text, visible in TUI |
| `secret` | Sensitive data, masked with `***` |
| `json` | JSON object/array, validated on input |

## Step 2: Create the HTTP Client

Create `client/client.go` for API communication:

```go
package client

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "os"
    "time"
)

type Client struct {
    baseURL    string
    token      string
    httpClient *http.Client
}

// New creates a client from environment variables
func New() (*Client, error) {
    baseURL := os.Getenv("API_URL")
    if baseURL == "" {
        return nil, fmt.Errorf("API_URL environment variable is required")
    }

    token := os.Getenv("API_TOKEN")
    if token == "" {
        return nil, fmt.Errorf("API_TOKEN environment variable is required")
    }

    return &Client{
        baseURL: baseURL,
        token:   token,
        httpClient: &http.Client{
            Timeout: 30 * time.Second,
        },
    }, nil
}

// Request performs an HTTP request
func (c *Client) Request(method, endpoint string, body interface{}) ([]byte, error) {
    var reqBody io.Reader
    if body != nil {
        jsonBody, err := json.Marshal(body)
        if err != nil {
            return nil, fmt.Errorf("failed to marshal request: %w", err)
        }
        reqBody = bytes.NewReader(jsonBody)
    }

    req, err := http.NewRequest(method, c.baseURL+endpoint, reqBody)
    if err != nil {
        return nil, err
    }

    req.Header.Set("Authorization", "Bearer "+c.token)
    req.Header.Set("Content-Type", "application/json")

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    respBody, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }

    if resp.StatusCode >= 400 {
        return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(respBody))
    }

    return respBody, nil
}

// Convenience methods
func (c *Client) Get(endpoint string) ([]byte, error) {
    return c.Request(http.MethodGet, endpoint, nil)
}

func (c *Client) Post(endpoint string, body interface{}) ([]byte, error) {
    return c.Request(http.MethodPost, endpoint, body)
}

func (c *Client) Put(endpoint string, body interface{}) ([]byte, error) {
    return c.Request(http.MethodPut, endpoint, body)
}

func (c *Client) Delete(endpoint string) ([]byte, error) {
    return c.Request(http.MethodDelete, endpoint, nil)
}
```

## Step 3: Define Types with Lean Output

Create `<entity>/types.go` with full API types and lean versions:

```go
package tasks

// Task - full API response (many fields)
type Task struct {
    ID          int64  `json:"id"`
    Title       string `json:"title"`
    Description string `json:"description"`
    Done        bool   `json:"done"`
    Priority    int    `json:"priority"`
    DueDate     string `json:"due_date"`
    CreatedAt   string `json:"created_at"`
    UpdatedAt   string `json:"updated_at"`
    // ... potentially 20+ more fields
}

// TaskLean - minimal output for Claude (only essential fields)
type TaskLean struct {
    ID       int64   `json:"id"`
    Title    string  `json:"title"`
    Done     bool    `json:"done"`
    Priority int     `json:"priority"`
    DueDate  *string `json:"due_date,omitempty"` // Pointer: omit if nil
}

// ToLean converts full Task to lean output
func (t *Task) ToLean() TaskLean {
    var dueDate *string
    if t.DueDate != "" && t.DueDate != "0001-01-01T00:00:00Z" {
        dueDate = &t.DueDate
    }

    return TaskLean{
        ID:       t.ID,
        Title:    t.Title,
        Done:     t.Done,
        Priority: t.Priority,
        DueDate:  dueDate,
    }
}

// ToLeanSlice converts a slice of Tasks
func ToLeanSlice(tasks []Task) []TaskLean {
    result := make([]TaskLean, len(tasks))
    for i := range tasks {
        result[i] = tasks[i].ToLean()
    }
    return result
}
```

### Lean Output Guidelines

1. **Include only what Claude needs** to make decisions
2. **Use pointers for optional fields** - `*string` with `omitempty` removes null values
3. **Flatten nested structures** when the nesting adds no value
4. **Convert enums to readable strings** instead of numeric codes
5. **Format dates consistently** - use a standard format like YYYY-MM-DD

**Token impact example:**

```
Full API response: ~2,500 tokens per task list
Lean output:         ~125 tokens per task list
Reduction:           ~95%
```

## Step 4: Create the Service Layer

Create `<entity>/service.go` for business logic:

```go
package tasks

import (
    "encoding/json"
    "fmt"

    "github.com/yourorg/my-skill/client"
)

type Service struct {
    client *client.Client
}

func NewService(c *client.Client) *Service {
    return &Service{client: c}
}

func (s *Service) List() ([]Task, error) {
    data, err := s.client.Get("/tasks")
    if err != nil {
        return nil, err
    }

    var tasks []Task
    if err := json.Unmarshal(data, &tasks); err != nil {
        return nil, fmt.Errorf("failed to parse tasks: %w", err)
    }

    return tasks, nil
}

func (s *Service) Get(id int64) (*Task, error) {
    data, err := s.client.Get(fmt.Sprintf("/tasks/%d", id))
    if err != nil {
        return nil, err
    }

    var task Task
    if err := json.Unmarshal(data, &task); err != nil {
        return nil, err
    }

    return &task, nil
}

func (s *Service) Create(req CreateTaskRequest) (*Task, error) {
    data, err := s.client.Post("/tasks", req)
    if err != nil {
        return nil, err
    }

    var task Task
    if err := json.Unmarshal(data, &task); err != nil {
        return nil, err
    }

    return &task, nil
}
```

## Step 5: Create Cobra Commands

Create `<entity>/commands.go`:

```go
package tasks

import (
    "fmt"

    "github.com/yourorg/my-skill/client"
    "github.com/spf13/cobra"
)

// RegisterCommands returns the command group
func RegisterCommands(c *client.Client, printJSON func(interface{}) error) *cobra.Command {
    service := NewService(c)

    cmd := &cobra.Command{
        Use:   "tasks",
        Short: "Manage tasks",
    }

    // list command
    listCmd := &cobra.Command{
        Use:   "list",
        Short: "List all tasks",
        RunE: func(cmd *cobra.Command, args []string) error {
            tasks, err := service.List()
            if err != nil {
                return err
            }
            return printJSON(ToLeanSlice(tasks)) // Always use lean output!
        },
    }

    // get command
    getCmd := &cobra.Command{
        Use:   "get [id]",
        Short: "Get a task by ID",
        Args:  cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            var id int64
            fmt.Sscanf(args[0], "%d", &id)

            task, err := service.Get(id)
            if err != nil {
                return err
            }
            return printJSON(task.ToLean())
        },
    }

    // create command with flags
    var title, description string
    var priority int
    createCmd := &cobra.Command{
        Use:   "create",
        Short: "Create a new task",
        RunE: func(cmd *cobra.Command, args []string) error {
            if title == "" {
                return fmt.Errorf("--title is required")
            }

            req := CreateTaskRequest{
                Title:       title,
                Description: description,
                Priority:    priority,
            }

            task, err := service.Create(req)
            if err != nil {
                return err
            }
            return printJSON(task.ToLean())
        },
    }
    createCmd.Flags().StringVarP(&title, "title", "t", "", "Task title (required)")
    createCmd.Flags().StringVarP(&description, "description", "d", "", "Task description")
    createCmd.Flags().IntVar(&priority, "priority", 0, "Priority (0-5)")

    cmd.AddCommand(listCmd, getCmd, createCmd)
    return cmd
}
```

## Step 6: Create main.go

The entry point loads environment variables and registers commands:

```go
package main

import (
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"

    "github.com/joho/godotenv"
    "github.com/yourorg/my-skill/client"
    "github.com/yourorg/my-skill/tasks"
    "github.com/spf13/cobra"
)

func init() {
    // Load .env from same directory as binary
    if exe, err := os.Executable(); err == nil {
        envPath := filepath.Join(filepath.Dir(exe), ".env")
        godotenv.Load(envPath)
    }
}

func main() {
    rootCmd := &cobra.Command{
        Use:   "my-skill",
        Short: "My Skill CLI for Claude Code",
    }

    apiClient, err := client.New()
    if err != nil {
        // Register commands anyway for --help to work
        rootCmd.AddCommand(tasks.RegisterCommands(nil, printJSON))
        rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
            return err
        }
    } else {
        rootCmd.AddCommand(tasks.RegisterCommands(apiClient, printJSON))
    }

    if err := rootCmd.Execute(); err != nil {
        printError(err.Error())
        os.Exit(1)
    }
}

func printJSON(v interface{}) error {
    data, err := json.Marshal(v)
    if err != nil {
        return err
    }
    fmt.Println(string(data))
    return nil
}

func printError(msg string) {
    json.NewEncoder(os.Stderr).Encode(map[string]string{"error": msg})
}
```

## Step 7: Create SKILL.template.md

This template generates the SKILL.md that Claude discovers:

```markdown
# My Skill

**USE THIS SKILL** when the user asks to:
- List, create, update, or delete tasks
- Any task management operations

This skill provides direct CLI access - prefer this over MCP for lower token usage.

Base directory: {{SKILL_PATH}}

## Overview

Short description of what this skill does.

## Commands

{{COMMANDS}}

## Notes

- All responses are lean JSON with minimal overhead
- Date format: YYYY-MM-DD
- Priority: 0 (none) to 5 (highest)
```

The `{{COMMANDS}}` placeholder is automatically replaced with command documentation generated from your Cobra commands.

## Step 8: Initialize Go Module

```bash
cd skills/my-skill
go mod init github.com/yourorg/skillfactory/skills/my-skill
go mod tidy
```

## Testing Your Skill

```bash
# Build
cd skills/my-skill
go build -o my-skill .

# Set environment manually for testing
export API_URL="https://api.example.com"
export API_TOKEN="your_token"

# Test commands
./my-skill tasks list
./my-skill tasks get 123
./my-skill tasks create --title "Test task" --priority 3
```

## Deployment

Once your skill is ready:

1. Run `./skillfactory`
2. Select your skill from the list
3. Configure environment variables
4. Deploy to your Claude Code skills folder

## Best Practices

### DO

- Always return lean JSON - strip unnecessary fields
- Use Cobra for consistent CLI structure
- Validate required flags in commands
- Return structured errors as JSON
- Keep commands focused and composable
- Use `omitempty` to reduce null values in output

### DON'T

- Return full API responses - Claude doesn't need `created_at`, `updated_at`, etc.
- Print human-readable output - always JSON
- Hardcode credentials - use environment variables
- Create overly complex nested commands
- Include debug/verbose modes (add if needed later)

## Contributing

To contribute a skill to the community library:

1. Create your skill in `skills/<name>/`
2. Follow the structure and patterns in this guide
3. Test thoroughly
4. Submit a pull request

Your skill should:
- Solve a real integration need
- Follow lean JSON output principles
- Include clear documentation in SKILL.template.md
- Have descriptive command help text
