package tui

// ScrollbarConfig holds configuration for rendering a scrollbar
type ScrollbarConfig struct {
	TotalItems   int // Total number of items
	VisibleItems int // Number of visible items
	ScrollOffset int // Current scroll position
	Height       int // Height of the scrollbar in characters
}

// ScrollbarChars defines the characters used for the scrollbar
var ScrollbarChars = struct {
	Up     string
	Down   string
	Track  string
	Thumb  string
	Empty  string
}{
	Up:    "▲",
	Down:  "▼",
	Track: "│",
	Thumb: "┃",
	Empty: " ",
}

// RenderScrollbar returns a slice of characters representing the scrollbar
// Each element corresponds to one row of the scrollbar
func RenderScrollbar(cfg ScrollbarConfig) []string {
	if cfg.Height <= 0 || cfg.TotalItems <= cfg.VisibleItems {
		// No scrollbar needed - return empty spaces
		result := make([]string, cfg.Height)
		for i := range result {
			result[i] = ScrollbarChars.Empty
		}
		return result
	}

	result := make([]string, cfg.Height)

	// Calculate thumb size proportional to visible/total ratio
	thumbSize := (cfg.VisibleItems * cfg.Height) / cfg.TotalItems
	if thumbSize < 1 {
		thumbSize = 1
	}
	if thumbSize > cfg.Height {
		thumbSize = cfg.Height
	}

	// Calculate max scroll and thumb position
	maxScroll := cfg.TotalItems - cfg.VisibleItems
	if maxScroll <= 0 {
		maxScroll = 1
	}

	// Thumb position: 0 at top, height-thumbSize at bottom
	var thumbPos int
	if cfg.ScrollOffset >= maxScroll {
		// At bottom - thumb at the very bottom
		thumbPos = cfg.Height - thumbSize
	} else if cfg.ScrollOffset <= 0 {
		// At top - thumb at the very top
		thumbPos = 0
	} else {
		// In between - proportional position
		thumbPos = (cfg.ScrollOffset * (cfg.Height - thumbSize)) / maxScroll
	}

	// Render each row
	for i := 0; i < cfg.Height; i++ {
		if i >= thumbPos && i < thumbPos+thumbSize {
			// Thumb position
			result[i] = ScrollbarChars.Thumb
		} else if i < thumbPos {
			// Track above thumb - show up arrow at top if can scroll up
			if i == 0 && cfg.ScrollOffset > 0 {
				result[i] = ScrollbarChars.Up
			} else {
				result[i] = ScrollbarChars.Track
			}
		} else {
			// Track below thumb - show down arrow at bottom if can scroll down
			if i == cfg.Height-1 && cfg.ScrollOffset < maxScroll {
				result[i] = ScrollbarChars.Down
			} else {
				result[i] = ScrollbarChars.Track
			}
		}
	}

	return result
}

// GetScrollChar returns the appropriate scrollbar character for a given row index
// This is a convenience function for row-by-row rendering
func GetScrollChar(cfg ScrollbarConfig, rowIndex int) string {
	scrollbar := RenderScrollbar(cfg)
	if rowIndex >= 0 && rowIndex < len(scrollbar) {
		return scrollbar[rowIndex]
	}
	return ScrollbarChars.Empty
}

// NeedsScrollbar returns true if a scrollbar is needed
func NeedsScrollbar(totalItems, visibleItems int) bool {
	return totalItems > visibleItems
}
