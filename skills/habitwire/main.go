// HabitWire CLI Skill - Lean JSON output for Claude Code
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"habitwire/categories"
	"habitwire/client"
	"habitwire/habits"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

func init() {
	// Load .env from same directory as binary
	if exe, err := os.Executable(); err == nil {
		envPath := filepath.Join(filepath.Dir(exe), ".env")
		godotenv.Load(envPath)
	}
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "habitwire",
		Short: "HabitWire CLI for Claude Code",
	}

	// Create client (will fail later if env vars missing)
	apiClient, err := client.New()
	if err != nil {
		// Register commands anyway for --help to work
		rootCmd.AddCommand(
			habits.RegisterCommands(nil, printJSON),
			categories.RegisterCommands(nil, printJSON),
		)
		// Only fail if actually trying to run a command
		rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
			return err
		}
	} else {
		rootCmd.AddCommand(
			habits.RegisterCommands(apiClient, printJSON),
			categories.RegisterCommands(apiClient, printJSON),
		)
	}

	if err := rootCmd.Execute(); err != nil {
		printError(err.Error())
		os.Exit(1)
	}
}

func printJSON(v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

func printError(msg string) {
	json.NewEncoder(os.Stderr).Encode(map[string]string{"error": msg})
}
