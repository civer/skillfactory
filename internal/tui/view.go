// Package tui provides the terminal user interface for SkillFactory
package tui

import (
	"fmt"
	"strings"
)

// View renders the current view
func (m Model) View() string {
	if m.quitting {
		return "Goodbye!\n"
	}

	var b strings.Builder

	// Header
	b.WriteString(titleStyle.Render("Claude SkillFactory"))
	b.WriteString(mutedStyle.Render("  " + m.version))
	b.WriteString("\n\n")

	// Main content based on current view
	switch m.currentView {
	case ViewSkillList:
		b.WriteString(m.renderSkillList())
	case ViewConfig:
		b.WriteString(m.renderConfig())
	case ViewDeploy:
		b.WriteString(m.renderDeploy())
	case ViewConfirm:
		b.WriteString(m.renderConfirm())
	case ViewOverwrite:
		b.WriteString(m.renderOverwrite())
	case ViewBuilding:
		b.WriteString(m.renderBuilding())
	case ViewDone:
		b.WriteString(m.renderDone())
	}

	// Error message
	if m.errorMsg != "" {
		b.WriteString("\n")
		b.WriteString(errorStyle.Render("✗ " + m.errorMsg))
	}

	// Help
	b.WriteString("\n\n")
	b.WriteString(m.renderHelp())

	return b.String()
}

func (m Model) renderSkillList() string {
	var b strings.Builder

	b.WriteString(inputLabelStyle.Render("Available Skills"))
	b.WriteString("\n\n")

	if len(m.manifests) == 0 && len(m.skillErrors) == 0 {
		b.WriteString(mutedStyle.Render("  No skills found in skills/ directory"))
		b.WriteString("\n")
		b.WriteString(mutedStyle.Render("  Add a skill.yaml to register a skill"))
	} else {
		// Valid skills
		for i, manifest := range m.manifests {
			cursor := "  "
			style := normalStyle
			if i == m.skillCursor {
				cursor = "▸ "
				style = selectedStyle
			}

			b.WriteString(cursor)
			b.WriteString(style.Render(manifest.Name))
			b.WriteString("\n")
			b.WriteString("    ")
			b.WriteString(mutedStyle.Render(manifest.Description))
			b.WriteString("\n")
		}

		// Error skills
		if len(m.skillErrors) > 0 {
			b.WriteString("\n")
			b.WriteString(errorStyle.Render("Skills with Errors"))
			b.WriteString("\n\n")

			for i, skillErr := range m.skillErrors {
				idx := len(m.manifests) + i
				cursor := "  "
				style := mutedStyle
				if idx == m.skillCursor {
					cursor = "▸ "
					style = errorStyle
				}

				b.WriteString(cursor)
				b.WriteString(style.Render(skillErr.Name))
				b.WriteString("\n")
				b.WriteString("    ")
				b.WriteString(mutedStyle.Render(skillErr.Error.Error()))
				b.WriteString("\n")
			}
		}
	}

	return boxStyle.Render(b.String())
}

func (m Model) renderConfig() string {
	var b strings.Builder

	if m.selectedSkill != nil {
		b.WriteString(inputLabelStyle.Render(fmt.Sprintf("Step 1: %s Environment", m.selectedSkill.Name)))
	} else {
		b.WriteString(inputLabelStyle.Render("Step 1: Skill Environment"))
	}
	b.WriteString("\n\n")

	for i, input := range m.configInputs {
		// Label with focus indicator
		labelStyle := mutedStyle
		prefix := "  "
		if i == m.configFocus {
			labelStyle = inputLabelStyle
			prefix = "▸ "
		}

		// Show label
		label := m.configLabels[i]

		// Add required indicator
		if m.selectedSkill != nil && i < len(m.selectedSkill.Variables) {
			v := m.selectedSkill.Variables[i]
			if v.Required {
				label += " *"
			}
		}

		b.WriteString(prefix)
		b.WriteString(labelStyle.Render(label))
		b.WriteString("\n")

		// Input field
		b.WriteString("  ")
		b.WriteString(input.View())
		b.WriteString("\n\n")
	}

	b.WriteString(mutedStyle.Render("  * required"))

	return boxStyle.Render(b.String())
}

func (m Model) renderDeploy() string {
	var b strings.Builder

	b.WriteString(inputLabelStyle.Render("Step 2: Deploy Settings"))
	b.WriteString("\n\n")

	for i, input := range m.deployInputs {
		// Label with focus indicator
		labelStyle := mutedStyle
		prefix := "  "
		if i == m.deployFocus {
			labelStyle = inputLabelStyle
			prefix = "▸ "
		}

		// Show label - all deploy fields are required
		label := m.deployLabels[i] + " *"

		b.WriteString(prefix)
		b.WriteString(labelStyle.Render(label))
		b.WriteString("\n")

		// Input field
		b.WriteString("  ")
		b.WriteString(input.View())
		b.WriteString("\n\n")
	}

	b.WriteString(mutedStyle.Render("  * required"))

	return boxStyle.Render(b.String())
}

