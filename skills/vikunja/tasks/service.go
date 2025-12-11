// Package tasks provides task operations for Vikunja API
package tasks

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/petervogelmann/skillfactory/skills/vikunja/client"
)

// formatDate converts date strings to RFC3339 format with local timezone
// Supported formats:
//   - YYYY-MM-DD (defaults to 00:00)
//   - YYYY-MM-DDTHH:MM (e.g., 2025-12-13T12:00)
//   - Full RFC3339 (passed through)
func formatDate(date string) string {
	if date == "" {
		return ""
	}
	// Already complete RFC3339 (contains T and timezone)?
	if strings.Contains(date, "T") && (strings.Contains(date, "Z") || strings.Contains(date, "+") || strings.Contains(date, "-") && strings.Count(date, "-") > 2) {
		return date
	}
	// YYYY-MM-DDTHH:MM format?
	if strings.Contains(date, "T") {
		t, err := time.ParseInLocation("2006-01-02T15:04", date, time.Local)
		if err != nil {
			// Fallback: append seconds and Z
			return date + ":00Z"
		}
		return t.Format(time.RFC3339)
	}
	// YYYY-MM-DD → Parse as local time at 00:00, then output RFC3339
	t, err := time.ParseInLocation("2006-01-02", date, time.Local)
	if err != nil {
		// Fallback: just append T00:00:00Z
		return date + "T00:00:00Z"
	}
	return t.Format(time.RFC3339)
}

// Service handles task operations
type Service struct {
	client *client.Client
}

// NewService creates a new task service
func NewService(c *client.Client) *Service {
	return &Service{client: c}
}

// ListOptions contains options for listing tasks
type ListOptions struct {
	ProjectID  int64
	IncludeDone bool
	Filter     string // Vikunja filter query (e.g., "priority >= 3")
	Search     string // Search in task text
	SortBy     string // Field to sort by (e.g., "due_date", "priority")
	OrderBy    string // "asc" or "desc"
}

// List retrieves tasks with optional filters
func (s *Service) List(opts ListOptions) ([]Task, error) {
	var endpoint string

	if opts.ProjectID > 0 {
		endpoint = fmt.Sprintf("/projects/%d/tasks", opts.ProjectID)
	} else {
		endpoint = "/tasks/all"
	}

	// Build query parameters with proper URL encoding
	params := url.Values{}

	// Build filter
	var filters []string
	if !opts.IncludeDone {
		filters = append(filters, "done = false")
	}
	if opts.Filter != "" {
		filters = append(filters, opts.Filter)
	}
	if len(filters) > 0 {
		params.Set("filter", strings.Join(filters, " && "))
	}

	// Search
	if opts.Search != "" {
		params.Set("s", opts.Search)
	}

	// Sorting
	if opts.SortBy != "" {
		params.Set("sort_by", opts.SortBy)
	}
	if opts.OrderBy != "" {
		params.Set("order_by", opts.OrderBy)
	}

	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}

	data, err := s.client.Get(endpoint)
	if err != nil {
		return nil, err
	}

	var tasks []Task
	if err := json.Unmarshal(data, &tasks); err != nil {
		return nil, fmt.Errorf("failed to parse tasks: %w", err)
	}

	return tasks, nil
}

// Get retrieves a single task by ID
func (s *Service) Get(taskID int64) (*Task, error) {
	endpoint := fmt.Sprintf("/tasks/%d", taskID)

	data, err := s.client.Get(endpoint)
	if err != nil {
		return nil, err
	}

	var task Task
	if err := json.Unmarshal(data, &task); err != nil {
		return nil, fmt.Errorf("failed to parse task: %w", err)
	}

	return &task, nil
}

// Create creates a new task in the specified project
func (s *Service) Create(projectID int64, req CreateTaskRequest) (*Task, error) {
	endpoint := fmt.Sprintf("/projects/%d/tasks", projectID)

	// Format dates to RFC3339
	req.DueDate = formatDate(req.DueDate)
	req.StartDate = formatDate(req.StartDate)
	req.EndDate = formatDate(req.EndDate)

	data, err := s.client.Put(endpoint, req)
	if err != nil {
		return nil, err
	}

	var task Task
	if err := json.Unmarshal(data, &task); err != nil {
		return nil, fmt.Errorf("failed to parse created task: %w", err)
	}

	return &task, nil
}

// Update updates an existing task
// It first fetches the current task, applies changes, then sends the full object
func (s *Service) Update(taskID int64, req UpdateTaskRequest) (*Task, error) {
	// First, get the current task
	task, err := s.Get(taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task for update: %w", err)
	}

	// Apply changes to the task
	if req.Title != "" {
		task.Title = req.Title
	}
	if req.Description != nil {
		task.Description = *req.Description
	}
	if req.Priority != nil {
		task.Priority = *req.Priority
	}
	if req.DueDate != "" {
		task.DueDate = formatDate(req.DueDate)
	}
	if req.StartDate != "" {
		task.StartDate = formatDate(req.StartDate)
	}
	if req.EndDate != "" {
		task.EndDate = formatDate(req.EndDate)
	}
	if req.HexColor != "" {
		task.HexColor = req.HexColor
	}
	if req.IsFavorite != nil {
		task.IsFavorite = *req.IsFavorite
	}
	if req.PercentDone != nil {
		task.PercentDone = *req.PercentDone
	}
	if req.Done != nil {
		task.Done = *req.Done
	}
	if req.ProjectID != nil {
		task.ProjectID = *req.ProjectID
	}

	// Send the full task object
	endpoint := fmt.Sprintf("/tasks/%d", taskID)
	data, err := s.client.Post(endpoint, task)
	if err != nil {
		return nil, err
	}

	var updatedTask Task
	if err := json.Unmarshal(data, &updatedTask); err != nil {
		return nil, fmt.Errorf("failed to parse updated task: %w", err)
	}

	return &updatedTask, nil
}

// Done marks a task as done
func (s *Service) Done(taskID int64) (*Task, error) {
	done := true
	return s.Update(taskID, UpdateTaskRequest{Done: &done})
}

// Delete deletes a task
func (s *Service) Delete(taskID int64) error {
	endpoint := fmt.Sprintf("/tasks/%d", taskID)
	_, err := s.client.Delete(endpoint)
	return err
}

// LabelTaskRequest represents a request to add a label to a task
type LabelTaskRequest struct {
	LabelID int64 `json:"label_id"`
}

// TaskLabel represents a label attached to a task
type TaskLabel struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	HexColor    string `json:"hex_color,omitempty"`
}

// GetLabels retrieves all labels for a task
func (s *Service) GetLabels(taskID int64) ([]TaskLabel, error) {
	endpoint := fmt.Sprintf("/tasks/%d/labels", taskID)

	data, err := s.client.Get(endpoint)
	if err != nil {
		return nil, err
	}

	var labels []TaskLabel
	if err := json.Unmarshal(data, &labels); err != nil {
		return nil, fmt.Errorf("failed to parse labels: %w", err)
	}

	return labels, nil
}

// AddLabel adds a label to a task
func (s *Service) AddLabel(taskID, labelID int64) error {
	endpoint := fmt.Sprintf("/tasks/%d/labels", taskID)

	req := LabelTaskRequest{LabelID: labelID}
	_, err := s.client.Put(endpoint, req)
	return err
}

// RemoveLabel removes a label from a task
func (s *Service) RemoveLabel(taskID, labelID int64) error {
	endpoint := fmt.Sprintf("/tasks/%d/labels/%d", taskID, labelID)
	_, err := s.client.Delete(endpoint)
	return err
}
