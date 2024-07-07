package database

import (
	"database/sql"
	"fmt"

	"DytForum/models"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func InitDB(dataSourceName string) error {
	var err error
	DB, err = sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return err
	}
	if err = DB.Ping(); err != nil {
		return err
	}
	return createTables()
}

func createTables() error {
	usersTable := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL,
		email TEXT NOT NULL UNIQUE,
		facebook_id TEXT
	);
	`
	threadsTable := `
	CREATE TABLE IF NOT EXISTS threads (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		category TEXT NOT NULL,
		title TEXT NOT NULL,
		content TEXT NOT NULL,
		likes INTEGER DEFAULT 0,
    	dislikes INTEGER DEFAULT 0,
		user_id INTEGER NOT NULL,
		FOREIGN KEY(user_id) REFERENCES users(id)
	);
	`
	commentTable := `
	CREATE TABLE IF NOT EXISTS comments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		thread_id INTEGER NOT NULL,
		user_id INTEGER NOT NULL,
		content TEXT,
		likes INTEGER DEFAULT 0,
    	dislikes INTEGER DEFAULT 0,
		username TEXT,
		FOREIGN KEY(user_id) REFERENCES users(id),
		FOREIGN KEY(thread_id) REFERENCES threads(id)
	);
	`
	likesTable := `
	CREATE TABLE IF NOT EXISTS likes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		thread_id INTEGER,
		user_id INTEGER,
		like_status INTEGER,
		FOREIGN KEY(user_id) REFERENCES users(id),
		FOREIGN KEY(thread_id) REFERENCES threads(id)
	);
	`

	_, err := DB.Exec(usersTable)
	if err != nil {
		return fmt.Errorf("error creating users table: %v", err)
	}

	_, err = DB.Exec(threadsTable)
	if err != nil {
		return fmt.Errorf("error creating threads table: %v", err)
	}
	_, err = DB.Exec(commentTable)
	if err != nil {
		return fmt.Errorf("error creating comments table: %v", err)
	}
	_, err = DB.Exec(likesTable)
	if err != nil {
		return fmt.Errorf("error creating likes table: %v", err)
	}
	return nil
}

func GetUserByUsername(username string) (models.User, error) {
	var user models.User
	err := DB.QueryRow("SELECT id, email, username, password FROM users WHERE username = ?", username).Scan(&user.ID, &user.Email, &user.Username, &user.Password)
	return user, err
}

func GetThreadsByUserID(userID int) ([]models.Thread, error) {
	rows, err := DB.Query("SELECT id, title, content, category FROM threads WHERE user_id = ?", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var threads []models.Thread
	for rows.Next() {
		var thread models.Thread
		err := rows.Scan(&thread.ID, &thread.Title, &thread.Content, &thread.Category)
		if err != nil {
			return nil, err
		}
		threads = append(threads, thread)
	}
	return threads, nil
}
