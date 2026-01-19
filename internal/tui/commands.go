// Package tui provides the terminal user interface for SkillFactory
package tui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// buildCompleteMsg is sent when a build completes
type buildCompleteMsg struct {
	output string
	err    error
}

// deployCompleteMsg is sent when a deploy completes
type deployCompleteMsg struct {
	err error
}

// startBuild starts the build process for the selected skill
func (m Model) startBuild() tea.Cmd {
	return func() tea.Msg {
		if m.selectedSkill == nil {
			return buildCompleteMsg{err: fmt.Errorf("no skill selected")}
		}

		// Get skill source path
		skillPath := m.selectedSkill.Path
		binaryName := m.selectedSkill.Build.Binary
		if binaryName == "" {
			binaryName = m.selectedSkill.Name
		}

		// Build to dist directory
		distDir := filepath.Join(m.projectRoot, "dist")
		os.MkdirAll(distDir, 0755)
		outputPath := filepath.Join(distDir, binaryName)

		// Run go build
		cmd := exec.Command("go", "build", "-o", outputPath, ".")
		cmd.Dir = skillPath

		output, err := cmd.CombinedOutput()
		if err != nil {
			return buildCompleteMsg{
				output: string(output),
				err:    fmt.Errorf("build failed: %w", err),
			}
		}

		return buildCompleteMsg{
			output: fmt.Sprintf("Built: %s", outputPath),
		}
	}
}

// deploySkill deploys the built skill to the configured path
func (m Model) deploySkill() tea.Cmd {
	return func() tea.Msg {
		if m.selectedSkill == nil {
			return deployCompleteMsg{err: fmt.Errorf("no skill selected")}
		}

		deployPath := m.getDeployPath()
		if deployPath == "" {
			return deployCompleteMsg{err: fmt.Errorf("deploy path not configured")}
		}

		binaryName := m.selectedSkill.Build.Binary
		if binaryName == "" {
			binaryName = m.selectedSkill.Name
		}

		// Source paths
		distDir := filepath.Join(m.projectRoot, "dist")
		srcBinary := filepath.Join(distDir, binaryName)

		// Destination paths
		dstBinDir := filepath.Join(deployPath, "bin")
		dstBinary := filepath.Join(dstBinDir, binaryName)

		// Ensure destination directories exist
		os.MkdirAll(dstBinDir, 0755)

		// Copy binary (remove old one first to avoid issues with running processes)
		binaryData, err := os.ReadFile(srcBinary)
		if err != nil {
			return deployCompleteMsg{err: fmt.Errorf("failed to read binary: %w", err)}
		}

		// Remove existing binary first to ensure clean overwrite
		os.Remove(dstBinary)

		if err := os.WriteFile(dstBinary, binaryData, 0755); err != nil {
			return deployCompleteMsg{err: fmt.Errorf("failed to write binary: %w", err)}
		}

		// Generate .env file with environment variables
		envPath := filepath.Join(dstBinDir, ".env")
		envContent := m.generateEnvFile()
		if err := os.WriteFile(envPath, []byte(envContent), 0600); err != nil {
			return deployCompleteMsg{err: fmt.Errorf("failed to write .env: %w", err)}
		}

		// Generate SKILL.md
		if err := m.generateSkillDocs(); err != nil {
			return deployCompleteMsg{err: fmt.Errorf("failed to generate docs: %w", err)}
		}

		// Cleanup: remove dist directory
		os.RemoveAll(distDir)

		return deployCompleteMsg{}
	}
}

// generateEnvFile creates a .env file with environment variables
func (m Model) generateEnvFile() string {
	var b strings.Builder

	b.WriteString("# Auto-generated environment file\n")

	for _, v := range m.selectedSkill.Variables {
		if value, ok := m.configValues[v.Name]; ok && value != "" {
			b.WriteString(fmt.Sprintf("%s=%s\n", v.Name, value))
		}
	}

	return b.String()
}

// generateSkillDocs generates the SKILL.md file
func (m Model) generateSkillDocs() error {
	if m.selectedSkill == nil {
		return fmt.Errorf("no skill selected")
	}

	// Read template if exists
	templatePath := filepath.Join(m.selectedSkill.Path, m.selectedSkill.Docs.Template)
	var content string

	templateData, err := os.ReadFile(templatePath)
	if err != nil {
		// No template, generate basic docs
		content = m.generateBasicDocs()
	} else {
		content = string(templateData)
		// Remove any existing frontmatter from template
		content = stripFrontmatter(content)
		// Replace placeholders
		content = m.replacePlaceholders(content)
	}

	// Prepend generated frontmatter from skill.yaml
	frontmatter := m.generateFrontmatter()
	content = frontmatter + content

	// Write SKILL.md
	outputPath := filepath.Join(m.getDeployPath(), "SKILL.md")
	return os.WriteFile(outputPath, []byte(content), 0644)
}

