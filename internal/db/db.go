package db

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

const dbFileName = "scheduler.db"

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
