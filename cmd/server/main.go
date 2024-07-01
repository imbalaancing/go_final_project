package main

import (
	"log"
	"net/http"
	"os"

	"github.com/imbalaancing/go_final_project/internal/api"
	"github.com/imbalaancing/go_final_project/internal/db"
)

func main() {
	dbFileName := os.Getenv("TODO_DBFILE")
	if dbFileName == "" {
		dbFileName = "scheduler.db"
	}

	database, err := db.InitDB(dbFileName)
	if err != nil {
		log.Fatalf("Ошибка инициализации базы данных: %v", err)
	}
	defer database.Close()

	storage := db.NewTaskStorage(database)

	fs := http.FileServer(http.Dir("./web"))
	http.Handle("/web/", http.StripPrefix("/web/", fs))

	http.HandleFunc("/api/nextdate", api.NextDateHandler)

	http.HandleFunc("/api/task", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			api.AddTaskHandler(w, r, storage)
		case http.MethodGet:
			api.GetTaskHandler(w, r, storage)
		case http.MethodPut:
			api.UpdateTaskHandler(w, r, storage)
		case http.MethodDelete:
			api.DeleteTaskHandler(w, r, storage)
		default:
			http.Error(w, "Неподдерживаемый метод", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/api/task/done", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			api.MarkTaskDoneHandler(w, r, storage)
		} else {
			http.Error(w, "Неподдерживаемый метод", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/api/tasks", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			api.GetTasksHandler(w, r, storage)
		} else {
			http.Error(w, "Неподдерживаемый метод", http.StatusMethodNotAllowed)
		}
	})

	port := os.Getenv("TODO_PORT")
	if port == "" {
		port = "7540"
	}

	log.Printf("Запуск сервера на порту %s...\n", port)
	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal(err)
	}
}
