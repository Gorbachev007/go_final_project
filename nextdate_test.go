package main

import (
	"testing"
	"time"
)

// Helper function to parse date strings for convenience in tests
func parseDate(dateStr string) time.Time {
	date, _ := time.Parse("20060102", dateStr)
	return date
}

// TestNextDate tests the NextDate function with various repeat rules
func TestNextDate(t *testing.T) {
	tests := []struct {
		now        string
		date       string
		repeat     string
		expected   string
		shouldFail bool
	}{
		// Yearly repeat tests
		{"20250701", "20250701", "y", "20260701", false},
		{"20240229", "20240229", "y", "20250301", false},
		{"20240301", "20240301", "y", "20250301", false},
		{"20201231", "20201231", "y", "20211231", false},

		// Daily repeat tests
		{"20240202", "20240202", "d 30", "20240303", false},
		{"20240228", "20240228", "d 1", "20240229", false},
		{"20240227", "20240227", "d 1", "20240228", false},
		{"20240228", "20240228", "d 365", "20250227", false},

		// Weekly repeat tests
		{"20230101", "20230101", "w 1,3,5", "20230102", false},
		{"20230102", "20230101", "w 1,3,5", "20230104", false},
		{"20230105", "20230101", "w 1,3,5", "20230106", false},
		{"20230107", "20230107", "w 2,4,6", "20230110", false},

		// Monthly repeat tests
		{"20240228", "20240228", "m 28", "20240328", false},
		{"20240228", "20240228", "m 31", "20240331", false},
		{"20240228", "20240228", "m -1", "20240229", false},
		{"20240222", "20240222", "m -2,-3", "20240227", false},
		{"20240201", "20240201", "m -1,18", "20240218", false},

		// Edge cases
		{"20240101", "20201231", "y", "20241231", false},   // Incorrect date handling (shouldFail true)
		{"20240301", "20240228", "d 1", "20240302", false}, // End of month daily increment
		{"20240301", "20240229", "y", "20250301", false},   // Leap year to non-leap year transition
	}

	for _, test := range tests {
		now := parseDate(test.now)
		date := test.date
		repeat := test.repeat
		expected := test.expected

		result, err := NextDate(now, date, repeat)
		if err != nil && !test.shouldFail {
			t.Errorf("Unexpected error for inputs (%s, %s, %s): %v", test.now, test.date, test.repeat, err)
		}
		if result != expected && !test.shouldFail {
			t.Errorf("For inputs (%s, %s, %s), expected %s, but got %s", test.now, test.date, test.repeat, expected, result)
		}
		if err == nil && test.shouldFail {
			t.Errorf("Expected failure for inputs (%s, %s, %s), but got result: %s", test.now, test.date, test.repeat, result)
		}
	}
}

// TestLeapYear checks that the isLeapYear function is correct
func TestLeapYear(t *testing.T) {
	leapYears := []int{2000, 2004, 2008, 2012, 2016, 2020, 2024}
	nonLeapYears := []int{1900, 2001, 2002, 2003, 2005, 2100}

	for _, year := range leapYears {
		if !isLeapYear(year) {
			t.Errorf("Year %d should be a leap year but was not detected as one.", year)
		}
	}

	for _, year := range nonLeapYears {
		if isLeapYear(year) {
			t.Errorf("Year %d should not be a leap year but was detected as one.", year)
		}
	}
}

// TestLastDayOfMonth checks that the lastDayOfMonth function is correct
func TestLastDayOfMonth(t *testing.T) {
	daysInMonth := map[int]int{
		1:  31,
		2:  28,
		3:  31,
		4:  30,
		5:  31,
		6:  30,
		7:  31,
		8:  31,
		9:  30,
		10: 31,
		11: 30,
		12: 31,
	}

	for month, expected := range daysInMonth {
		if lastDayOfMonth(2023, time.Month(month)) != expected {
			t.Errorf("Expected %d days in month %d, but got %d", expected, month, lastDayOfMonth(2023, time.Month(month)))
		}
	}

	// Test February in a leap year
	if lastDayOfMonth(2024, time.February) != 29 {
		t.Errorf("Expected 29 days in February 2024, but got %d", lastDayOfMonth(2024, time.February))
	}
}
