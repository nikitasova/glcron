package tui

import (
	"glcron/internal/models"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ConfigFormField int

const (
	ConfigFieldName ConfigFormField = iota
	ConfigFieldURL
	ConfigFieldToken
	ConfigFieldSave
	ConfigFieldCancel
)

type ConfigFormModel struct {
	isNew        bool
	configIndex  int
	focusedField ConfigFormField
	width        int
	height       int

	nameInput  textinput.Model
	urlInput   textinput.Model
	tokenInput textinput.Model
}

func NewConfigFormModel() ConfigFormModel {
	nameInput := textinput.New()
	nameInput.Placeholder = "Configuration name"
	nameInput.CharLimit = 50
	nameInput.Width = 40
	nameInput.Cursor.Style = CursorStyle
	nameInput.Focus()

	urlInput := textinput.New()
	urlInput.Placeholder = "https://gitlab.com/group/project"
	urlInput.CharLimit = 200
	urlInput.Width = 50
	urlInput.Cursor.Style = CursorStyle

	tokenInput := textinput.New()
	tokenInput.Placeholder = "glpat-..."
	tokenInput.CharLimit = 100
	tokenInput.Width = 50
	tokenInput.EchoMode = textinput.EchoPassword
	tokenInput.Cursor.Style = CursorStyle

	return ConfigFormModel{
		nameInput:  nameInput,
		urlInput:   urlInput,
		tokenInput: tokenInput,
	}
}

func (m *ConfigFormModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

func (m *ConfigFormModel) SetConfig(config *models.Config, index int, isNew bool) {
	m.isNew = isNew
	m.configIndex = index

	if config != nil {
		m.nameInput.SetValue(config.Name)
		m.urlInput.SetValue(config.ProjectURL)
		m.tokenInput.SetValue(config.Token)
	} else {
		m.nameInput.SetValue("")
		m.urlInput.SetValue("")
		m.tokenInput.SetValue("")
	}

	m.focusedField = ConfigFieldName
	m.nameInput.Focus()
}

func (m ConfigFormModel) Update(msg tea.Msg) (ConfigFormModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, Navigate(ScreenConfigList)
		case "tab", "down":
			m.nextField()
		case "shift+tab", "up":
			m.prevField()
		case "enter":
			return m.handleEnter()
		case "ctrl+s":
			return m.save()
		case "left":
			// Only switch buttons if on button fields, otherwise pass to text input
			if m.focusedField == ConfigFieldCancel {
				m.focusedField = ConfigFieldSave
			} else if m.focusedField == ConfigFieldName || m.focusedField == ConfigFieldURL || m.focusedField == ConfigFieldToken {
				return m.handleInputKey(msg)
			}
		case "right":
			// Only switch buttons if on button fields, otherwise pass to text input
			if m.focusedField == ConfigFieldSave {
				m.focusedField = ConfigFieldCancel
			} else if m.focusedField == ConfigFieldName || m.focusedField == ConfigFieldURL || m.focusedField == ConfigFieldToken {
				return m.handleInputKey(msg)
			}
		default:
			return m.handleInputKey(msg)
		}
	}

	return m, nil
}

func (m *ConfigFormModel) nextField() {
	m.blurCurrent()
	switch m.focusedField {
	case ConfigFieldName:
		m.focusedField = ConfigFieldURL
	case ConfigFieldURL:
		m.focusedField = ConfigFieldToken
	case ConfigFieldToken:
		m.focusedField = ConfigFieldSave
	case ConfigFieldSave:
		m.focusedField = ConfigFieldCancel
	case ConfigFieldCancel:
		m.focusedField = ConfigFieldName
	}
	m.focusCurrent()
}

func (m *ConfigFormModel) prevField() {
	m.blurCurrent()
	switch m.focusedField {
	case ConfigFieldName:
		m.focusedField = ConfigFieldCancel
	case ConfigFieldURL:
		m.focusedField = ConfigFieldName
	case ConfigFieldToken:
		m.focusedField = ConfigFieldURL
	case ConfigFieldSave:
		m.focusedField = ConfigFieldToken
	case ConfigFieldCancel:
		m.focusedField = ConfigFieldSave
	}
	m.focusCurrent()
}

func (m *ConfigFormModel) blurCurrent() {
	m.nameInput.Blur()
	m.urlInput.Blur()
	m.tokenInput.Blur()
}

func (m *ConfigFormModel) focusCurrent() {
	switch m.focusedField {
	case ConfigFieldName:
		m.nameInput.Focus()
	case ConfigFieldURL:
		m.urlInput.Focus()
	case ConfigFieldToken:
		m.tokenInput.Focus()
	}
}

func (m ConfigFormModel) handleEnter() (ConfigFormModel, tea.Cmd) {
	switch m.focusedField {
	case ConfigFieldSave:
		return m.save()
	case ConfigFieldCancel:
		return m, Navigate(ScreenConfigList)
	default:
		m.nextField()
	}
	return m, nil
}

func (m ConfigFormModel) handleInputKey(msg tea.KeyMsg) (ConfigFormModel, tea.Cmd) {
	var cmd tea.Cmd
	switch m.focusedField {
	case ConfigFieldName:
		m.nameInput, cmd = m.nameInput.Update(msg)
	case ConfigFieldURL:
		m.urlInput, cmd = m.urlInput.Update(msg)
	case ConfigFieldToken:
		m.tokenInput, cmd = m.tokenInput.Update(msg)
	}
	return m, cmd
}

