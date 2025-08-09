package storage

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

func InitNotesDB() (*sql.DB, error) {
	var err error
	db, err := sql.Open("sqlite3", "./notes.db")
	if err != nil {
		return nil, err
	}

	// Проверяем подключение
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	// Создаем таблицу если не существует
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS notes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			content TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		    user_id INTEGER
		)`)
	if err != nil {
		return nil, err
	}

	log.Println("Notes database initialized successfully")
	return db, nil
}

func InitUsersDB() (*sql.DB, error) {
	var err error
	db, err := sql.Open("sqlite3", "./users.db")
	if err != nil {
		return nil, err
	}

	// Проверяем подключение
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	// Создаем таблицу если не существует
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL
	);`)
	if err != nil {
		return nil, err
	}

	log.Println("Users database initialized successfully")
	return db, nil
}

func GetNotesDB() *sql.DB {
	db, err := sql.Open("sqlite3", "./notes.db")
	if err != nil {
		db, _ = InitNotesDB()
	}
	return db
}

func GetUsersDB() *sql.DB {
	db, err := sql.Open("sqlite3", "./users.db")
	if err != nil {
		db, _ = InitUsersDB()
	}
	return db
}
