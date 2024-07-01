package task

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const DATE_FORMAT = "20060102"

type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment,omitempty"`
	Repeat  string `json:"repeat"`
}

func NextDate(now time.Time, date string, repeat string) (string, error) {
	if repeat == "" {
		return "", nil
	}

	startDate, err := time.Parse(DATE_FORMAT, date)
	if err != nil {
		return "", fmt.Errorf("неверный формат даты: %v", err)
	}

	parts := strings.Split(repeat, " ")
	param := parts[0]
	switch param {
	case "y":
		currDate := startDate.AddDate(1, 0, 0)
		for now.After(currDate) || now.Equal(currDate) {
			currDate = currDate.AddDate(1, 0, 0)
		}
		return currDate.Format(DATE_FORMAT), nil

	case "d":
		if len(parts) == 1 {
			return "", fmt.Errorf("не указан интервал в днях")
		}
		days, err := strconv.Atoi(parts[1])
		if err != nil || days < 1 || days > 400 {
			return "", fmt.Errorf("неверный формат интервала в днях: %v", err)
		}
		currDate := startDate.AddDate(0, 0, days)
		for now.After(currDate) {
			currDate = currDate.AddDate(0, 0, days)
		}
		return currDate.Format(DATE_FORMAT), nil

	default:
		return "", fmt.Errorf("неподдерживаемый формат %s", param)
	}
}

func (t *Task) ValidateTask() error {
	if t.Title == "" {
		return fmt.Errorf("не указан заголовок задачи")
	}

	if t.Date != "" {
		if _, err := time.Parse(DATE_FORMAT, t.Date); err != nil {
			return fmt.Errorf("неверный формат даты")
		}
	}

	if t.Date == "" || t.Date < time.Now().Format(DATE_FORMAT) {
		t.Date = time.Now().Format(DATE_FORMAT)
	}

	if t.Repeat != "" {
		newDate, err := NextDate(time.Now(), t.Date, t.Repeat)
		if err != nil {
			return err
		}
		t.Date = newDate
	}

	return nil
}
