package task

import (
	"fmt"
	"time"

	"github.com/imbalaancing/go_final_project/internal/date"
)

type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment,omitempty"`
	Repeat  string `json:"repeat"`
}

func (t *Task) ValidateTask() error {
	if t.Title == "" {
		return fmt.Errorf("не указан заголовок задачи")
	}

	if t.Date != "" {
		if _, err := time.Parse(date.DATE_FORMAT, t.Date); err != nil {
			return fmt.Errorf("неверный формат даты")
		}
	}

	if t.Date == "" || t.Date < time.Now().Format(date.DATE_FORMAT) {
		t.Date = time.Now().Format(date.DATE_FORMAT)
	}

	if t.Repeat != "" {
		newDate, err := date.NextDate(time.Now(), t.Date, t.Repeat)
		if err != nil {
			return err
		}
		t.Date = newDate
	}

	return nil
}
