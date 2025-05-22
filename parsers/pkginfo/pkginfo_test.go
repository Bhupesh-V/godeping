package pkginfo

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestExtractPublishDate(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected time.Time
	}{
		{
			name:     "No date in HTML",
			html:     "<html><body>No date here</body></html>",
			expected: time.Time{},
		},
		{
			name:     "Standard date format",
			html:     `<span data-test-id="UnitHeader-commitTime">Jan 15, 2023</span>`,
			expected: time.Date(2023, time.January, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "Date with 'Published:' prefix",
			html:     `<span data-test-id="UnitHeader-commitTime">Published: Feb 5, 2024</span>`,
			expected: time.Date(2024, time.February, 5, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "Full month name date",
			html:     `<span data-test-id="UnitHeader-commitTime">January 10, 2022</span>`,
			expected: time.Date(2022, time.January, 10, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "Single digit day",
			html:     `<span data-test-id="UnitHeader-commitTime">Mar 2, 2021</span>`,
			expected: time.Date(2021, time.March, 2, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractPublishDate(tt.html)
			assert.Equal(t, tt.expected, result)
		})
	}
}