// generateFrontmatter creates YAML frontmatter from skill manifest
func (m Model) generateFrontmatter() string {
	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString(fmt.Sprintf("name: %s\n", m.selectedSkill.Name))
	b.WriteString(fmt.Sprintf("description: %s\n", m.selectedSkill.GetSkillDescription()))
	b.WriteString("---\n\n")
	return b.String()
}

// stripFrontmatter removes existing YAML frontmatter from content
func stripFrontmatter(content string) string {
	if !strings.HasPrefix(content, "---") {
		return content
	}
	// Find the closing ---
	rest := content[3:]
	idx := strings.Index(rest, "---")
	if idx == -1 {
		return content
	}
	// Return everything after the closing --- and any leading newlines
	result := strings.TrimLeft(rest[idx+3:], "\n")
	return result
}

// generateBasicDocs generates basic documentation
func (m Model) generateBasicDocs() string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("# %s\n\n", m.selectedSkill.Name))
	b.WriteString(m.selectedSkill.Description)
	b.WriteString("\n\n")

	b.WriteString("## Commands\n\n")
	b.WriteString("Run `" + m.selectedSkill.Build.Binary + " --help` to see available commands.\n")

	return b.String()
}

// replacePlaceholders replaces template placeholders
func (m Model) replacePlaceholders(content string) string {
	binaryName := m.selectedSkill.Build.Binary
	if binaryName == "" {
		binaryName = m.selectedSkill.Name
	}

	// Replace SKILL_PATH placeholder
	content = strings.Replace(content, "{{SKILL_PATH}}", m.getDeployPath(), -1)

	// Replace PROJECT_IDS_TABLE if we have PROJECT_IDS configured
	if projectIDs, ok := m.configValues["PROJECT_IDS"]; ok && projectIDs != "" {
		table := m.generateProjectIDsTable(projectIDs)
		content = strings.Replace(content, "{{PROJECT_IDS_TABLE}}", table, 1)
	} else {
		content = strings.Replace(content, "{{PROJECT_IDS_TABLE}}", "No project IDs configured.", 1)
	}

	// Extract commands from built binary
	distDir := filepath.Join(m.projectRoot, "dist")
	binaryPath := filepath.Join(distDir, binaryName)

	// Generate commands with binary path for SKILL.md (binary loads .env automatically)
	deployedBinaryPath := filepath.Join(m.getDeployPath(), "bin", binaryName)
	commands := m.extractCommands(binaryPath, deployedBinaryPath)
	content = strings.Replace(content, "{{COMMANDS}}", commands, 1)

	return content
}

// extractCommands runs the binary with --help recursively and documents all leaf commands with flags
// binaryPath is the built binary, displayPath is what to show in docs
func (m Model) extractCommands(binaryPath string, displayPath string) string {
	// Run binary --help to get top-level help
	output, err := runHelp(binaryPath)
	if err != nil {
		return "Run `" + displayPath + " --help` to see available commands."
	}

	// Extract top-level subcommands
	topLevel := parseSubcommands(output)
	if len(topLevel) == 0 {
		return "Run `" + displayPath + " --help` to see available commands."
	}

	var b strings.Builder

	for _, cmd := range topLevel {
		// Get second-level subcommands
		cmdOutput, err := runHelp(binaryPath, cmd)
		if err != nil {
			continue
		}

		secondLevel := parseSubcommands(cmdOutput)

		if len(secondLevel) == 0 {
			// This is a leaf command - document it with flags
			b.WriteString(formatCommand(displayPath, cmd, cmdOutput))
		} else {
			// Has subcommands - recurse one more level
			for _, sub := range secondLevel {
				subOutput, err := runHelp(binaryPath, cmd, sub)
				if err != nil {
					continue
				}
				fullCmd := cmd + " " + sub
				b.WriteString(formatCommand(displayPath, fullCmd, subOutput))
			}
		}
	}

	if b.Len() == 0 {
		return "Run `" + displayPath + " --help` to see available commands."
	}

	return b.String()
}

