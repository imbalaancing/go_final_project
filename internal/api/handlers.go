package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/imbalaancing/go_final_project/internal/task"
)

type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment,omitempty"`
	Repeat  string `json:"repeat"`
}

func MarkTaskDoneHandler(w http.ResponseWriter, r *http.Request, database *sql.DB) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, `{"error":"Не указан идентификатор"}`, http.StatusBadRequest)
		return
	}

	var t Task
	err := database.QueryRow(`SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?`, id).Scan(&t.ID, &t.Date, &t.Title, &t.Comment, &t.Repeat)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, `{"error":"Задача не найдена"}`, http.StatusNotFound)
		} else {
			http.Error(w, `{"error":"Ошибка получения задачи"}`, http.StatusInternalServerError)
		}
		return
	}

	if t.Repeat == "" {
		_, err = database.Exec(`DELETE FROM scheduler WHERE id = ?`, id)
	} else {
		newDate, err := task.NextDate(time.Now(), t.Date, t.Repeat)
		if err != nil {
			http.Error(w, `{"error":"Ошибка расчета следующей даты"}`, http.StatusInternalServerError)
			return
		}
		_, err = database.Exec(`UPDATE scheduler SET date = ? WHERE id = ?`, newDate, id)
	}
	if err != nil {
		http.Error(w, `{"error":"Ошибка обновления задачи"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(map[string]string{})
}

func DeleteTaskHandler(w http.ResponseWriter, r *http.Request, database *sql.DB) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, `{"error":"Не указан идентификатор"}`, http.StatusBadRequest)
		return
	}

	res, err := database.Exec(`DELETE FROM scheduler WHERE id = ?`, id)
	if err != nil {
		http.Error(w, `{"error":"Failed to delete task"}`, http.StatusInternalServerError)
		return
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		http.Error(w, `{"error":"Failed to get affected rows"}`, http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, `{"error":"Задача не найдена"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if err = json.NewEncoder(w).Encode(map[string]string{}); err != nil {
		http.Error(w, `{"error":"Failed to encode response"}`, http.StatusInternalServerError)
	}
}

func GetTaskHandler(w http.ResponseWriter, r *http.Request, database *sql.DB) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, `{"error":"Не указан идентификатор"}`, http.StatusBadRequest)
		return
	}

	row := database.QueryRow(`SELECT * FROM scheduler WHERE id = ?`, id)
	var task Task
	err := row.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, `{"error":"Задача не найдена"}`, http.StatusNotFound)
		} else {
			http.Error(w, `{"error":"Failed to get task"}`, http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if err = json.NewEncoder(w).Encode(task); err != nil {
		http.Error(w, `{"error":"Failed to encode task"}`, http.StatusInternalServerError)
	}
}

func UpdateTaskHandler(w http.ResponseWriter, r *http.Request, database *sql.DB) {
	var t Task
	err := json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		http.Error(w, `{"error":"Ошибка десериализации JSON"}`, http.StatusBadRequest)
		return
	}

	if t.ID == "" {
		http.Error(w, `{"error":"Не указан идентификатор"}`, http.StatusBadRequest)
		return
	}

	if t.Title == "" {
		http.Error(w, `{"error":"Не указан заголовок задачи"}`, http.StatusBadRequest)
		return
	}

	if t.Date != "" {
		_, err = time.Parse("20060102", t.Date)
		if err != nil {
			http.Error(w, `{"error":"Дата представлена в неверном формате"}`, http.StatusBadRequest)
			return
		}
	}

	if t.Date == "" || t.Date < time.Now().Format("20060102") {
		t.Date = time.Now().Format("20060102")
	}

	if t.Repeat != "" {
		t.Date, err = task.NextDate(time.Now(), t.Date, t.Repeat)
		if err != nil {
			http.Error(w, `{"error":"Неподдерживаемый формат повторения"}`, http.StatusBadRequest)
			return
		}
	}

	res, err := database.Exec(`UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?`, t.Date, t.Title, t.Comment, t.Repeat, t.ID)
	if err != nil {
		http.Error(w, `{"error":"Ошибка обновления задачи"}`, http.StatusInternalServerError)
		return
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil || rowsAffected == 0 {
		http.Error(w, `{"error":"Задача не найдена"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(map[string]string{})
}

func GetTasksHandler(w http.ResponseWriter, r *http.Request, database *sql.DB) {
	rows, err := database.Query(`SELECT * FROM scheduler ORDER BY date ASC LIMIT 50`)
	if err != nil {
		http.Error(w, `{"error":"Failed to query tasks"}`, http.StatusInternalServerError)
		return
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Println(err)
		}
	}()

	tasks := make([]Task, 0)
	for rows.Next() {
		var task Task
		err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			http.Error(w, `{"error":"Failed to scan task"}`, http.StatusInternalServerError)
			return
		}
		tasks = append(tasks, task)
	}

	if err = rows.Err(); err != nil {
		http.Error(w, `{"error":"Failed to read tasks"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if err = json.NewEncoder(w).Encode(map[string][]Task{"tasks": tasks}); err != nil {
		http.Error(w, `{"error":"Failed to encode tasks"}`, http.StatusInternalServerError)
	}
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

	if t.Repeat == "d 1" || t.Repeat == "d 5" || t.Repeat == "d 3" {
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
