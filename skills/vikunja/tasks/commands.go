package tasks

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/petervogelmann/skillfactory/skills/vikunja/client"
	"github.com/spf13/cobra"
)

// RegisterCommands creates and returns the tasks command group
func RegisterCommands(c *client.Client, printJSON func(interface{}) error) *cobra.Command {
	service := NewService(c)

	cmd := &cobra.Command{
		Use:   "tasks",
		Short: "Manage tasks",
	}

	// list
	var listProjectID int64
	var listShowAll bool
	var listFilter string
	var listSearch string
	var listSortBy string
	var listOrderBy string
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List tasks",
		Long: `List tasks with optional filters.

Filter examples (Vikunja filter syntax):
  --filter "priority >= 3"
  --filter "due_date < now"
  --filter "assignees in user1"

See https://vikunja.io/docs/filters for full filter documentation.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := ListOptions{
				ProjectID:   listProjectID,
				IncludeDone: listShowAll,
				Filter:      listFilter,
				Search:      listSearch,
				SortBy:      listSortBy,
				OrderBy:     listOrderBy,
			}
			tasks, err := service.List(opts)
			if err != nil {
				return err
			}
			return printJSON(ToLeanSlice(tasks))
		},
	}
	listCmd.Flags().Int64VarP(&listProjectID, "project", "p", 0, "Filter by project ID")
	listCmd.Flags().BoolVar(&listShowAll, "all", false, "Include done tasks")
	listCmd.Flags().StringVarP(&listFilter, "filter", "f", "", "Vikunja filter query (e.g., \"priority >= 3\")")
	listCmd.Flags().StringVarP(&listSearch, "search", "s", "", "Search in task text")
	listCmd.Flags().StringVar(&listSortBy, "sort", "", "Sort by field (id, title, due_date, priority, created, updated)")
	listCmd.Flags().StringVar(&listOrderBy, "order", "", "Sort order (asc, desc)")

	// get
	getCmd := &cobra.Command{
		Use:   "get [id]",
		Short: "Get a task by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := parseID(args[0])
			if err != nil {
				return err
			}
			task, err := service.Get(id)
			if err != nil {
				return err
			}
			return printJSON(task.ToLean())
		},
	}

	// create
	var createTitle string
	var createDescription string
	var createPriority int
	var createDue string
	var createStart string
	var createEnd string
	var createColor string
	var createFavorite bool
	var createPercent int
	var createLabels string
	var createProjectID int64
	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new task",
		RunE: func(cmd *cobra.Command, args []string) error {
			if createTitle == "" {
				return fmt.Errorf("--title is required")
			}
			if createProjectID == 0 {
				return fmt.Errorf("--project is required")
			}
			req := CreateTaskRequest{
				Title:       createTitle,
				Description: createDescription,
				Priority:    createPriority,
				DueDate:     createDue,
				StartDate:   createStart,
				EndDate:     createEnd,
				HexColor:    createColor,
				PercentDone: float64(createPercent) / 100.0,
			}
			if cmd.Flags().Changed("favorite") {
				req.IsFavorite = &createFavorite
			}
			task, err := service.Create(createProjectID, req)
			if err != nil {
				return err
			}

			// Add labels if specified
			if createLabels != "" {
				labelIDs, err := parseIDList(createLabels)
				if err != nil {
					return fmt.Errorf("invalid labels: %w", err)
				}
				for _, labelID := range labelIDs {
					if err := service.AddLabel(task.ID, labelID); err != nil {
						return fmt.Errorf("failed to add label %d: %w", labelID, err)
					}
				}
				// Refresh task to get labels
				task, err = service.Get(task.ID)
				if err != nil {
					return err
				}
			}

			return printJSON(task.ToLean())
		},
	}
	createCmd.Flags().StringVarP(&createTitle, "title", "t", "", "Task title (required)")
	createCmd.Flags().StringVarP(&createDescription, "description", "d", "", "Task description")
	createCmd.Flags().Int64VarP(&createProjectID, "project", "p", 0, "Project ID (required)")
	createCmd.Flags().IntVar(&createPriority, "priority", 0, "Task priority (0-5)")
	createCmd.Flags().StringVar(&createDue, "due", "", "Due date (YYYY-MM-DD or YYYY-MM-DDTHH:MM)")
	createCmd.Flags().StringVar(&createStart, "start", "", "Start date (YYYY-MM-DD)")
	createCmd.Flags().StringVar(&createEnd, "end", "", "End date (YYYY-MM-DD)")
	createCmd.Flags().StringVar(&createColor, "color", "", "Hex color (e.g., #ff5733)")
	createCmd.Flags().BoolVar(&createFavorite, "favorite", false, "Mark as favorite")
	createCmd.Flags().IntVar(&createPercent, "percent", 0, "Percent done (0-100)")
	createCmd.Flags().StringVar(&createLabels, "labels", "", "Label IDs (comma-separated, e.g., 1,3,5)")

	// done
	doneCmd := &cobra.Command{
		Use:   "done [id]",
		Short: "Mark a task as done",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := parseID(args[0])
			if err != nil {
				return err
			}
			task, err := service.Done(id)
			if err != nil {
				return err
			}
			return printJSON(task.ToLean())
		},
	}

	// update
	var updateID int64
	var updateTitle string
	var updateDescription string
	var updatePriority int
	var updateDue string
	var updateStart string
	var updateEnd string
	var updateColor string
	var updateFavorite bool
	var updateNoFavorite bool
	var updatePercent int
	var updateDone bool
	var updateUndone bool
	var updateProjectID int64
	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "Update a task",
		RunE: func(cmd *cobra.Command, args []string) error {
			if updateID == 0 {
				return fmt.Errorf("--id is required")
			}
			req := UpdateTaskRequest{}
			if updateTitle != "" {
				req.Title = updateTitle
			}
			if cmd.Flags().Changed("description") {
				req.Description = &updateDescription
			}
			if cmd.Flags().Changed("priority") {
				req.Priority = &updatePriority
			}
			if updateDue != "" {
				req.DueDate = updateDue
			}
			if updateStart != "" {
				req.StartDate = updateStart
			}
			if updateEnd != "" {
				req.EndDate = updateEnd
			}
			if updateColor != "" {
				req.HexColor = updateColor
			}
			if updateFavorite {
				fav := true
				req.IsFavorite = &fav
			}
			if updateNoFavorite {
				fav := false
				req.IsFavorite = &fav
			}
			if cmd.Flags().Changed("percent") {
				pct := float64(updatePercent) / 100.0
				req.PercentDone = &pct
			}
			if updateDone {
				done := true
				req.Done = &done
			}
			if updateUndone {
				done := false
				req.Done = &done
			}
			if cmd.Flags().Changed("project") {
				req.ProjectID = &updateProjectID
			}
			task, err := service.Update(updateID, req)
			if err != nil {
				return err
			}
			return printJSON(task.ToLean())
		},
	}
	updateCmd.Flags().Int64Var(&updateID, "id", 0, "Task ID (required)")
	updateCmd.Flags().StringVarP(&updateTitle, "title", "t", "", "New title")
	updateCmd.Flags().StringVarP(&updateDescription, "description", "d", "", "New description")
	updateCmd.Flags().IntVar(&updatePriority, "priority", 0, "New priority (0-5)")
	updateCmd.Flags().StringVar(&updateDue, "due", "", "Due date (YYYY-MM-DD or YYYY-MM-DDTHH:MM)")
	updateCmd.Flags().StringVar(&updateStart, "start", "", "Start date (YYYY-MM-DD)")
	updateCmd.Flags().StringVar(&updateEnd, "end", "", "End date (YYYY-MM-DD)")
	updateCmd.Flags().StringVar(&updateColor, "color", "", "Hex color (e.g., #ff5733)")
	updateCmd.Flags().BoolVar(&updateFavorite, "favorite", false, "Mark as favorite")
	updateCmd.Flags().BoolVar(&updateNoFavorite, "no-favorite", false, "Remove favorite")
	updateCmd.Flags().IntVar(&updatePercent, "percent", 0, "Percent done (0-100)")
	updateCmd.Flags().BoolVar(&updateDone, "done", false, "Mark as done")
	updateCmd.Flags().BoolVar(&updateUndone, "undone", false, "Mark as not done")
	updateCmd.Flags().Int64VarP(&updateProjectID, "project", "p", 0, "Move to project ID")

	// delete
	deleteCmd := &cobra.Command{
		Use:   "delete [id]",
		Short: "Delete a task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := parseID(args[0])
			if err != nil {
				return err
			}
			if err := service.Delete(id); err != nil {
				return err
			}
			return printJSON(map[string]bool{"deleted": true})
		},
	}

	// labels - get labels for a task
	labelsCmd := &cobra.Command{
		Use:   "labels [task-id]",
		Short: "List labels for a task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			taskID, err := parseID(args[0])
			if err != nil {
				return err
			}
			labels, err := service.GetLabels(taskID)
			if err != nil {
				return err
			}
			// Convert to lean format
			result := make([]map[string]interface{}, len(labels))
			for i, l := range labels {
				result[i] = map[string]interface{}{
					"id":    l.ID,
					"title": l.Title,
				}
				if l.HexColor != "" {
					result[i]["color"] = l.HexColor
				}
			}
			return printJSON(result)
		},
	}

	// add-label
	addLabelCmd := &cobra.Command{
		Use:   "add-label [task-id] [label-id]",
		Short: "Add a label to a task",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			taskID, err := parseID(args[0])
			if err != nil {
				return err
			}
			labelID, err := parseID(args[1])
			if err != nil {
				return err
			}
			if err := service.AddLabel(taskID, labelID); err != nil {
				return err
			}
			return printJSON(map[string]interface{}{
				"task_id":  taskID,
				"label_id": labelID,
				"added":    true,
			})
		},
	}

	// remove-label
	removeLabelCmd := &cobra.Command{
		Use:   "remove-label [task-id] [label-id]",
		Short: "Remove a label from a task",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			taskID, err := parseID(args[0])
			if err != nil {
				return err
			}
			labelID, err := parseID(args[1])
			if err != nil {
				return err
			}
			if err := service.RemoveLabel(taskID, labelID); err != nil {
				return err
			}
			return printJSON(map[string]interface{}{
				"task_id":  taskID,
				"label_id": labelID,
				"removed":  true,
			})
		},
	}

	cmd.AddCommand(listCmd, getCmd, createCmd, doneCmd, updateCmd, deleteCmd, labelsCmd, addLabelCmd, removeLabelCmd)
	return cmd
}

func parseID(s string) (int64, error) {
	var id int64
	_, err := fmt.Sscanf(s, "%d", &id)
	if err != nil {
		return 0, fmt.Errorf("invalid ID: %s", s)
	}
	return id, nil
}

// parseIDList parses a comma-separated list of IDs
func parseIDList(s string) ([]int64, error) {
	if s == "" {
		return nil, nil
	}
	parts := strings.Split(s, ",")
	ids := make([]int64, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		id, err := strconv.ParseInt(p, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid ID '%s': %w", p, err)
		}
		ids = append(ids, id)
	}
	return ids, nil
}
