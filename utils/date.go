package utils

import (
	"fmt"
	"strings"
	"time"

	iso8601duration "github.com/sosodev/duration"
)

// parseDuration parses a duration string like "1y2m" into a time.Duration
func GetTimeDurationFromRelativeDate(s string) (time.Duration, error) {
	// append Prefix "P" to the string to make it ISO 8601 compliant
	if s == "" {
		return 0, fmt.Errorf("duration cannot be empty")
	}

	s = strings.ToUpper("P" + s)

	d, err := iso8601duration.Parse(s)
	if err != nil {
		return 0, fmt.Errorf("failed to parse duration: %w", err)
	}

	return d.ToTimeDuration(), nil
}