func (m Model) renderConfirm() string {
	var b strings.Builder

	b.WriteString(inputLabelStyle.Render("Step 3: Confirm"))
	b.WriteString("\n\n")

	// Show skill info
	if m.selectedSkill != nil {
		b.WriteString(mutedStyle.Render("  Skill:         "))
		b.WriteString(normalStyle.Render(m.selectedSkill.Name))
		b.WriteString("\n\n")

		// Show all configured values
		if len(m.selectedSkill.Variables) > 0 {
			b.WriteString(mutedStyle.Render("  Environment:"))
			b.WriteString("\n")
			for _, v := range m.selectedSkill.Variables {
				value := m.configValues[v.Name]
				if v.Type == "secret" && len(value) > 8 {
					value = value[:8] + "..."
				}
				b.WriteString(mutedStyle.Render(fmt.Sprintf("    %-12s ", v.Label+":")))
				b.WriteString(normalStyle.Render(value))
				b.WriteString("\n")
			}
			b.WriteString("\n")
		}
	}

	// Deploy configuration
	b.WriteString(mutedStyle.Render("  Deploy:"))
	b.WriteString("\n")
	b.WriteString(mutedStyle.Render("    Skills Folder: "))
	b.WriteString(normalStyle.Render(m.skillsFolder))
	b.WriteString("\n")
	b.WriteString(mutedStyle.Render("    Skill Name:    "))
	b.WriteString(normalStyle.Render(m.skillFolderName))
	b.WriteString("\n")
	b.WriteString(mutedStyle.Render("    Target:        "))
	b.WriteString(successStyle.Render(m.getDeployPath()))
	b.WriteString("\n\n")

	b.WriteString(normalStyle.Render("  Build and deploy this skill?"))
	b.WriteString("\n")
	b.WriteString(mutedStyle.Render("  [Y] Yes  [N] Back"))

	return boxStyle.Render(b.String())
}

func (m Model) renderBuilding() string {
	var b strings.Builder

	b.WriteString(inputLabelStyle.Render("Building..."))
	b.WriteString("\n\n")

	skillName := "skill"
	if m.selectedSkill != nil {
		skillName = m.selectedSkill.Name
	}
	b.WriteString(mutedStyle.Render("  Compiling " + skillName + "..."))

	return boxStyle.Render(b.String())
}

func (m Model) renderDone() string {
	var b strings.Builder

	if m.statusMsg != "" {
		b.WriteString(successStyle.Render("✓ " + m.statusMsg))
		b.WriteString("\n\n")

		b.WriteString(mutedStyle.Render("  Deployed to: "))
		b.WriteString(normalStyle.Render(m.getDeployPath()))
		b.WriteString("\n\n")

		b.WriteString(mutedStyle.Render("  The skill is now ready to use!"))
	} else if m.errorMsg != "" {
		b.WriteString(errorStyle.Render("✗ Build/Deploy failed"))
		b.WriteString("\n\n")
		if m.buildOutput != "" {
			b.WriteString(mutedStyle.Render("  Output: " + m.buildOutput))
		}
	}

	return boxStyle.Render(b.String())
}

func (m Model) renderOverwrite() string {
	var b strings.Builder

	b.WriteString(errorStyle.Render("⚠️  Skill already exists"))
	b.WriteString("\n\n")

	skillName := "skill"
	if m.selectedSkill != nil {
		skillName = m.selectedSkill.Name
	}

	b.WriteString(mutedStyle.Render("  The skill \""))
	b.WriteString(normalStyle.Render(skillName))
	b.WriteString(mutedStyle.Render("\" already exists at:"))
	b.WriteString("\n")
	b.WriteString("  ")
	b.WriteString(normalStyle.Render(m.getDeployPath()))
	b.WriteString("\n\n")

	b.WriteString(normalStyle.Render("  Overwrite?"))
	b.WriteString("\n")
	b.WriteString(mutedStyle.Render("  [Y] Yes, overwrite  [N] Cancel"))

	return boxStyle.Render(b.String())
}

func (m Model) renderHelp() string {
	var help string

	switch m.currentView {
	case ViewSkillList:
		help = "↑/↓: Navigate • Enter: Select • q: Quit"
	case ViewConfig:
		help = "↑/↓/Tab: Navigate • Enter: Next Step • Esc: Back"
	case ViewDeploy:
		help = "↑/↓/Tab: Navigate • Enter: Next Step • Esc: Back"
	case ViewConfirm:
		help = "Y/Enter: Build & Deploy • N/Esc: Back"
	case ViewOverwrite:
		help = "Y: Overwrite • N/Esc: Cancel"
	case ViewBuilding:
		help = "Building..."
	case ViewDone:
		help = "Enter/q: Quit • R: Configure another skill"
	}

	return helpStyle.Render(help)
}
