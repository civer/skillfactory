// Package habits provides habit operations for HabitWire API
package habits

import (
	"encoding/json"
	"fmt"
	"net/url"

	"habitwire/client"
)

// Service handles habit operations
type Service struct {
	client *client.Client
}

// NewService creates a new habit service
func NewService(c *client.Client) *Service {
	return &Service{client: c}
}

// List retrieves all active habits
func (s *Service) List(categoryID string) ([]Habit, error) {
	endpoint := "/habits"
	if categoryID != "" {
		endpoint += "?category=" + url.QueryEscape(categoryID)
	}

	data, err := s.client.Get(endpoint)
	if err != nil {
		return nil, err
	}

	var habits []Habit
	if err := json.Unmarshal(data, &habits); err != nil {
		return nil, fmt.Errorf("failed to parse habits: %w", err)
	}

	return habits, nil
}

// ListAll retrieves all habits including archived
func (s *Service) ListAll(categoryID string) ([]Habit, error) {
	endpoint := "/habits/all"
	if categoryID != "" {
		endpoint += "?category=" + url.QueryEscape(categoryID)
	}

	data, err := s.client.Get(endpoint)
	if err != nil {
		return nil, err
	}

	var habits []Habit
	if err := json.Unmarshal(data, &habits); err != nil {
		return nil, fmt.Errorf("failed to parse habits: %w", err)
	}

	return habits, nil
}

// ListArchived retrieves only archived habits
func (s *Service) ListArchived() ([]Habit, error) {
	data, err := s.client.Get("/habits/archived")
	if err != nil {
		return nil, err
	}

	var habits []Habit
	if err := json.Unmarshal(data, &habits); err != nil {
		return nil, fmt.Errorf("failed to parse habits: %w", err)
	}

	return habits, nil
}

// Get retrieves a single habit by ID
func (s *Service) Get(habitID string) (*Habit, error) {
	endpoint := fmt.Sprintf("/habits/%s", habitID)

	data, err := s.client.Get(endpoint)
	if err != nil {
		return nil, err
	}

	var habit Habit
	if err := json.Unmarshal(data, &habit); err != nil {
		return nil, fmt.Errorf("failed to parse habit: %w", err)
	}

	return &habit, nil
}

// Create creates a new habit
func (s *Service) Create(req CreateHabitRequest) (*Habit, error) {
	data, err := s.client.Post("/habits", req)
	if err != nil {
		return nil, err
	}

	var habit Habit
	if err := json.Unmarshal(data, &habit); err != nil {
		return nil, fmt.Errorf("failed to parse created habit: %w", err)
	}

	return &habit, nil
}

// Update updates an existing habit
func (s *Service) Update(habitID string, req UpdateHabitRequest) (*Habit, error) {
	endpoint := fmt.Sprintf("/habits/%s", habitID)

	data, err := s.client.Put(endpoint, req)
	if err != nil {
		return nil, err
	}

	var habit Habit
	if err := json.Unmarshal(data, &habit); err != nil {
		return nil, fmt.Errorf("failed to parse updated habit: %w", err)
	}

	return &habit, nil
}

// Delete archives a habit (soft delete)
func (s *Service) Delete(habitID string) error {
	endpoint := fmt.Sprintf("/habits/%s", habitID)
	_, err := s.client.Delete(endpoint)
	return err
}

// GetStats retrieves statistics for a habit
func (s *Service) GetStats(habitID string) (*HabitStats, error) {
	endpoint := fmt.Sprintf("/habits/%s/stats", habitID)

	data, err := s.client.Get(endpoint)
	if err != nil {
		return nil, err
	}

	var stats HabitStats
	if err := json.Unmarshal(data, &stats); err != nil {
		return nil, fmt.Errorf("failed to parse stats: %w", err)
	}

	return &stats, nil
}

// Check records a check-in for a habit
func (s *Service) Check(habitID string, req CheckRequest) (*CheckIn, error) {
	endpoint := fmt.Sprintf("/habits/%s/check", habitID)

	data, err := s.client.Post(endpoint, req)
	if err != nil {
		return nil, err
	}

	var checkin CheckIn
	if err := json.Unmarshal(data, &checkin); err != nil {
		return nil, fmt.Errorf("failed to parse check-in: %w", err)
	}

	return &checkin, nil
}

// Uncheck removes a check-in for a habit
func (s *Service) Uncheck(habitID string, req UncheckRequest) error {
	endpoint := fmt.Sprintf("/habits/%s/uncheck", habitID)
	_, err := s.client.Post(endpoint, req)
	return err
}

// Skip marks a habit as skipped for a date
func (s *Service) Skip(habitID string, req SkipRequest) (*CheckIn, error) {
	endpoint := fmt.Sprintf("/habits/%s/skip", habitID)

	data, err := s.client.Post(endpoint, req)
	if err != nil {
		return nil, err
	}

	var checkin CheckIn
	if err := json.Unmarshal(data, &checkin); err != nil {
		return nil, fmt.Errorf("failed to parse skip response: %w", err)
	}

	return &checkin, nil
}

// GetCheckIns retrieves check-ins for a habit within a date range
func (s *Service) GetCheckIns(habitID string, from, to string) ([]CheckIn, error) {
	endpoint := fmt.Sprintf("/habits/%s/checkins", habitID)

	params := url.Values{}
	if from != "" {
		params.Set("from", from)
	}
	if to != "" {
		params.Set("to", to)
	}
	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}

	data, err := s.client.Get(endpoint)
	if err != nil {
		return nil, err
	}

	var checkins []CheckIn
	if err := json.Unmarshal(data, &checkins); err != nil {
		return nil, fmt.Errorf("failed to parse check-ins: %w", err)
	}

	return checkins, nil
}
