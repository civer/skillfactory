// Package categories provides category-related types and operations for HabitWire API
package categories

// Category represents a HabitWire category
type Category struct {
	ID        string `json:"id,omitempty"`
	Name      string `json:"name"`
	Icon      string `json:"icon,omitempty"`
	Color     string `json:"color,omitempty"`
	SortOrder int    `json:"sort_order,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

// CategoryLean represents lean category output for CLI
type CategoryLean struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Icon  string `json:"icon,omitempty"`
	Color string `json:"color,omitempty"`
}

// CreateCategoryRequest represents a category creation request
type CreateCategoryRequest struct {
	Name      string `json:"name"`
	Icon      string `json:"icon,omitempty"`
	Color     string `json:"color,omitempty"`
	SortOrder int    `json:"sort_order,omitempty"`
}

// UpdateCategoryRequest represents a category update request
type UpdateCategoryRequest struct {
	Name      string `json:"name,omitempty"`
	Icon      string `json:"icon,omitempty"`
	Color     string `json:"color,omitempty"`
	SortOrder *int   `json:"sort_order,omitempty"`
}

// ToLean converts a full Category to lean output
func (c *Category) ToLean() CategoryLean {
	return CategoryLean{
		ID:    c.ID,
		Name:  c.Name,
		Icon:  c.Icon,
		Color: c.Color,
	}
}

// ToLeanSlice converts a slice of Categories to lean output
func ToLeanSlice(categories []Category) []CategoryLean {
	result := make([]CategoryLean, len(categories))
	for i := range categories {
		result[i] = categories[i].ToLean()
	}
	return result
}