// runHelp executes a command with --help and returns the output
func runHelp(binaryPath string, args ...string) (string, error) {
	cmdArgs := append(args, "--help")
	cmd := exec.Command(binaryPath, cmdArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// formatCommand formats a leaf command with its description and flags
func formatCommand(displayPath string, cmdPath string, helpOutput string) string {
	var b strings.Builder

	// Parse the help output
	description := parseDescription(helpOutput)
	usage := parseUsage(helpOutput)
	flags := parseFlags(helpOutput)

	b.WriteString("### " + cmdPath + "\n\n")

	if description != "" {
		b.WriteString(description + "\n\n")
	}

	if usage != "" {
		b.WriteString("**Usage:** `" + displayPath + " " + usage + "`\n\n")
	}

	if len(flags) > 0 {
		b.WriteString("**Flags:**\n")
		for _, flag := range flags {
			b.WriteString("- " + flag + "\n")
		}
		b.WriteString("\n")
	}

	return b.String()
}

// parseDescription extracts the description from Cobra help output (first non-empty line)
func parseDescription(helpOutput string) string {
	lines := strings.Split(helpOutput, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !strings.HasPrefix(trimmed, "Usage:") {
			return trimmed
		}
		if strings.HasPrefix(trimmed, "Usage:") {
			break
		}
	}
	return ""
}

// parseUsage extracts the usage pattern from Cobra help output
func parseUsage(helpOutput string) string {
	lines := strings.Split(helpOutput, "\n")
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "Usage:") {
			// Get the next line or the rest of this line
			if i+1 < len(lines) {
				usage := strings.TrimSpace(lines[i+1])
				// Remove the binary path prefix, keep just the command pattern
				parts := strings.Fields(usage)
				if len(parts) > 1 {
					// Skip the binary name, return the rest
					return strings.Join(parts[1:], " ")
				}
			}
		}
	}
	return ""
}

// parseFlags extracts flags from Cobra help output
func parseFlags(helpOutput string) []string {
	var flags []string
	lines := strings.Split(helpOutput, "\n")
	inFlagsSection := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "Flags:") {
			inFlagsSection = true
			continue
		}

		// End flags section at next section or empty line followed by non-flag
		if inFlagsSection {
			if trimmed == "" {
				continue
			}
			if strings.HasSuffix(trimmed, ":") {
				break
			}

			// Parse flag line: "  -t, --title string   Task title (required)"
			if strings.HasPrefix(trimmed, "-") {
				// Format: `-short, --long type   description`
				flag := formatFlag(trimmed)
				if flag != "" && !strings.Contains(flag, "--help") {
					flags = append(flags, flag)
				}
			}
		}
	}

	return flags
}

// formatFlag formats a Cobra flag line into a readable format
func formatFlag(line string) string {
	// Input: "  -t, --title string    Task title (required)"
	// Output: "`-t, --title` (string): Task title (required)"

	parts := strings.Fields(line)
	if len(parts) < 2 {
		return ""
	}

	var flagPart string
	var typePart string
	var descParts []string

	i := 0
	// Collect flag names (-t, --title)
	for i < len(parts) && (strings.HasPrefix(parts[i], "-") || parts[i] == ",") {
		if parts[i] != "," {
			if flagPart != "" {
				flagPart += ", "
			}
			flagPart += strings.TrimSuffix(parts[i], ",")
		}
		i++
	}

	// Next part might be type (string, int, etc.) or description
	if i < len(parts) {
		// Common types in Cobra
		commonTypes := []string{"string", "int", "int64", "bool", "float64", "duration", "stringArray", "intSlice"}
		isType := false
		for _, t := range commonTypes {
			if parts[i] == t {
				isType = true
				break
			}
		}
		if isType {
			typePart = parts[i]
			i++
		}
	}

	// Rest is description
	if i < len(parts) {
		descParts = parts[i:]
	}

	result := "`" + flagPart + "`"
	if typePart != "" {
		result += " (" + typePart + ")"
	}
	if len(descParts) > 0 {
		result += ": " + strings.Join(descParts, " ")
	}

	return result
}

// parseSubcommands extracts subcommand names from Cobra help output
func parseSubcommands(helpText string) []string {
	var commands []string

	lines := strings.Split(helpText, "\n")
	inCommandsSection := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Start of commands section
		if strings.HasPrefix(trimmed, "Available Commands:") {
			inCommandsSection = true
			continue
		}

		// End of commands section (empty line or new section)
		if inCommandsSection {
			if trimmed == "" || strings.HasSuffix(trimmed, ":") {
				if strings.HasSuffix(trimmed, ":") && !strings.HasPrefix(trimmed, "Available") {
					break
				}
				continue
			}

			// Parse command name (first word)
			parts := strings.Fields(trimmed)
			if len(parts) > 0 {
				cmdName := parts[0]
				// Skip help and completion commands
				if cmdName != "help" && cmdName != "completion" {
					commands = append(commands, cmdName)
				}
			}
		}
	}

	return commands
}

// generateProjectIDsTable generates a markdown table from PROJECT_IDS JSON
func (m Model) generateProjectIDsTable(jsonStr string) string {
	// Simple JSON parsing for {"Name": ID} format
	// For now, just return the raw JSON prettified
	var b strings.Builder
	b.WriteString("| ID | Projekt |\n")
	b.WriteString("|----|---------|")

	// TODO: Parse JSON properly and generate table
	// For now, include raw config
	b.WriteString("\n\nConfig: `")
	b.WriteString(jsonStr)
	b.WriteString("`")

	return b.String()
}
