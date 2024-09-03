package main

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

// DateFormat is the constant format for dates used throughout the application
const DateFormat = "20060102"

// NextDate computes the next date for a task based on a repeat rule
func NextDate(now time.Time, date string, repeat string) (string, error) {
	if repeat == "" {
		return "", errors.New("правило повторения пустое")
	}

	// Parse the initial task date using the DateFormat constant
	taskDate, err := time.Parse(DateFormat, date)
	if err != nil {
		return "", fmt.Errorf("неверный формат даты: %v", err)
	}

	// Start calculation from taskDate or now, whichever is later
	startDate := now
	if taskDate.After(now) {
		startDate = taskDate
	}

	switch {
	case strings.HasPrefix(repeat, "d "): // Handle daily repetition rule "d <number>"
		daysStr := strings.TrimSpace(repeat[2:])
		days, err := strconv.Atoi(daysStr)
		if err != nil || days <= 0 || days > 400 {
			return "", fmt.Errorf("неверное правило повторения 'd': %v", err)
		}

		// Increment taskDate by specified number of days until it exceeds both 'now' and 'date'
		for !taskDate.After(startDate) {
			taskDate = taskDate.AddDate(0, 0, days)
		}

		return taskDate.Format(DateFormat), nil

	case repeat == "y": // Handle yearly repetition rule "y"
		// Adjusted logic to directly calculate next yearly occurrence without checking if taskDate is after now
		for !taskDate.After(startDate) {
			year := taskDate.Year() + 1
			month := taskDate.Month()
			day := taskDate.Day()

			// Handle leap year scenario
			if month == time.February && day == 29 && !isLeapYear(year) {
				taskDate = time.Date(year, time.March, 1, 0, 0, 0, 0, taskDate.Location())
			} else {
				taskDate = time.Date(year, month, day, 0, 0, 0, 0, taskDate.Location())
			}
		}

		return taskDate.Format(DateFormat), nil

	case strings.HasPrefix(repeat, "w "): // Handle weekly repetition rule "w <days>"
		daysStr := strings.TrimSpace(repeat[2:])
		days := strings.Split(daysStr, ",")
		if len(days) == 0 {
			return "", fmt.Errorf("неверное правило повторения 'w': дни не указаны")
		}

		var daysOfWeek []int
		for _, dayStr := range days {
			day, err := strconv.Atoi(dayStr)
			if err != nil || day < 1 || day > 7 {
				return "", fmt.Errorf("неверное правило повторения 'w': неверный день '%s'", dayStr)
			}
			if day == 7 {
				day = 0
			}
			daysOfWeek = append(daysOfWeek, day)
		}
		sort.Ints(daysOfWeek)

		// Set initial startDate to the maximum of taskDate and now
		startDate = taskDate
		if now.After(taskDate) {
			startDate = now
		}

		initialDate := taskDate

		// Loop until the weekday matches one in the daysOfWeek
		for !containsInt(daysOfWeek, int(startDate.Weekday())) || !(startDate.YearDay() > initialDate.YearDay()) { // Weekday+1 to adjust for Monday = 1, Sunday = 7
			startDate = startDate.AddDate(0, 0, 1)
		}
		return startDate.Format(DateFormat), nil

	case strings.HasPrefix(repeat, "m "): // Handle monthly repetition rule "m <days> [<months>]"
		parts := strings.Split(strings.TrimSpace(repeat[2:]), " ")
		if len(parts) == 0 {
			return "", fmt.Errorf("неверное правило повторения 'm': дни не указаны")
		}

		dayParts := strings.Split(parts[0], ",")
		var daysOfMonth []int
		for _, dayStr := range dayParts {
			day, err := strconv.Atoi(dayStr)
			if err != nil || day == 0 || day < -31 || day > 31 {
				return "", fmt.Errorf("неверное правило повторения 'm': неверный день '%s'", dayStr)
			}
			daysOfMonth = append(daysOfMonth, day)
		}

		var months []int
		if len(parts) > 1 {
			monthParts := strings.Split(parts[1], ",")
			for _, monthStr := range monthParts {
				month, err := strconv.Atoi(monthStr)
				if err != nil || month < 1 || month > 12 {
					return "", fmt.Errorf("неверное правило повторения 'm': неверный месяц '%s'", monthStr)
				}
				months = append(months, month)
			}
		}

		sort.Ints(daysOfMonth)
		sort.Ints(months)

		for {
			currentYear, currentMonth := taskDate.Year(), taskDate.Month()

			if len(months) > 0 && !containsInt(months, int(currentMonth)) {
				nextMonth := findNextMonth(int(currentMonth), months)
				if nextMonth <= int(currentMonth) {
					taskDate = time.Date(currentYear+1, time.Month(nextMonth), 1, 0, 0, 0, 0, taskDate.Location())
				} else {
					taskDate = time.Date(currentYear, time.Month(nextMonth), 1, 0, 0, 0, 0, taskDate.Location())
				}
				continue
			}

			// Find the earliest valid date in the current month
			var nextValidDate time.Time
			for _, day := range daysOfMonth {
				var candidateDate time.Time
				lastDay := lastDayOfMonth(currentYear, currentMonth)

				if day > 0 {
					if day <= lastDay {
						candidateDate = time.Date(currentYear, currentMonth, day, 0, 0, 0, 0, taskDate.Location())
					} else {
						continue
					}
				} else {
					if -day <= lastDay {
						candidateDate = time.Date(currentYear, currentMonth, lastDay+day+1, 0, 0, 0, 0, taskDate.Location())
					} else {
						continue
					}
				}

				// Find the closest date after startDate
				if candidateDate.After(startDate) && (nextValidDate.IsZero() || candidateDate.Before(nextValidDate)) {
					nextValidDate = candidateDate
				}
			}

			// If a valid date is found, return it
			if !nextValidDate.IsZero() {
				return nextValidDate.Format(DateFormat), nil
			}

			// If no valid date is found in the current month, move to the next month
			taskDate = taskDate.AddDate(0, 1, 0)
			taskDate = time.Date(taskDate.Year(), taskDate.Month(), 1, 0, 0, 0, 0, taskDate.Location())
		}

	default:
		return "", fmt.Errorf("неподдерживаемый формат правила повторения: '%s'", repeat)
	}
}

// lastDayOfMonth returns the last day of the specified month and year
func lastDayOfMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

// containsInt checks if a slice contains a specific integer
func containsInt(slice []int, value int) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

// findNextMonth finds the next month in a sorted list that comes after or is equal to the current month
func findNextMonth(currentMonth int, months []int) int {
	for _, month := range months {
		if month >= currentMonth {
			return month
		}
	}
	return months[0] // If nothing is found, return the first month in the sorted list
}

// isLeapYear determines if a year is a leap year
func isLeapYear(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}
