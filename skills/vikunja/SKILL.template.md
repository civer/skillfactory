# Vikunja Skill

**USE THIS SKILL** when the user asks to:
- Create, update, delete, or list tasks/todos/Aufgaben
- Manage projects or labels in Vikunja
- Mark tasks as done
- Any task management operations

This skill provides direct CLI access to Vikunja - prefer this over MCP for lower token usage.

Base directory: {{SKILL_PATH}}

## Overview

Task management via Vikunja API with lean JSON output (~95% fewer tokens than MCP integration).

## Project IDs

| ID | Name |
|----|------|
| 1 | Backlog |
| 2 | Aktiv |

## Label IDs (GTD-Kontexte)

| ID | Name |
|----|------|
| 2 | @computer |
| 3 | @energie-hoch |
| 4 | @telefon |
| 5 | @draussen |
| 6 | @energie-niedrig |

## Notes

- All responses are lean JSON with minimal overhead
- Date format for `--due`, `--start`, `--end`: `YYYY-MM-DD`
- Priority: 0 (none) to 5 (highest)
- Labels: Use `--labels 1,2,3` on create, or `add-label`/`remove-label` commands

## Commands

{{COMMANDS}}
