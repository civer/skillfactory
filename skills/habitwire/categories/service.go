// Package categories provides category operations for HabitWire API
package categories

import (
	"encoding/json"
	"fmt"

	"habitwire/client"
)

// Service handles category operations
type Service struct {
	client *client.Client
}

// NewService creates a new category service
func NewService(c *client.Client) *Service {
	return &Service{client: c}
}

// List retrieves all categories
func (s *Service) List() ([]Category, error) {
	data, err := s.client.Get("/categories")
	if err != nil {
		return nil, err
	}

	var categories []Category
	if err := json.Unmarshal(data, &categories); err != nil {
		return nil, fmt.Errorf("failed to parse categories: %w", err)
	}

	return categories, nil
}

// Get retrieves a single category by ID
func (s *Service) Get(categoryID string) (*Category, error) {
	endpoint := fmt.Sprintf("/categories/%s", categoryID)

	data, err := s.client.Get(endpoint)
	if err != nil {
		return nil, err
	}

	var category Category
	if err := json.Unmarshal(data, &category); err != nil {
		return nil, fmt.Errorf("failed to parse category: %w", err)
	}

	return &category, nil
}

// Create creates a new category
func (s *Service) Create(req CreateCategoryRequest) (*Category, error) {
	data, err := s.client.Post("/categories", req)
	if err != nil {
		return nil, err
	}

	var category Category
	if err := json.Unmarshal(data, &category); err != nil {
		return nil, fmt.Errorf("failed to parse created category: %w", err)
	}

	return &category, nil
}

// Update updates an existing category
func (s *Service) Update(categoryID string, req UpdateCategoryRequest) (*Category, error) {
	endpoint := fmt.Sprintf("/categories/%s", categoryID)

	data, err := s.client.Put(endpoint, req)
	if err != nil {
		return nil, err
	}

	var category Category
	if err := json.Unmarshal(data, &category); err != nil {
		return nil, fmt.Errorf("failed to parse updated category: %w", err)
	}

	return &category, nil
}

// Delete deletes a category
func (s *Service) Delete(categoryID string) error {
	endpoint := fmt.Sprintf("/categories/%s", categoryID)
	_, err := s.client.Delete(endpoint)
	return err
}
