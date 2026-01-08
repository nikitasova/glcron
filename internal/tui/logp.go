package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// LogType defines the type of log message for styling
type LogType string

const (
	LogTypeInfo    LogType = "info"
	LogTypeSuccess LogType = "success"
	LogTypeWarning LogType = "warning"
	LogTypeError   LogType = "error"
)

// LogPanel is a reusable component for displaying status messages
type LogPanel struct {
	message  string
	logType  LogType
	visible  bool
}

// NewLogPanel creates a new log panel
func NewLogPanel() *LogPanel {
	return &LogPanel{
		message: "",
		logType: LogTypeInfo,
		visible: false,
	}
}

// Show displays a log message with the specified type
func (l *LogPanel) Show(message string, logType LogType) {
	l.message = message
	l.logType = logType
	l.visible = true
}

// Info shows an info message
func (l *LogPanel) Info(message string) {
	l.Show(message, LogTypeInfo)
}

// Success shows a success message
func (l *LogPanel) Success(message string) {
	l.Show(message, LogTypeSuccess)
}

// Warning shows a warning message
func (l *LogPanel) Warning(message string) {
	l.Show(message, LogTypeWarning)
}

// Error shows an error message
func (l *LogPanel) Error(message string) {
	l.Show(message, LogTypeError)
}

// Loading shows a loading message
func (l *LogPanel) Loading(message string) {
	l.Show(message, LogTypeWarning)
}

// Clear hides the log panel
func (l *LogPanel) Clear() {
	l.message = ""
	l.visible = false
}

// IsVisible returns whether the log panel is visible
func (l *LogPanel) IsVisible() bool {
	return l.visible
}

// Message returns the current message
func (l *LogPanel) Message() string {
	return l.message
}

// Type returns the current log type
func (l *LogPanel) Type() LogType {
	return l.logType
}

// GetStyle returns the appropriate style for the current log type
func (l *LogPanel) GetStyle() lipgloss.Style {
	switch l.logType {
	case LogTypeSuccess:
		return GreenStyle
	case LogTypeError:
		return RedStyle
	case LogTypeWarning:
		return YellowStyle
	default:
		return GrayStyle
	}
}

// Render returns the styled log message
func (l *LogPanel) Render() string {
	if !l.visible || l.message == "" {
		return ""
	}
	return l.GetStyle().Render(l.message)
}

// RenderWithIcon returns the styled log message with an icon prefix
func (l *LogPanel) RenderWithIcon() string {
	if !l.visible || l.message == "" {
		return ""
	}
	
	icon := ""
	switch l.logType {
	case LogTypeSuccess:
		icon = "✓ "
	case LogTypeError:
		icon = "✗ "
	case LogTypeWarning:
		icon = "⟳ "
	case LogTypeInfo:
		icon = "● "
	}
	
	return l.GetStyle().Render(icon + l.message)
}

// RenderForTitle returns the message formatted for inclusion in a title bar
func (l *LogPanel) RenderForTitle() string {
	if !l.visible || l.message == "" {
		return ""
	}
	return " " + l.RenderWithIcon() + " "
}
