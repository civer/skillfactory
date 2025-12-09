// Package tui provides the terminal user interface for SkillFactory
package tui

import (
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/petervogelmann/skillfactory/internal/skill"
)

// View represents different screens in the TUI
type View int

const (
	ViewSkillList View = iota // Start here: list of available skills
	ViewConfig                // Configure skill environment variables
	ViewDeploy                // Configure deploy settings (folder, name)
	ViewConfirm               // Confirm and deploy
	ViewOverwrite             // Warning: skill already exists
	ViewBuilding              // Building in progress
	ViewDone                  // Success/Error result
)

// Model represents the application state
type Model struct {
	projectRoot   string
	version       string
	manifests     []*skill.Manifest
	skillErrors   []skill.SkillError // Skills that failed to load
	currentView   View
	skillCursor   int
	selectedSkill *skill.Manifest
	selectedError *skill.SkillError // Selected error skill (for viewing errors)

	// Skill environment variable inputs
	configInputs      []textinput.Model
	configLabels      []string
	configFocus       int

	// Deploy settings inputs
	deployInputs      []textinput.Model
	deployLabels      []string
	deployFocus       int

	// Deploy configuration (saved values)
	skillsFolder    string // Base folder for skills (e.g., /path/to/.claude/skills/)
	skillFolderName string // Subfolder name for this skill (default: skill name)

	// Configured values
	configValues map[string]string

	// Status messages
	statusMsg string
	errorMsg  string

	// Build state
	building    bool
	buildOutput string

	width    int
	height   int
	quitting bool
}

// NewModel creates a new TUI model
func NewModel(projectRoot string, version string) Model {
	// Discover skills
	manifests, skillErrors, err := skill.DiscoverSkills(projectRoot)
	if err != nil {
		// Will show error in UI
		manifests = []*skill.Manifest{}
		skillErrors = []skill.SkillError{}
	}

	return Model{
		projectRoot:  projectRoot,
		version:      version,
		manifests:    manifests,
		skillErrors:  skillErrors,
		currentView:  ViewSkillList,
		configValues: make(map[string]string),
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Global quit
		if msg.String() == "ctrl+c" {
			m.quitting = true
			return m, tea.Quit
		}

		// Config view: handle skill variable inputs
		if m.currentView == ViewConfig {
			switch msg.String() {
			case "esc":
				m.currentView = ViewSkillList
				m.errorMsg = ""
				return m, nil
			case "tab", "down":
				m.configInputs[m.configFocus].Blur()
				m.configFocus = (m.configFocus + 1) % len(m.configInputs)
				m.configInputs[m.configFocus].Focus()
				return m, textinput.Blink
			case "shift+tab", "up":
				m.configInputs[m.configFocus].Blur()
				m.configFocus--
				if m.configFocus < 0 {
					m.configFocus = len(m.configInputs) - 1
				}
				m.configInputs[m.configFocus].Focus()
				return m, textinput.Blink
			case "ctrl+d", "enter":
				// Validate and continue to deploy settings
				if m.validateConfigInputs() {
					m.saveConfigInputs()
					m.setupDeployInputs()
					m.currentView = ViewDeploy
				}
				return m, textinput.Blink
			default:
				// Pass ALL other keys to the text input
				return m.updateConfigInputs(msg)
			}
		}

		// Deploy view: handle deploy settings inputs
		if m.currentView == ViewDeploy {
			switch msg.String() {
			case "esc":
				m.currentView = ViewConfig
				m.errorMsg = ""
				// Re-focus config inputs
				m.configFocus = 0
				for i := range m.configInputs {
					m.configInputs[i].Blur()
				}
				if len(m.configInputs) > 0 {
					m.configInputs[0].Focus()
				}
				return m, textinput.Blink
			case "tab", "down":
				m.deployInputs[m.deployFocus].Blur()
				m.deployFocus = (m.deployFocus + 1) % len(m.deployInputs)
				m.deployInputs[m.deployFocus].Focus()
				return m, textinput.Blink
			case "shift+tab", "up":
				m.deployInputs[m.deployFocus].Blur()
				m.deployFocus--
				if m.deployFocus < 0 {
					m.deployFocus = len(m.deployInputs) - 1
				}
				m.deployInputs[m.deployFocus].Focus()
				return m, textinput.Blink
			case "ctrl+d", "enter":
				// Validate and continue to confirm
				if m.validateDeployInputs() {
					m.saveDeployInputs()
					m.currentView = ViewConfirm
				}
				return m, nil
			default:
				// Pass ALL other keys to the text input
				return m.updateDeployInputs(msg)
			}
		}

		return m.handleKeyPress(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case buildCompleteMsg:
		m.building = false
		m.buildOutput = msg.output
		if msg.err != nil {
			m.errorMsg = msg.err.Error()
			m.currentView = ViewDone
		} else {
			// Build succeeded, now deploy
			return m, m.deploySkill()
		}
		return m, nil

	case deployCompleteMsg:
		m.currentView = ViewDone
		if msg.err != nil {
			m.errorMsg = msg.err.Error()
		} else {
			m.statusMsg = "Skill deployed successfully!"
		}
		return m, nil
	}

	return m, nil
}

func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.currentView {
	case ViewSkillList:
		return m.handleSkillListView(msg)
	case ViewConfirm:
		return m.handleConfirmView(msg)
	case ViewOverwrite:
		return m.handleOverwriteView(msg)
	case ViewDone:
		return m.handleDoneView(msg)
	}
	return m, nil
}

