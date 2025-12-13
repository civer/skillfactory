# HabitWire Skill

CLI skill for habit tracking via the HabitWire API.

## Setup

1. Deploy this skill using SkillFactory
2. Configure `HABITWIRE_URL` and `HABITWIRE_API_KEY`

## Important: Configure CLAUDE.md

For optimal AI agent behavior, add the following section to your `CLAUDE.md` file. This ensures the agent reads critical business logic (e.g., TARGET habits require read-before-write) before performing operations.

```markdown
## Habit Tracking (HabitWire)

**Skill:** `habitwire` - Habit management via API

**MANDATORY:** Before any mutating operations (check, skip, create, update, delete), ALWAYS read `.claude/skills/habitwire/SKILL.md` first! Contains critical business logic (e.g., TARGET habits: Read-before-Write).

### Proactive Pattern Recognition
When user statements sound like habit tracking, **automatically check** if a matching habit exists:

**Trigger Patterns:**
- Quantities + units: "drank 750ml", "ran 5km", "meditated 30 minutes"
- "I did today..." + trackable activity
- Tracking keywords: drank, ran, trained, meditated, read, slept

**Workflow:**
1. Pattern detected → Run `habits list` (silently, no output to user)
2. Match found → Check-in with extracted value
3. No match → Respond normally (don't create habits without explicit request)

**Examples:**
| User says | Action |
|-----------|--------|
| "Drank 750ml" | → Check "Drink water" with value=750 |
| "Went for a 5km run today" | → Check "Running" with value=5 (if exists) |
| "No workout today" | → Set skip marker (if habit exists) |
```

## Commands Overview

```
habitwire categories list|get|create|update|delete
habitwire habits list|list-all|list-archived|get|create|update|delete
habitwire habits stats <id>
habitwire habits check|uncheck|skip|checkins <id>
```

See `SKILL.md` after deployment for full command documentation and business logic reference.
