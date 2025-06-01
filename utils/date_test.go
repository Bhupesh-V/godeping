package utils

import (
	"testing"
	"time"
)

func TestGetTimeDurationFromRelativeDate(t *testing.T) {
	// Library constants to use in calculations
	const (
		hoursPerDay   = 24
		hoursPerWeek  = hoursPerDay * 7
		hoursPerMonth = 365.0 * hoursPerDay / 12 // ~30.42 days
		hoursPerYear  = 365 * hoursPerDay
	)

	tests := []struct {
		name           string
		input          string
		expectedOutput time.Duration
		expectError    bool
	}{
		// Valid inputs with single units
		{
			name:           "Years only",
			input:          "2y",
			expectedOutput: time.Duration(2 * hoursPerYear * float64(time.Hour)),
			expectError:    false,
		},
		{
			name:           "Months only",
			input:          "6m",
			expectedOutput: time.Duration(6 * hoursPerMonth * float64(time.Hour)),
			expectError:    false,
		},
		{
			name:           "Weeks only",
			input:          "4w",
			expectedOutput: time.Duration(4 * hoursPerWeek * float64(time.Hour)),
			expectError:    false,
		},
		{
			name:           "Days only",
			input:          "10d",
			expectedOutput: time.Duration(10 * hoursPerDay * float64(time.Hour)),
			expectError:    false,
		},

		// Valid inputs with multiple units
		{
			name:           "Years and months",
			input:          "1y2m",
			expectedOutput: time.Duration((hoursPerYear + 2*hoursPerMonth) * float64(time.Hour)),
			expectError:    false,
		},
		{
			name:           "Months and days",
			input:          "3m15d",
			expectedOutput: time.Duration((3*hoursPerMonth + 15*hoursPerDay) * float64(time.Hour)),
			expectError:    false,
		},
		{
			name:           "Years, months, and days",
			input:          "1y3m5d",
			expectedOutput: time.Duration((hoursPerYear + 3*hoursPerMonth + 5*hoursPerDay) * float64(time.Hour)),
			expectError:    false,
		},
		{
			name:           "Full complex duration",
			input:          "2y6m1w3d",
			expectedOutput: time.Duration((2*hoursPerYear + 6*hoursPerMonth + hoursPerWeek + 3*hoursPerDay) * float64(time.Hour)),
			expectError:    false,
		},

		// Case insensitivity tests
		{
			name:           "Lowercase input",
			input:          "1y3m",
			expectedOutput: time.Duration((hoursPerYear + 3*hoursPerMonth) * float64(time.Hour)),
			expectError:    false,
		},
		{
			name:           "Mixed case input",
			input:          "1Y3m",
			expectedOutput: time.Duration((hoursPerYear + 3*hoursPerMonth) * float64(time.Hour)),
			expectError:    false,
		},

		// Edge cases
		{
			name:           "Zero duration",
			input:          "0y0m0d",
			expectedOutput: 0,
			expectError:    false,
		},
		{
			name:           "Single day",
			input:          "1d",
			expectedOutput: time.Duration(hoursPerDay * float64(time.Hour)),
			expectError:    false,
		},

		// Error cases
		{
			name:        "Empty string",
			input:       "",
			expectError: true,
		},
		{
			name:        "Invalid format",
			input:       "abc",
			expectError: true,
		},
		// Renabled once https://github.com/sosodev/duration/issues/35 is fixed
		// {
		// 	name:        "Missing unit",
		// 	input:       "1",
		// 	expectError: true,
		// },
		{
			name:        "Invalid unit",
			input:       "1x",
			expectError: true,
		},
		{
			name:        "Invalid character in duration",
			input:       "1y-2m",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetTimeDurationFromRelativeDate(tt.input)

			// Check error expectation
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none in test: %s", tt.name)
				}
				return
			}

			// Check no error when not expected
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Check duration value
			if result != tt.expectedOutput {
				t.Errorf("Expected duration %v, got %v in %s", tt.expectedOutput, result, tt.name)
			}
		})
	}
}

// Test specific edge cases related to library behavior
func TestGetTimeDurationFromRelativeDate_SpecificCases(t *testing.T) {
	// Define the library's exact constants
	const (
		hoursPerDay   = 24
		hoursPerWeek  = hoursPerDay * 7
		hoursPerMonth = 365.0 * hoursPerDay / 12
		hoursPerYear  = 365 * hoursPerDay
	)

	tests := []struct {
		name        string
		input       string
		shouldEqual bool
		comparison  time.Duration
	}{
		{
			name:        "Year approximation check",
			input:       "1y",
			shouldEqual: true,
			comparison:  time.Duration(hoursPerYear * float64(time.Hour)), // Verify library uses 365 days for a year
		},
		{
			name:        "Month approximation check",
			input:       "1m",
			shouldEqual: true,
			comparison:  time.Duration(hoursPerMonth * float64(time.Hour)), // Verify library uses ~30.42 days for a month
		},
		{
			name:        "Week approximation check",
			input:       "1w",
			shouldEqual: true,
			comparison:  time.Duration(hoursPerWeek * float64(time.Hour)), // Verify library uses 7 days for a week
		},
		{
			name:        "Month is not exactly 30 days",
			input:       "1m",
			shouldEqual: false,
			comparison:  30 * 24 * time.Hour, // Library doesn't use exactly 30 days for a month
		},
		{
			name:        "Combination precision test",
			input:       "1y1m1w1d",
			shouldEqual: true,
			comparison:  time.Duration((hoursPerYear + hoursPerMonth + hoursPerWeek + hoursPerDay) * float64(time.Hour)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetTimeDurationFromRelativeDate(tt.input)

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if tt.shouldEqual && result != tt.comparison {
				t.Errorf("Expected duration %v, got %v", tt.comparison, result)
			}

			if !tt.shouldEqual && result == tt.comparison {
				t.Errorf("Expected duration to be different from %v", tt.comparison)
			}
		})
	}
}
