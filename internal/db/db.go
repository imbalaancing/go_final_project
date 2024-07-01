package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/imbalaancing/go_final_project/internal/date"
	"github.com/imbalaancing/go_final_project/internal/task"
	_ "github.com/mattn/go-sqlite3"
)

const dbFileName = "scheduler.db"
const TaskLimit = 50

const createTableQuery = `
CREATE TABLE IF NOT EXISTS scheduler (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	date TEXT NOT NULL,
	title TEXT NOT NULL,
	comment TEXT,
	repeat TEXT
);
CREATE INDEX IF NOT EXISTS idx_date ON scheduler(date);
`

func InitDB() (*sql.DB, error) {
	if _, err := os.Stat(dbFileName); os.IsNotExist(err) {
		file, err := os.Create(dbFileName)
		if err != nil {
			return nil, err
		}
		file.Close()
		log.Println("Создан файл базы данных.")
	}

	db, err := sql.Open("sqlite3", dbFileName)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(createTableQuery)
	if err != nil {
		return nil, err
	}
	log.Println("Таблица scheduler готова.")

	return db, nil
}

func InsertTask(db *sql.DB, t *task.Task) error {
	query := `INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`
	res, err := db.Exec(query, t.Date, t.Title, t.Comment, t.Repeat)
	if err != nil {
		return err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}

	t.ID = fmt.Sprintf("%d", id)
	return nil
}

func GetTasks(db *sql.DB) ([]task.Task, error) {
	query := `SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date ASC LIMIT ?`
	rows, err := db.Query(query, TaskLimit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := []task.Task{}
	for rows.Next() {
		var t task.Task
		if err := rows.Scan(&t.ID, &t.Date, &t.Title, &t.Comment, &t.Repeat); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

func GetTask(db *sql.DB, id string) (*task.Task, error) {
	query := `SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?`
	row := db.QueryRow(query, id)

	var t task.Task
	err := row.Scan(&t.ID, &t.Date, &t.Title, &t.Comment, &t.Repeat)
	if err != nil {
		return nil, err
	}

	return &t, nil
}

func UpdateTask(db *sql.DB, t *task.Task) error {
	query := `UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?`
	_, err := db.Exec(query, t.Date, t.Title, t.Comment, t.Repeat, t.ID)
	return err
}

func MarkTaskDone(db *sql.DB, id string) error {
	t, err := GetTask(db, id)
	if err != nil {
		return err
	}

	if t.Repeat == "" {
		return DeleteTask(db, id)
	}

	nextDate, err := date.NextDate(time.Now(), t.Date, t.Repeat)
	if err != nil {
		return err
	}

	t.Date = nextDate
	return UpdateTask(db, t)
}

func DeleteTask(db *sql.DB, id string) error {
	query := `DELETE FROM scheduler WHERE id = ?`
	_, err := db.Exec(query, id)
	return err
}
