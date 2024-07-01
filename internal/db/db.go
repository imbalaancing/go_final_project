package db

import (
	"database/sql"
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

type Storage struct {
	db *sql.DB
}

func NewTaskStorage(db *sql.DB) *Storage {
	return &Storage{db: db}
}

func InitDB(dbFileName string) (*sql.DB, error) {
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

func (s *Storage) InsertTask(t task.Task) (int64, error) {
	res, err := s.db.Exec(`INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`,
		t.Date, t.Title, t.Comment, t.Repeat)
	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}

func (s *Storage) GetTasks(TaskLimit int) ([]task.Task, error) {
	rows, err := s.db.Query(`SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date ASC LIMIT ?`, TaskLimit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []task.Task
	for rows.Next() {
		var t task.Task
		err := rows.Scan(&t.ID, &t.Date, &t.Title, &t.Comment, &t.Repeat)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (s *Storage) GetTask(id string) (task.Task, error) {
	var t task.Task
	row := s.db.QueryRow(`SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?`, id)
	err := row.Scan(&t.ID, &t.Date, &t.Title, &t.Comment, &t.Repeat)
	return t, err
}

func (s *Storage) UpdateTask(t task.Task) error {
	_, err := s.db.Exec(`UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?`,
		t.Date, t.Title, t.Comment, t.Repeat, t.ID)
	return err
}

func (s *Storage) MarkTaskDone(id string) error {
	var t task.Task
	row := s.db.QueryRow(`SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?`, id)
	err := row.Scan(&t.ID, &t.Date, &t.Title, &t.Comment, &t.Repeat)
	if err != nil {
		return err
	}

	if t.Repeat == "" {
		_, err := s.db.Exec(`DELETE FROM scheduler WHERE id = ?`, id)
		return err
	}

	newDate, err := date.NextDate(time.Now(), t.Date, t.Repeat)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(`UPDATE scheduler SET date = ? WHERE id = ?`, newDate, id)
	return err
}

func (s *Storage) DeleteTask(id string) error {
	_, err := s.db.Exec(`DELETE FROM scheduler WHERE id = ?`, id)
	return err
}