func (m Model) handleSkillListView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	totalItems := len(m.manifests) + len(m.skillErrors)

	switch msg.String() {
	case "q", "esc":
		m.quitting = true
		return m, tea.Quit
	case "up", "k":
		if m.skillCursor > 0 {
			m.skillCursor--
		}
	case "down", "j":
		if m.skillCursor < totalItems-1 {
			m.skillCursor++
		}
	case "enter":
		if m.skillCursor < len(m.manifests) {
			// Valid skill selected
			m.selectedSkill = m.manifests[m.skillCursor]
			m.selectedError = nil
			m.currentView = ViewConfig
			m.setupInputsFromManifest()
			return m, textinput.Blink
		} else if m.skillCursor < totalItems {
			// Error skill selected - show error details
			errorIdx := m.skillCursor - len(m.manifests)
			m.selectedError = &m.skillErrors[errorIdx]
			m.selectedSkill = nil
			m.errorMsg = m.selectedError.Error.Error()
		}
	}
	return m, nil
}

func (m Model) handleConfirmView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "n":
		m.currentView = ViewDeploy
		m.deployFocus = 0
		for i := range m.deployInputs {
			m.deployInputs[i].Blur()
		}
		if len(m.deployInputs) > 0 {
			m.deployInputs[0].Focus()
		}
		return m, textinput.Blink
	case "enter", "y":
		// Check if skill already exists
		if m.skillExists() {
			m.currentView = ViewOverwrite
			return m, nil
		}
		// Start build
		m.currentView = ViewBuilding
		m.building = true
		m.errorMsg = ""
		m.statusMsg = ""
		return m, m.startBuild()
	}
	return m, nil
}

func (m Model) handleOverwriteView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "n":
		// Go back to confirm view
		m.currentView = ViewConfirm
		return m, nil
	case "y":
		// Proceed with build (overwrite)
		m.currentView = ViewBuilding
		m.building = true
		m.errorMsg = ""
		m.statusMsg = ""
		return m, m.startBuild()
	}
	return m, nil
}

// skillExists checks if the deploy path already contains a skill
func (m Model) skillExists() bool {
	deployPath := m.getDeployPath()
	if deployPath == "" {
		return false
	}

	// Check if the bin directory with binary exists
	binaryName := m.selectedSkill.Build.Binary
	if binaryName == "" {
		binaryName = m.selectedSkill.Name
	}
	binaryPath := filepath.Join(deployPath, "bin", binaryName)

	_, err := os.Stat(binaryPath)
	return err == nil
}

func (m Model) handleDoneView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter", "q", "esc":
		m.quitting = true
		return m, tea.Quit
	case "r":
		// Restart - go back to skill list
		m.currentView = ViewSkillList
		m.errorMsg = ""
		m.statusMsg = ""
		m.buildOutput = ""
		return m, nil
	}
	return m, nil
}

// setupInputsFromManifest creates input fields for skill environment variables
func (m *Model) setupInputsFromManifest() {
	if m.selectedSkill == nil {
		return
	}

	// Create inputs for skill variables only
	numInputs := len(m.selectedSkill.Variables)
	m.configInputs = make([]textinput.Model, numInputs)
	m.configLabels = make([]string, numInputs)

	for i, v := range m.selectedSkill.Variables {
		input := textinput.New()
		input.Placeholder = v.Placeholder
		if input.Placeholder == "" && v.Default != "" {
			input.Placeholder = v.Default
		}
		input.CharLimit = 200
		input.Width = 50

		if v.Type == "secret" {
			input.EchoMode = textinput.EchoPassword
		}

		// Load existing value if any
		if val, ok := m.configValues[v.Name]; ok {
			input.SetValue(val)
		}

		m.configInputs[i] = input
		m.configLabels[i] = v.Label
	}

	// Focus first input
	m.configFocus = 0
	for i := range m.configInputs {
		m.configInputs[i].Blur()
	}
	if len(m.configInputs) > 0 {
		m.configInputs[0].Focus()
	}
}

