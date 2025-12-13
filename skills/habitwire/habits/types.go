// Package habits provides habit-related types and operations for HabitWire API
package habits

// CheckIn represents a habit check-in
type CheckIn struct {
	ID         string   `json:"id,omitempty"`
	HabitID    string   `json:"habit_id,omitempty"`
	Date       string   `json:"date"`
	Value      *float64 `json:"value,omitempty"`
	Notes      string   `json:"notes,omitempty"`
	Skipped    bool     `json:"skipped,omitempty"`
	SkipReason string   `json:"skip_reason,omitempty"`
	CreatedAt  string   `json:"created_at,omitempty"`
}

// CheckInLean represents lean check-in output for CLI
type CheckInLean struct {
	Date    string   `json:"date"`
	Value   *float64 `json:"value,omitempty"`
	Skipped bool     `json:"skipped,omitempty"`
}

// Habit represents a HabitWire habit
type Habit struct {
	ID               string    `json:"id,omitempty"`
	Title            string    `json:"title"`
	Description      string    `json:"description,omitempty"`
	HabitType        string    `json:"habit_type,omitempty"`
	FrequencyType    string    `json:"frequency_type,omitempty"`
	FrequencyValue   int       `json:"frequency_value,omitempty"`
	ActiveDays       []int     `json:"active_days,omitempty"`
	TargetValue      *float64  `json:"target_value,omitempty"`
	DefaultIncrement *float64  `json:"default_increment,omitempty"`
	Unit             string    `json:"unit,omitempty"`
	CategoryID       *string   `json:"category_id,omitempty"`
	Icon             string    `json:"icon,omitempty"`
	ArchivedAt       *string   `json:"archived_at,omitempty"`
	CreatedAt        string    `json:"created_at,omitempty"`
	UpdatedAt        string    `json:"updated_at,omitempty"`
	CheckIns         []CheckIn `json:"checkins,omitempty"`
}

// HabitLean represents lean habit output for CLI
type HabitLean struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	HabitType   string   `json:"habit_type"`
	CategoryID  *string  `json:"category_id,omitempty"`
	Icon        string   `json:"icon,omitempty"`
	TargetValue *float64 `json:"target_value,omitempty"`
	Unit        string   `json:"unit,omitempty"`
}

// HabitStats represents habit statistics
type HabitStats struct {
	CurrentStreak  int     `json:"current_streak"`
	LongestStreak  int     `json:"longest_streak"`
	CompletionRate float64 `json:"completion_rate"`
	TotalCheckins  int     `json:"total_checkins"`
}

// CreateHabitRequest represents a habit creation request
type CreateHabitRequest struct {
	Title            string   `json:"title"`
	Description      string   `json:"description,omitempty"`
	HabitType        string   `json:"habit_type,omitempty"`
	FrequencyType    string   `json:"frequency_type"`
	FrequencyValue   int      `json:"frequency_value,omitempty"`
	ActiveDays       []int    `json:"active_days,omitempty"`
	TargetValue      *float64 `json:"target_value,omitempty"`
	DefaultIncrement *float64 `json:"default_increment,omitempty"`
	Unit             string   `json:"unit,omitempty"`
	CategoryID       *string  `json:"category_id,omitempty"`
	Icon             string   `json:"icon,omitempty"`
}

// UpdateHabitRequest represents fields to update on a habit
type UpdateHabitRequest struct {
	Title            string   `json:"title,omitempty"`
	Description      *string  `json:"description,omitempty"`
	HabitType        string   `json:"habit_type,omitempty"`
	FrequencyType    string   `json:"frequency_type,omitempty"`
	FrequencyValue   *int     `json:"frequency_value,omitempty"`
	ActiveDays       []int    `json:"active_days,omitempty"`
	TargetValue      *float64 `json:"target_value,omitempty"`
	DefaultIncrement *float64 `json:"default_increment,omitempty"`
	Unit             *string  `json:"unit,omitempty"`
	CategoryID       *string  `json:"category_id,omitempty"`
	Icon             *string  `json:"icon,omitempty"`
}

// CheckRequest represents a check-in request
type CheckRequest struct {
	Date  string   `json:"date,omitempty"`
	Value *float64 `json:"value,omitempty"`
	Notes string   `json:"notes,omitempty"`
}

// UncheckRequest represents an uncheck request
type UncheckRequest struct {
	Date string `json:"date,omitempty"`
}

// SkipRequest represents a skip request
type SkipRequest struct {
	Date   string `json:"date,omitempty"`
	Reason string `json:"reason,omitempty"`
}

// ToLean converts a full Habit to lean output
func (h *Habit) ToLean() HabitLean {
	return HabitLean{
		ID:          h.ID,
		Title:       h.Title,
		HabitType:   h.HabitType,
		CategoryID:  h.CategoryID,
		Icon:        h.Icon,
		TargetValue: h.TargetValue,
		Unit:        h.Unit,
	}
}

// ToLeanSlice converts a slice of Habits to lean output
func ToLeanSlice(habits []Habit) []HabitLean {
	result := make([]HabitLean, len(habits))
	for i := range habits {
		result[i] = habits[i].ToLean()
	}
	return result
}

// ToLean converts a full CheckIn to lean output
func (c *CheckIn) ToLean() CheckInLean {
	return CheckInLean{
		Date:    c.Date,
		Value:   c.Value,
		Skipped: c.Skipped,
	}
}

// CheckInsToLeanSlice converts a slice of CheckIns to lean output
func CheckInsToLeanSlice(checkins []CheckIn) []CheckInLean {
	result := make([]CheckInLean, len(checkins))
	for i := range checkins {
		result[i] = checkins[i].ToLean()
	}
	return result
}
