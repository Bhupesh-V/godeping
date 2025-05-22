package pkginfo

import (
	"regexp"
	"strings"
	"time"
)

// ExtractPublishDate extracts the last published date from pkg.go.dev HTML
func ExtractPublishDate(html string) time.Time {
	// Look for the published date in the span with data-test-id="UnitHeader-commitTime"
	datePattern := regexp.MustCompile(`<span[^>]*data-test-id="UnitHeader-commitTime"[^>]*>([^<]+)</span>`)
	matches := datePattern.FindStringSubmatch(html)

	if len(matches) < 2 {
		return time.Time{} // Return zero time if not found
	}

	// Parse the date string (format: "Jan 23, 2024")
	dateStr := strings.TrimSpace(matches[1])
	dateStr = strings.TrimPrefix(dateStr, "Published:")
	dateStr = strings.TrimSpace(dateStr)

	// Try different formats as the exact format might vary
	formats := []string{
		"Jan 2, 2006",
		"Jan 02, 2006",
		"January 2, 2006",
		"January 02, 2006",
	}

	for _, format := range formats {
		date, err := time.Parse(format, dateStr)
		if err == nil {
			return date
		}
	}

	return time.Time{} // Return zero time if parsing fails
}