func (m ConfigFormModel) save() (ConfigFormModel, tea.Cmd) {
	return m, func() tea.Msg {
		return saveConfigMsg{
			index: m.configIndex,
			name:  m.nameInput.Value(),
			url:   m.urlInput.Value(),
			token: m.tokenInput.Value(),
			isNew: m.isNew,
		}
	}
}

func (m ConfigFormModel) View() string {
	leftWidth := (m.width * 3) / 5
	rightWidth := m.width - leftWidth - 1

	leftLines := m.renderFormPanel(leftWidth)
	rightLines := m.renderHelpPanel(rightWidth)

	var result []string
	maxLines := maxInt(len(leftLines), len(rightLines))

	for i := 0; i < maxLines; i++ {
		left := ""
		if i < len(leftLines) {
			left = leftLines[i]
		}
		left = padToWidth(left, leftWidth)

		right := ""
		if i < len(rightLines) {
			right = rightLines[i]
		}
		right = padToWidth(right, rightWidth)

		result = append(result, left+"│"+right)
	}

	return strings.Join(result, "\n")
}

func (m ConfigFormModel) renderFormPanel(width int) []string {
	// Use shared styles
	label := LabelStyle
	selected := SelectedStyle

	var lines []string

	// Title with icon
	title := " ✏️ Edit Configuration "
	if m.isNew {
		title = " ➕ New Configuration "
	}
	titleWidth := lipgloss.Width(title)
	borderLen := width - titleWidth - 4
	if borderLen < 0 {
		borderLen = 0
	}
	lines = append(lines, BorderTopLeft+BorderTop+title+strings.Repeat(BorderTop, borderLen)+BorderTop+BorderTopRight)

	var content []string

	// Label column width for alignment
	labelWidth := 20

	// Name field - cursor shows when focused
	nameLabel := label.Render(padRight("  Name", labelWidth))
	nameValue := m.nameInput.View()
	content = append(content, nameLabel+" "+nameValue)
	content = append(content, "") // Gap

	// URL field - cursor shows when focused
	urlLabel := label.Render(padRight("  Project URL", labelWidth))
	urlValue := m.urlInput.View()
	content = append(content, urlLabel+" "+urlValue)
	content = append(content, "") // Gap

	// Token field - cursor shows when focused
	tokenLabel := label.Render(padRight("  Access Token", labelWidth))
	tokenValue := m.tokenInput.View()
	content = append(content, tokenLabel+" "+tokenValue)
	content = append(content, "") // Gap
	content = append(content, "") // Extra gap before buttons

	// Buttons
	saveBtn := " 💾 Save "
	cancelBtn := " ✖ Cancel "
	if m.focusedField == ConfigFieldSave {
		content = append(content, "  "+selected.Render(saveBtn)+"   "+cancelBtn)
	} else if m.focusedField == ConfigFieldCancel {
		content = append(content, "  "+saveBtn+"   "+selected.Render(cancelBtn))
	} else {
		content = append(content, "  "+saveBtn+"   "+cancelBtn)
	}

	// Render content with borders
	for _, line := range content {
		paddedLine := " " + padToWidth(line, width-4) + " "
		lines = append(lines, "│"+paddedLine+"│")
	}

	// Fill remaining space
	for len(lines) < m.height-2 {
		lines = append(lines, "│"+strings.Repeat(" ", width-2)+"│")
	}

	lines = append(lines, "└"+strings.Repeat("─", width-2)+"┘")

	return lines
}

func (m ConfigFormModel) renderHelpPanel(width int) []string {
	// Use shared styles
	heading := TitleStyle
	highlight := YellowStyle
	example := GreenStyle
	muted := GrayStyle

	var lines []string

	// Title
	title := " Help "
	titleWidth := lipgloss.Width(title)
	borderLen := width - titleWidth - 4
	if borderLen < 0 {
		borderLen = 0
	}
	lines = append(lines, BorderTopLeft+BorderTop+title+strings.Repeat(BorderTop, borderLen)+BorderTop+BorderTopRight)

	var content []string

	content = append(content, heading.Render("Configuration Setup"))
	content = append(content, "")
	content = append(content, "Connect glcron to your GitLab project by")
	content = append(content, "providing the project URL and access token.")
	content = append(content, "")

	content = append(content, heading.Render("Project URL Format"))
	content = append(content, "")
	content = append(content, "  "+example.Render("https://gitlab.com/group/project"))
	content = append(content, "  "+example.Render("https://gitlab.company.com/team/repo"))
	content = append(content, "")

	content = append(content, heading.Render("Creating an Access Token"))
	content = append(content, "")
	content = append(content, "1. Go to GitLab → Settings → Access Tokens")
	content = append(content, "2. Create token with "+highlight.Render("api")+" scope")
	content = append(content, "3. Copy and paste the token here")
	content = append(content, "")
	content = append(content, muted.Render("Token stored in: ~/.config/glcron/"))
	content = append(content, "")

	content = append(content, heading.Render("Keyboard Shortcuts"))
	content = append(content, "")
	content = append(content, "  "+highlight.Render("↑/↓")+"       Navigate fields")
	content = append(content, "  "+highlight.Render("Tab")+"       Next field")
	content = append(content, "  "+highlight.Render("Ctrl+S")+"    Save")
	content = append(content, "  "+highlight.Render("Esc")+"       Cancel")

	// Render content with borders
	for _, line := range content {
		paddedLine := " " + padToWidth(line, width-4) + " "
		lines = append(lines, "│"+paddedLine+"│")
	}

	// Fill remaining space
	for len(lines) < m.height-2 {
		lines = append(lines, "│"+strings.Repeat(" ", width-2)+"│")
	}

	lines = append(lines, "└"+strings.Repeat("─", width-2)+"┘")

	return lines
}
