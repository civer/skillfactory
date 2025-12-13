package categories

import (
	"fmt"

	"habitwire/client"

	"github.com/spf13/cobra"
)

// RegisterCommands creates and returns the categories command group
func RegisterCommands(c *client.Client, printJSON func(interface{}) error) *cobra.Command {
	service := NewService(c)

	cmd := &cobra.Command{
		Use:   "categories",
		Short: "Manage categories",
	}

	// list
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all categories",
		RunE: func(cmd *cobra.Command, args []string) error {
			categories, err := service.List()
			if err != nil {
				return err
			}
			return printJSON(ToLeanSlice(categories))
		},
	}

	// get
	getCmd := &cobra.Command{
		Use:   "get [id]",
		Short: "Get a category by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			category, err := service.Get(args[0])
			if err != nil {
				return err
			}
			return printJSON(category.ToLean())
		},
	}

	// create
	var createName string
	var createIcon string
	var createColor string
	var createSortOrder int
	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new category",
		RunE: func(cmd *cobra.Command, args []string) error {
			if createName == "" {
				return fmt.Errorf("--name is required")
			}
			req := CreateCategoryRequest{
				Name:  createName,
				Icon:  createIcon,
				Color: createColor,
			}
			if cmd.Flags().Changed("sort-order") {
				req.SortOrder = createSortOrder
			}
			category, err := service.Create(req)
			if err != nil {
				return err
			}
			return printJSON(category.ToLean())
		},
	}
	createCmd.Flags().StringVarP(&createName, "name", "n", "", "Category name (required)")
	createCmd.Flags().StringVar(&createIcon, "icon", "", "Lucide icon name")
	createCmd.Flags().StringVar(&createColor, "color", "", "Color (hex or name)")
	createCmd.Flags().IntVar(&createSortOrder, "sort-order", 0, "Sort order")

	// update
	var updateName string
	var updateIcon string
	var updateColor string
	var updateSortOrder int
	updateCmd := &cobra.Command{
		Use:   "update [id]",
		Short: "Update a category",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			req := UpdateCategoryRequest{}
			if updateName != "" {
				req.Name = updateName
			}
			if updateIcon != "" {
				req.Icon = updateIcon
			}
			if updateColor != "" {
				req.Color = updateColor
			}
			if cmd.Flags().Changed("sort-order") {
				req.SortOrder = &updateSortOrder
			}
			category, err := service.Update(args[0], req)
			if err != nil {
				return err
			}
			return printJSON(category.ToLean())
		},
	}
	updateCmd.Flags().StringVarP(&updateName, "name", "n", "", "New name")
	updateCmd.Flags().StringVar(&updateIcon, "icon", "", "New icon")
	updateCmd.Flags().StringVar(&updateColor, "color", "", "New color")
	updateCmd.Flags().IntVar(&updateSortOrder, "sort-order", 0, "New sort order")

	// delete
	deleteCmd := &cobra.Command{
		Use:   "delete [id]",
		Short: "Delete a category",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := service.Delete(args[0]); err != nil {
				return err
			}
			return printJSON(map[string]bool{"deleted": true})
		},
	}

	cmd.AddCommand(listCmd, getCmd, createCmd, updateCmd, deleteCmd)
	return cmd
}
