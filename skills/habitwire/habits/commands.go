package habits

import (
	"fmt"
	"strconv"
	"strings"

	"habitwire/client"

	"github.com/spf13/cobra"
)

// RegisterCommands creates and returns the habits command group
func RegisterCommands(c *client.Client, printJSON func(interface{}) error) *cobra.Command {
	service := NewService(c)

	cmd := &cobra.Command{
		Use:   "habits",
		Short: "Manage habits",
	}

	// list - active habits only
	var listCategory string
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List active habits",
		RunE: func(cmd *cobra.Command, args []string) error {
			habits, err := service.List(listCategory)
			if err != nil {
				return err
			}
			return printJSON(ToLeanSlice(habits))
		},
	}
	listCmd.Flags().StringVarP(&listCategory, "category", "c", "", "Filter by category ID")

	// list-all - including archived
	var listAllCategory string
	listAllCmd := &cobra.Command{
		Use:   "list-all",
		Short: "List all habits including archived",
		RunE: func(cmd *cobra.Command, args []string) error {
			habits, err := service.ListAll(listAllCategory)
			if err != nil {
				return err
			}
			return printJSON(ToLeanSlice(habits))
		},
	}
	listAllCmd.Flags().StringVarP(&listAllCategory, "category", "c", "", "Filter by category ID")

	// list-archived
	listArchivedCmd := &cobra.Command{
		Use:   "list-archived",
		Short: "List only archived habits",
		RunE: func(cmd *cobra.Command, args []string) error {
			habits, err := service.ListArchived()
			if err != nil {
				return err
			}
			return printJSON(ToLeanSlice(habits))
		},
	}

	// get
	getCmd := &cobra.Command{
		Use:   "get [id]",
		Short: "Get a habit by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			habit, err := service.Get(args[0])
			if err != nil {
				return err
			}
			return printJSON(habit.ToLean())
		},
	}

	// create
	var createTitle string
	var createDescription string
	var createType string
	var createFrequency string
	var createFrequencyValue int
	var createActiveDays string
	var createTarget float64
	var createIncrement float64
	var createUnit string
	var createCategory string
	var createIcon string
	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new habit",
		Long: `Create a new habit.

Habit types:
  SIMPLE  - Binary habit (done/not done)
  TARGET  - Goal-based habit with target value

Frequency types:
  DAILY   - Must check in every day
  WEEKLY  - Must complete on active days each week
  CUSTOM  - Same as WEEKLY with custom days

Active days (for WEEKLY/CUSTOM):
  0=Sunday, 1=Monday, 2=Tuesday, 3=Wednesday, 4=Thursday, 5=Friday, 6=Saturday
  Example: --active-days "1,3,5" for Mon/Wed/Fri`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if createTitle == "" {
				return fmt.Errorf("--title is required")
			}
			if createFrequency == "" {
				return fmt.Errorf("--frequency is required")
			}

			req := CreateHabitRequest{
				Title:         createTitle,
				Description:   createDescription,
				FrequencyType: strings.ToUpper(createFrequency),
			}

			if createType != "" {
				req.HabitType = strings.ToUpper(createType)
			}
			if createFrequencyValue > 0 {
				req.FrequencyValue = createFrequencyValue
			}
			if createActiveDays != "" {
				days, err := parseIntList(createActiveDays)
				if err != nil {
					return fmt.Errorf("invalid active-days: %w", err)
				}
				req.ActiveDays = days
			}
			if cmd.Flags().Changed("target") {
				req.TargetValue = &createTarget
			}
			if cmd.Flags().Changed("increment") {
				req.DefaultIncrement = &createIncrement
			}
			if createUnit != "" {
				req.Unit = createUnit
			}
			if createCategory != "" {
				req.CategoryID = &createCategory
			}
			if createIcon != "" {
				req.Icon = createIcon
			}

			habit, err := service.Create(req)
			if err != nil {
				return err
			}
			return printJSON(habit.ToLean())
		},
	}
	createCmd.Flags().StringVarP(&createTitle, "title", "t", "", "Habit title (required)")
	createCmd.Flags().StringVarP(&createDescription, "description", "d", "", "Habit description")
	createCmd.Flags().StringVar(&createType, "type", "", "Habit type: SIMPLE or TARGET (default: SIMPLE)")
	createCmd.Flags().StringVarP(&createFrequency, "frequency", "f", "", "Frequency type: DAILY, WEEKLY, or CUSTOM (required)")
	createCmd.Flags().IntVar(&createFrequencyValue, "frequency-value", 0, "Frequency value (default: 1)")
	createCmd.Flags().StringVar(&createActiveDays, "active-days", "", "Active days for WEEKLY/CUSTOM (comma-separated: 0-6)")
	createCmd.Flags().Float64Var(&createTarget, "target", 0, "Target value for TARGET habits")
	createCmd.Flags().Float64Var(&createIncrement, "increment", 0, "Default increment for TARGET habits")
	createCmd.Flags().StringVar(&createUnit, "unit", "", "Unit for TARGET habits (ml, min, km, etc.)")
	createCmd.Flags().StringVarP(&createCategory, "category", "c", "", "Category ID")
	createCmd.Flags().StringVar(&createIcon, "icon", "", "Lucide icon name")

	// update
	var updateTitle string
	var updateDescription string
	var updateType string
	var updateFrequency string
	var updateFrequencyValue int
	var updateActiveDays string
	var updateTarget float64
	var updateIncrement float64
	var updateUnit string
	var updateCategory string
	var updateIcon string
	updateCmd := &cobra.Command{
		Use:   "update [id]",
		Short: "Update a habit",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			req := UpdateHabitRequest{}

			if updateTitle != "" {
				req.Title = updateTitle
			}
			if cmd.Flags().Changed("description") {
				req.Description = &updateDescription
			}
			if updateType != "" {
				req.HabitType = strings.ToUpper(updateType)
			}
			if updateFrequency != "" {
				req.FrequencyType = strings.ToUpper(updateFrequency)
			}
			if cmd.Flags().Changed("frequency-value") {
				req.FrequencyValue = &updateFrequencyValue
			}
			if cmd.Flags().Changed("active-days") {
				days, err := parseIntList(updateActiveDays)
				if err != nil {
					return fmt.Errorf("invalid active-days: %w", err)
				}
				req.ActiveDays = days
			}
			if cmd.Flags().Changed("target") {
				req.TargetValue = &updateTarget
			}
			if cmd.Flags().Changed("increment") {
				req.DefaultIncrement = &updateIncrement
			}
			if cmd.Flags().Changed("unit") {
				req.Unit = &updateUnit
			}
			if cmd.Flags().Changed("category") {
				req.CategoryID = &updateCategory
			}
			if cmd.Flags().Changed("icon") {
				req.Icon = &updateIcon
			}

			habit, err := service.Update(args[0], req)
			if err != nil {
				return err
			}
			return printJSON(habit.ToLean())
		},
	}
	updateCmd.Flags().StringVarP(&updateTitle, "title", "t", "", "New title")
	updateCmd.Flags().StringVarP(&updateDescription, "description", "d", "", "New description")
	updateCmd.Flags().StringVar(&updateType, "type", "", "New habit type")
	updateCmd.Flags().StringVarP(&updateFrequency, "frequency", "f", "", "New frequency type")
	updateCmd.Flags().IntVar(&updateFrequencyValue, "frequency-value", 0, "New frequency value")
	updateCmd.Flags().StringVar(&updateActiveDays, "active-days", "", "New active days")
	updateCmd.Flags().Float64Var(&updateTarget, "target", 0, "New target value")
	updateCmd.Flags().Float64Var(&updateIncrement, "increment", 0, "New default increment")
	updateCmd.Flags().StringVar(&updateUnit, "unit", "", "New unit")
	updateCmd.Flags().StringVarP(&updateCategory, "category", "c", "", "New category ID")
	updateCmd.Flags().StringVar(&updateIcon, "icon", "", "New icon")

	// delete (archives)
	deleteCmd := &cobra.Command{
		Use:   "delete [id]",
		Short: "Delete (archive) a habit",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := service.Delete(args[0]); err != nil {
				return err
			}
			return printJSON(map[string]bool{"archived": true})
		},
	}

	// stats
	statsCmd := &cobra.Command{
		Use:   "stats [id]",
		Short: "Get habit statistics",
		Long: `Get statistics for a habit including:
  - current_streak: Current consecutive streak (days for DAILY, weeks for WEEKLY/CUSTOM)
  - longest_streak: Best streak ever achieved
  - completion_rate: Percentage of expected check-ins completed
  - total_checkins: Total number of completed check-ins`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			stats, err := service.GetStats(args[0])
			if err != nil {
				return err
			}
			return printJSON(stats)
		},
	}

	// check
	var checkDate string
	var checkValue float64
	var checkNotes string
	checkCmd := &cobra.Command{
		Use:   "check [id]",
		Short: "Record a check-in for a habit",
		Long: `Record a check-in for a habit.

For SIMPLE habits: Just checking in marks it as done.
For TARGET habits: Provide --value to record progress toward target.

Date defaults to today if not specified.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			req := CheckRequest{
				Date:  checkDate,
				Notes: checkNotes,
			}
			if cmd.Flags().Changed("value") {
				req.Value = &checkValue
			}
			checkin, err := service.Check(args[0], req)
			if err != nil {
				return err
			}
			return printJSON(checkin.ToLean())
		},
	}
	checkCmd.Flags().StringVar(&checkDate, "date", "", "Check-in date (YYYY-MM-DD, defaults to today)")
	checkCmd.Flags().Float64VarP(&checkValue, "value", "v", 0, "Value for TARGET habits")
	checkCmd.Flags().StringVarP(&checkNotes, "notes", "n", "", "Optional notes")

	// uncheck
	var uncheckDate string
	uncheckCmd := &cobra.Command{
		Use:   "uncheck [id]",
		Short: "Remove a check-in for a habit",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			req := UncheckRequest{
				Date: uncheckDate,
			}
			if err := service.Uncheck(args[0], req); err != nil {
				return err
			}
			return printJSON(map[string]bool{"unchecked": true})
		},
	}
	uncheckCmd.Flags().StringVar(&uncheckDate, "date", "", "Date to uncheck (YYYY-MM-DD, defaults to today)")

	// skip
	var skipDate string
	var skipReason string
	skipCmd := &cobra.Command{
		Use:   "skip [id]",
		Short: "Mark a habit as skipped for a date",
		Long: `Mark a habit as skipped for a date.

Skipped check-ins behavior depends on user settings:
  - If skippedBreaksStreak=false (default): Skipped counts as completed
  - If skippedBreaksStreak=true: Skipped breaks the streak`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			req := SkipRequest{
				Date:   skipDate,
				Reason: skipReason,
			}
			checkin, err := service.Skip(args[0], req)
			if err != nil {
				return err
			}
			return printJSON(checkin.ToLean())
		},
	}
	skipCmd.Flags().StringVar(&skipDate, "date", "", "Date to skip (YYYY-MM-DD, defaults to today)")
	skipCmd.Flags().StringVarP(&skipReason, "reason", "r", "", "Reason for skipping")

	// checkins - get check-in history
	var checkinsFrom string
	var checkinsTo string
	checkinsCmd := &cobra.Command{
		Use:   "checkins [id]",
		Short: "Get check-in history for a habit",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			checkins, err := service.GetCheckIns(args[0], checkinsFrom, checkinsTo)
			if err != nil {
				return err
			}
			return printJSON(CheckInsToLeanSlice(checkins))
		},
	}
	checkinsCmd.Flags().StringVar(&checkinsFrom, "from", "", "Start date (YYYY-MM-DD)")
	checkinsCmd.Flags().StringVar(&checkinsTo, "to", "", "End date (YYYY-MM-DD)")

	cmd.AddCommand(
		listCmd,
		listAllCmd,
		listArchivedCmd,
		getCmd,
		createCmd,
		updateCmd,
		deleteCmd,
		statsCmd,
		checkCmd,
		uncheckCmd,
		skipCmd,
		checkinsCmd,
	)
	return cmd
}

// parseIntList parses a comma-separated list of integers
func parseIntList(s string) ([]int, error) {
	if s == "" {
		return nil, nil
	}
	parts := strings.Split(s, ",")
	result := make([]int, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		n, err := strconv.Atoi(p)
		if err != nil {
			return nil, fmt.Errorf("invalid number '%s': %w", p, err)
		}
		result = append(result, n)
	}
	return result, nil
}
