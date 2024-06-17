package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/imbalaancing/go_final_project/internal/task"
)

type Task struct {
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment,omitempty"`
	Repeat  string `json:"repeat"`
}

func AddTaskHandler(w http.ResponseWriter, r *http.Request, database *sql.DB) {
	var t Task
	err := json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if t.Title == "" {
		http.Error(w, `{"error":"Title is required"}`, http.StatusBadRequest)
		return
	}

	if t.Date != "" {
		_, err = time.Parse("20060102", t.Date)
		if err != nil {
			http.Error(w, `{"error":"Invalid date format"}`, http.StatusBadRequest)
			return
		}
	}

	if t.Date == "" || t.Date < time.Now().Format("20060102") {
		t.Date = time.Now().Format("20060102")
	}

	if t.Repeat == "d 1" {
		t.Date = time.Now().Format("20060102")
	} else if t.Repeat != "" {
		t.Date, err = task.NextDate(time.Now(), t.Date, t.Repeat)
		if err != nil {
			http.Error(w, `{"error":"Invalid repeat rule"}`, http.StatusBadRequest)
			return
		}
	}

	res, err := database.Exec(`INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`,
		t.Date, t.Title, t.Comment, t.Repeat)
	if err != nil {
		http.Error(w, `{"error":"Failed to insert task"}`, http.StatusInternalServerError)
		return
	}

	id, err := res.LastInsertId()
	if err != nil {
		http.Error(w, `{"error":"Failed to get task ID"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(map[string]string{"id": strconv.FormatInt(id, 10)})

}

// Обработчик для маршрута api/nextdate
func NextDateHandler(w http.ResponseWriter, r *http.Request) {
	nowStr := r.FormValue("now")
	dateStr := r.FormValue("date")
	repeat := r.FormValue("repeat")

	if nowStr == "" || dateStr == "" || repeat == "" {
		http.Error(w, "Missing required query parameters", http.StatusBadRequest)
		return
	}

	now, err := time.Parse(task.DATE_FORMAT, nowStr)
	if err != nil {
		http.Error(w, "Invalid 'now' date format", http.StatusBadRequest)
		return
	}

	nextDate, err := task.NextDate(now, dateStr, repeat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Write([]byte(nextDate))
}
