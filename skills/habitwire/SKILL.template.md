# HabitWire Skill

**USE THIS SKILL** when the user asks to:
- Create, update, or manage habits
- Track habit check-ins or completions
- View habit statistics or streaks
- Manage habit categories

Base directory: {{SKILL_PATH}}

## Overview
Habit tracking via HabitWire API with lean JSON output.

---

## Business Logic Reference

### Habit Types
| Type | Description | Completion Rule |
|------|-------------|-----------------|
| **SIMPLE** | Binary habits (done/not done) | Any check-in counts as completed |
| **TARGET** | Goal-based with target value | Check-in value must be >= target_value |

### Frequency Types
| Type | Unit | active_days Required | Description |
|------|------|---------------------|-------------|
| **DAILY** | Days | No | Must check in every day |
| **WEEKLY** | Weeks | Yes | Must complete all active days in a week |
| **CUSTOM** | Weeks | Yes | Same as WEEKLY, custom day selection |

### active_days Array
Days of week: `[0,1,2,3,4,5,6]` where 0=Sunday, 6=Saturday
- Example: `[1,3,5]` = Monday, Wednesday, Friday
- Required for WEEKLY and CUSTOM frequency types

### Streak Calculation Rules

**DAILY habits:**
- Counts consecutive days backward from today
- Grace period: Today can be unchecked without breaking streak
- Past days must all have check-ins

**WEEKLY/CUSTOM habits:**
- Counts consecutive WEEKS (not days)
- Week is COMPLETED when ALL active days have check-ins
- Current week has grace period if past active days are done
- Week boundaries depend on user's weekStartsOn setting (0=Sun, 1=Mon)

**TARGET habits:**
- Each check-in must have `value >= target_value` to count
- Values below target do NOT count toward streak

**Skipped check-ins:**
- User setting `skippedBreaksStreak` controls behavior
- Default (false): Skipped counts as completed
- If true: Skipped breaks the streak

### Stats Response Fields
- `current_streak`: Current consecutive streak (days for DAILY, weeks for WEEKLY/CUSTOM)
- `longest_streak`: Best streak ever achieved
- `completion_rate`: (total_checkins / expected_days) * 100
- `total_checkins`: Number of completed check-ins

### Date Format
All dates use `YYYY-MM-DD` format (e.g., `2025-01-15`)

---

## Important: TARGET Habit Tracking Workflow

**CRITICAL for TARGET habits:** When tracking incremental values (e.g., "log 500ml water"), you MUST first fetch the current day's value to avoid overwriting existing data.

**Wrong approach:**
```
User: "I drank another 500ml water"
Agent: habitwire habits check <id> --value 500  # WRONG! Overwrites existing value
```

**Correct approach:**
```
User: "I drank another 500ml water"
Agent:
1. habitwire habits checkins <id> --from 2025-01-15 --to 2025-01-15  # Get today's value
2. If existing value (e.g., 250ml): habitwire habits check <id> --value 750  # Add to existing
   If no value: habitwire habits check <id> --value 500
```

The `check` command performs an **upsert** - it creates a new check-in or updates the existing one for that date. Always read first, then write the accumulated value.

---

## Commands
{{COMMANDS}}