// setupDeployInputs creates input fields for deploy settings
func (m *Model) setupDeployInputs() {
	m.deployInputs = make([]textinput.Model, 2)
	m.deployLabels = make([]string, 2)

	// Skills Folder input
	skillsFolderInput := textinput.New()
	skillsFolderInput.Placeholder = "/path/to/.claude/skills/"
	skillsFolderInput.CharLimit = 200
	skillsFolderInput.Width = 50
	if m.skillsFolder != "" {
		skillsFolderInput.SetValue(m.skillsFolder)
	}
	m.deployInputs[0] = skillsFolderInput
	m.deployLabels[0] = "Skills Folder"

	// Skill Name input - pre-filled with skill name
	skillNameInput := textinput.New()
	if m.selectedSkill != nil {
		skillNameInput.Placeholder = m.selectedSkill.Name
	}
	skillNameInput.CharLimit = 100
	skillNameInput.Width = 50
	if m.skillFolderName != "" {
		skillNameInput.SetValue(m.skillFolderName)
	} else if m.selectedSkill != nil {
		skillNameInput.SetValue(m.selectedSkill.Name)
	}
	m.deployInputs[1] = skillNameInput
	m.deployLabels[1] = "Skill Name"

	// Focus first input
	m.deployFocus = 0
	for i := range m.deployInputs {
		m.deployInputs[i].Blur()
	}
	m.deployInputs[0].Focus()
}

func (m *Model) validateConfigInputs() bool {
	if m.selectedSkill == nil {
		m.errorMsg = "No skill selected"
		return false
	}

	// Validate required variables
	for i, v := range m.selectedSkill.Variables {
		if v.Required && m.configInputs[i].Value() == "" {
			m.errorMsg = v.Label + " is required"
			return false
		}
	}

	m.errorMsg = ""
	return true
}

func (m *Model) validateDeployInputs() bool {
	// Validate skills folder
	skillsFolder := m.deployInputs[0].Value()
	if skillsFolder == "" {
		m.errorMsg = "Skills Folder is required"
		return false
	}

	// Validate skill name
	skillName := m.deployInputs[1].Value()
	if skillName == "" {
		m.errorMsg = "Skill Name is required"
		return false
	}

	m.errorMsg = ""
	return true
}

func (m *Model) saveConfigInputs() {
	if m.selectedSkill == nil {
		return
	}

	// Save variable values
	for i, v := range m.selectedSkill.Variables {
		m.configValues[v.Name] = m.configInputs[i].Value()
	}
}

func (m *Model) saveDeployInputs() {
	m.skillsFolder = m.deployInputs[0].Value()
	m.skillFolderName = m.deployInputs[1].Value()
}

// getDeployPath returns the full deploy path (skillsFolder + skillFolderName)
func (m Model) getDeployPath() string {
	return filepath.Join(m.skillsFolder, m.skillFolderName)
}

func (m Model) updateConfigInputs(msg tea.Msg) (tea.Model, tea.Cmd) {
	cmds := make([]tea.Cmd, len(m.configInputs))
	for i := range m.configInputs {
		m.configInputs[i], cmds[i] = m.configInputs[i].Update(msg)
	}
	return m, tea.Batch(cmds...)
}

func (m Model) updateDeployInputs(msg tea.Msg) (tea.Model, tea.Cmd) {
	cmds := make([]tea.Cmd, len(m.deployInputs))
	for i := range m.deployInputs {
		m.deployInputs[i], cmds[i] = m.deployInputs[i].Update(msg)
	}
	return m, tea.Batch(cmds...)
}

// GetProjectRoot returns the project root, finding it if needed
func GetProjectRoot() string {
	// Try to find project root by looking for go.mod
	dir, _ := os.Getwd()
	for i := 0; i < 10; i++ {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			// Check if this is SkillFactory
			if _, err := os.Stat(filepath.Join(dir, "skills")); err == nil {
				return dir
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	// Fallback: assume we're in the right place
	dir, _ = os.Getwd()
	return dir
}
