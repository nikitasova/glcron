package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Footer configuration
const (
	FooterDefaultGap = " │ " // Separator between footer items
)

// FooterItem represents a single hotkey item in the footer
type FooterItem struct {
	Key         string
	Description string
}

// Footer manages the footer/legend display for all screens
type Footer struct {
	Gap string // Separator between items
}

// NewFooter creates a new footer instance
func NewFooter() *Footer {
	return &Footer{
		Gap: FooterDefaultGap,
	}
}

// GetItems returns the footer items for a given screen
func (f *Footer) GetItems(screen Screen) []FooterItem {
	switch screen {
	case ScreenConfigList:
		return []FooterItem{
			{Key: "↑↓", Description: "Navigate"},
			{Key: "Enter", Description: "Select"},
			{Key: "c", Description: "Create"},
			{Key: "e", Description: "Edit"},
			{Key: "d", Description: "Delete"},
			{Key: "h", Description: "Help"},
			{Key: "q", Description: "Quit"},
		}

	case ScreenScheduleList:
		return []FooterItem{
			// {Key: "↑↓", Description: "Navigate"},
			{Key: "/", Description: "Search"},
			// {Key: "e", Description: "Edit"},
			{Key: "c", Description: "Create"},
			{Key: "y", Description: "Yonk"},
			{Key: "d", Description: "Delete"},
			{Key: "r", Description: "Run Pipeline"},
			{Key: "R", Description: "Quick Run"},
			{Key: "A", Description: "Toggle"},
			{Key: "t", Description: "Take ownership"},
			{Key: "u", Description: "Update"},
			{Key: "o", Description: "Configs"},
			{Key: "h", Description: "Help"},
			{Key: "q", Description: "Quit"},
		}

	case ScreenEditSchedule, ScreenNewSchedule:
		return []FooterItem{
			{Key: "↑↓", Description: "Navigate"},
			{Key: "Enter", Description: "Select/Toggle"},
			{Key: "Ctrl+S", Description: "Save"},
			{Key: "h", Description: "Help"},
			{Key: "Esc", Description: "Cancel"},
		}

	case ScreenEditConfig, ScreenNewConfig:
		return []FooterItem{
			{Key: "↑↓", Description: "Navigate"},
			{Key: "Tab", Description: "Next"},
			{Key: "Ctrl+S", Description: "Save"},
			{Key: "h", Description: "Help"},
			{Key: "Esc", Description: "Cancel"},
		}

	case ScreenQuickRun:
		return []FooterItem{
			{Key: "R", Description: "New Run"},
			{Key: "u", Description: "Update"},
			{Key: "↑↓", Description: "Navigate"},
			{Key: "h", Description: "Help"},
			{Key: "Esc", Description: "Back"},
			{Key: "q", Description: "Quit"},
		}

	default:
		return []FooterItem{
			{Key: "h", Description: "Help"},
			{Key: "q", Description: "Quit"},
		}
	}
}

// Render renders the footer for a given screen, centered within the given width
func (f *Footer) Render(screen Screen, width int) string {
	items := f.GetItems(screen)
	return f.RenderItems(items, width)
}

// RenderItems renders a list of footer items, centered within the given width
func (f *Footer) RenderItems(items []FooterItem, width int) string {
	var parts []string
	for _, item := range items {
		parts = append(parts, YellowStyle.Render(item.Key)+" "+item.Description)
	}

	gap := f.Gap
	if gap == "" {
		gap = FooterDefaultGap
	}
	legend := strings.Join(parts, gap)

	// Center the legend
	legendWidth := lipgloss.Width(legend)
	if legendWidth >= width {
		return legend
	}
	padding := (width - legendWidth) / 2
	return strings.Repeat(" ", padding) + legend
}

// RenderCustom renders custom footer items (for special cases)
func (f *Footer) RenderCustom(items []FooterItem, width int) string {
	return f.RenderItems(items, width)
}

// Common footer item sets that can be reused
var (
	// NavigationItems - basic navigation keys
	NavigationItems = []FooterItem{
		{Key: "↑↓", Description: "Navigate"},
	}

	// FormItems - common form navigation
	FormItems = []FooterItem{
		{Key: "↑↓", Description: "Navigate"},
		{Key: "Tab", Description: "Next"},
		{Key: "Enter", Description: "Select"},
	}

	// SaveCancelItems - save/cancel actions
	SaveCancelItems = []FooterItem{
		{Key: "Ctrl+S", Description: "Save"},
		{Key: "Esc", Description: "Cancel"},
	}

	// HelpQuitItems - help and quit
	HelpQuitItems = []FooterItem{
		{Key: "h", Description: "Help"},
		{Key: "q", Description: "Quit"},
	}

	// BackQuitItems - back and quit
	BackQuitItems = []FooterItem{
		{Key: "Esc", Description: "Back"},
		{Key: "q", Description: "Quit"},
	}
)

// CombineItems combines multiple footer item slices into one
func CombineItems(itemSets ...[]FooterItem) []FooterItem {
	var result []FooterItem
	for _, items := range itemSets {
		result = append(result, items...)
	}
	return result
}
