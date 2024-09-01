package main

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

// NextDate вычисляет следующую дату для задачи на основе правила повторения
func NextDate(now time.Time, date string, repeat string) (string, error) {
	if repeat == "" {
		return "", errors.New("правило повторения пустое")
	}

	// Разбираем исходную дату задачи
	taskDate, err := time.Parse("20060102", date)
	if err != nil {
		return "", fmt.Errorf("неверный формат даты: %v", err)
	}

	switch {
	case strings.HasPrefix(repeat, "d "):
		// Обработка правила "d <число>" (ежедневно)
		daysStr := strings.TrimSpace(repeat[2:])
		days, err := strconv.Atoi(daysStr)
		if err != nil || days <= 0 || days > 400 {
			return "", fmt.Errorf("неверное правило повторения 'd': %v", err)
		}

		// Увеличиваем taskDate до тех пор, пока она не станет больше 'now'
		for !taskDate.After(now) {
			taskDate = taskDate.AddDate(0, 0, days)
		}

		return taskDate.Format("20060102"), nil

	case repeat == "y":
		// Обработка правила "y" (ежегодно)
		for !taskDate.After(now) {
			year := taskDate.Year()
			month := taskDate.Month()
			day := taskDate.Day()

			// Обработка случая с високосным годом
			if month == time.February && day == 29 {
				for {
					year++
					// Найти следующую валидную дату 29 февраля или сдвинуть на 1 марта, если это не високосный год
					if isLeapYear(year) {
						taskDate = time.Date(year, month, day, 0, 0, 0, 0, taskDate.Location())
						break
					}
				}
			} else {
				taskDate = taskDate.AddDate(1, 0, 0)
			}
		}

		return taskDate.Format("20060102"), nil

	case strings.HasPrefix(repeat, "w "):
		// Обработка правила "w <дни>" (еженедельно)
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
			daysOfWeek = append(daysOfWeek, day)
		}

		sort.Ints(daysOfWeek)

		// Увеличиваем дату до следующего валидного дня недели, пока taskDate не станет больше 'now'
		for {
			currentWeekday := int(taskDate.Weekday())
			if currentWeekday == 0 {
				currentWeekday = 7
			}

			found := false
			for _, d := range daysOfWeek {
				if d > currentWeekday {
					daysUntilNext := d - currentWeekday
					taskDate = taskDate.AddDate(0, 0, daysUntilNext)
					found = true
					break
				}
			}

			if !found {
				daysUntilNext := 7 - currentWeekday + daysOfWeek[0]
				taskDate = taskDate.AddDate(0, 0, daysUntilNext)
			}

			if taskDate.After(now) {
				break
			}
		}

		return taskDate.Format("20060102"), nil

	case strings.HasPrefix(repeat, "m "):
		// Обработка правила "m <дни> [<месяцы>]" (ежемесячно)
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

		// Цикл для поиска следующей валидной даты по месячному правилу
		for {
			currentYear, currentMonth := taskDate.Year(), taskDate.Month()

			// Проверяем, находится ли текущий месяц в списке допустимых месяцев (если указан)
			if len(months) > 0 && !containsInt(months, int(currentMonth)) {
				nextMonth := findNextMonth(int(currentMonth), months)
				if nextMonth <= int(currentMonth) {
					taskDate = time.Date(currentYear+1, time.Month(nextMonth), 1, 0, 0, 0, 0, taskDate.Location())
				} else {
					taskDate = time.Date(currentYear, time.Month(nextMonth), 1, 0, 0, 0, 0, taskDate.Location())
				}
				continue
			}

			found := false
			for _, day := range daysOfMonth {
				var nextDate time.Time
				if day > 0 {
					// Положительное значение дня означает конкретный день месяца
					if day <= lastDayOfMonth(currentYear, currentMonth) {
						nextDate = time.Date(currentYear, currentMonth, day, 0, 0, 0, 0, taskDate.Location())
					} else {
						continue
					}
				} else {
					// Отрицательное значение дня означает отсчет дней с конца месяца
					lastDay := lastDayOfMonth(currentYear, currentMonth)
					if -day <= lastDay {
						nextDate = time.Date(currentYear, currentMonth, lastDay+day+1, 0, 0, 0, 0, taskDate.Location())
					} else {
						continue
					}
				}

				if nextDate.After(now) {
					taskDate = nextDate
					found = true
					break
				}
			}

			if found {
				break
			}

			// Переходим к следующему месяцу, если в текущем месяце нет валидной даты
			taskDate = taskDate.AddDate(0, 1, 0)
			// Сброс дня на 1 для начала проверки с начала следующего месяца
			taskDate = time.Date(taskDate.Year(), taskDate.Month(), 1, 0, 0, 0, 0, taskDate.Location())
		}

		return taskDate.Format("20060102"), nil

	default:
		return "", fmt.Errorf("неподдерживаемый формат правила повторения: '%s'", repeat)
	}
}

// lastDayOfMonth возвращает последний день указанного месяца и года
func lastDayOfMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

// containsInt проверяет, содержит ли срез определенное число
func containsInt(slice []int, value int) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

// findNextMonth находит следующий месяц в отсортированном списке, который идет после или равен текущему месяцу
func findNextMonth(currentMonth int, months []int) int {
	for _, month := range months {
		if month >= currentMonth {
			return month
		}
	}
	return months[0] // Если ничего не найдено, возвращаем первый месяц в отсортированном списке
}

// isLeapYear определяет, является ли год високосным
func isLeapYear(year int) bool {
	if year%4 == 0 && (year%100 != 0 || year%400 == 0) {
		return true
	}
	return false
}
