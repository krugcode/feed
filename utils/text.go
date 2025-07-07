package utils

import (
	"fmt"
	"time"
)

func FormatDate(dateStr string) string {
	if dateStr == "" {
		return ""
	}

	// Try multiple date formats that PocketBase might use
	formats := []string{
		time.RFC3339,
		"2006-01-02 15:04:05.999Z",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05Z",
	}

	var t time.Time
	var err error

	for _, format := range formats {
		t, err = time.Parse(format, dateStr)
		if err == nil {
			break
		}
	}

	if err != nil {
		return dateStr // Return original if parsing fails
	}

	// Format as desired - customize this to your preference
	return t.Format("Jan 2, 2006 at 3:04 PM")
}

func FormatRelativeDate(dateStr string) string {
	if dateStr == "" {
		return ""
	}

	t, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		return dateStr
	}

	now := time.Now()
	diff := now.Sub(t)

	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		minutes := int(diff.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	case diff < 7*24*time.Hour:
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	case diff < 30*24*time.Hour:
		weeks := int(diff.Hours() / (24 * 7))
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	default:
		return t.Format("Jan 2, 2006")
	}
}

func FormatShortDate(dateStr string) string {
	if dateStr == "" {
		return ""
	}

	t, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		return dateStr
	}

	return t.Format("Jan 2, 2006")
}

func FormatTime(dateStr string) string {
	if dateStr == "" {
		return ""
	}

	t, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		return dateStr
	}

	return t.Format("3:04 PM")
}

func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}

	// Handle UTF-8 properly
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}

	return string(runes[:maxLen-3]) + "..."
}
