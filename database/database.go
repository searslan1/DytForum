// database/database.go
package database

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func InitDB() {
	var err error
	DB, err = sql.Open("sqlite3", "./forum.db")
	if err != nil {
		panic(err)
	}

	createTables()
}

func createTables() {
	createUsersTable := `
    CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        username TEXT NOT NULL UNIQUE,
        password TEXT NOT NULL,
        email TEXT NOT NULL UNIQUE
    );`

	createThreadsTable := `
    CREATE TABLE IF NOT EXISTS threads (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        user_id INTEGER,
        category TEXT,
        title TEXT,
        content TEXT,
        FOREIGN KEY (user_id) REFERENCES users(id)
    );`

	createCommentsTable := `
    CREATE TABLE IF NOT EXISTS comments (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        thread_id INTEGER NOT NULL,
        user_id INTEGER NOT NULL,
        content TEXT NOT NULL,
        FOREIGN KEY (thread_id) REFERENCES threads(id),
        FOREIGN KEY (user_id) REFERENCES users(id)
    );`

	createLikesTable := `
    CREATE TABLE IF NOT EXISTS likes (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        thread_id INTEGER,
        user_id INTEGER,
        like_status INTEGER,
        FOREIGN KEY (thread_id) REFERENCES threads(id),
        FOREIGN KEY (user_id) REFERENCES users(id)
    );`

	_, err := DB.Exec(createUsersTable)
	if err != nil {
		panic(err)
	}
	_, err = DB.Exec(createThreadsTable)
	if err != nil {
		panic(err)
	}
	_, err = DB.Exec(createCommentsTable)
	if err != nil {
		panic(err)
	}
	_, err = DB.Exec(createLikesTable)
	if err != nil {
		panic(err)
	}
}
