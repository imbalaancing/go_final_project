package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/imbalaancing/go_final_project/internal/db"
	"github.com/imbalaancing/go_final_project/internal/task"
)

func GetTaskHandler(w http.ResponseWriter, r *http.Request, database *sql.DB) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, `{"error":"Не указан идентификатор"}`, http.StatusBadRequest)
		return
	}

	t, err := db.GetTask(database, id)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, `{"error":"Задача не найдена"}`, http.StatusNotFound)
		} else {
			http.Error(w, `{"error":"Не удалось получить задачу"}`, http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if err = json.NewEncoder(w).Encode(t); err != nil {
		http.Error(w, `{"error":"Не удалось закодировать задачу"}`, http.StatusInternalServerError)
	}
}

func UpdateTaskHandler(w http.ResponseWriter, r *http.Request, database *sql.DB) {
	var t task.Task
	err := json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		http.Error(w, `{"error":"Ошибка десериализации JSON"}`, http.StatusBadRequest)
		return
	}

	if err = task.ValidateTask(&t); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	err = db.UpdateTask(database, &t)
	if err != nil {
		http.Error(w, `{"error":"Ошибка обновления задачи"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(map[string]string{})
}

func GetTasksHandler(w http.ResponseWriter, r *http.Request, database *sql.DB) {
	tasks, err := db.GetTasks(database)
	if err != nil {
		http.Error(w, `{"error":"Не удалось запросить задачи"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if err = json.NewEncoder(w).Encode(map[string][]task.Task{"tasks": tasks}); err != nil {
		http.Error(w, `{"error":"Не удалось закодировать задачи"}`, http.StatusInternalServerError)
	}
}

func AddTaskHandler(w http.ResponseWriter, r *http.Request, database *sql.DB) {
	var t task.Task
	err := json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		http.Error(w, `{"error":"Недопустимый текст запроса"}`, http.StatusBadRequest)
		return
	}

	if err = task.ValidateTask(&t); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	err = db.InsertTask(database, &t)
	if err != nil {
		http.Error(w, `{"error":"Не удалось добавить задачу"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(map[string]string{"id": t.ID})
}

func DeleteTaskHandler(w http.ResponseWriter, r *http.Request, database *sql.DB) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, `{"error":"Не указан идентификатор"}`, http.StatusBadRequest)
		return
	}

	err := db.DeleteTask(database, id)
	if err != nil {
		http.Error(w, `{"error":"Ошибка удаления задачи"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(map[string]string{})
}

func MarkTaskDoneHandler(w http.ResponseWriter, r *http.Request, database *sql.DB) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, `{"error":"Не указан идентификатор"}`, http.StatusBadRequest)
		return
	}

	err := db.MarkTaskDone(database, id)
	if err != nil {
		http.Error(w, `{"error":"Ошибка выполнения задачи"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(map[string]string{})
}

func NextDateHandler(w http.ResponseWriter, r *http.Request) {
	nowStr := r.FormValue("now")
	dateStr := r.FormValue("date")
	repeat := r.FormValue("repeat")

	if nowStr == "" || dateStr == "" || repeat == "" {
		http.Error(w, "Отсутствуют необходимые параметры запроса", http.StatusBadRequest)
		return
	}

	now, err := time.Parse(task.DATE_FORMAT, nowStr)
	if err != nil {
		http.Error(w, "Недопустимый формат даты сейчас", http.StatusBadRequest)
		return
	}

	nextDate, err := task.NextDate(now, dateStr, repeat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Write([]byte(nextDate))
}
