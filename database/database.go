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
		facebook_id TEXT,
		google_id TEXT,
		github_id TEXT,
		role VARCHAR(20) DEFAULT 'user'
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
		approved INTEGER DEFAULT 0,
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
		comment_id INTEGER,
		user_id INTEGER,
		like_status INTEGER,
		FOREIGN KEY(user_id) REFERENCES users(id),
		FOREIGN KEY(thread_id) REFERENCES threads(id)
	);
	`
	rolesTable := `
	CREATE TABLE IF NOT EXISTS roles (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		role_name VARCHAR(20) NOT NULL UNIQUE
	);
	`
	moderatorRequestsTable := `
	CREATE TABLE IF NOT EXISTS moderator_requests (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		reason TEXT NOT NULL,
		status TEXT NOT NULL,
		FOREIGN KEY(user_id) REFERENCES users(id)
	);
	`
	categoriesTable := `
	CREATE TABLE IF NOT EXISTS categories (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE
	);
	`
	reportsTable := `
	CREATE TABLE IF NOT EXISTS reports (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		thread_id INTEGER NOT NULL,
		user_id INTEGER NOT NULL,
		reason TEXT NOT NULL,
		FOREIGN KEY(thread_id) REFERENCES threads(id),
		FOREIGN KEY(user_id) REFERENCES users(id)
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
	_, err = DB.Exec(rolesTable)
	if err != nil {
		return fmt.Errorf("error creating roles table: %v", err)
	}
	_, err = DB.Exec(moderatorRequestsTable)
	if err != nil {
		return fmt.Errorf("error creating moderator requests table: %v", err)
	}
	_, err = DB.Exec(categoriesTable)
	if err != nil {
		return fmt.Errorf("error creating categories table: %v", err)
	}
	_, err = DB.Exec(reportsTable)
	if err != nil {
		return fmt.Errorf("error creating reports table: %v", err)
	}
	_, err = DB.Exec(`INSERT INTO roles (role_name) VALUES ('user'), ('moderator'), ('admin') ON CONFLICT(role_name) DO NOTHING;`)
	if err != nil {
		return fmt.Errorf("error inserting roles: %v", err)
	}
	return nil
}

func GetUserByUsername(username string) (models.User, error) {
	var user models.User
	var googleID sql.NullString
	var githubID sql.NullInt64
	var facebookID sql.NullString

	err := DB.QueryRow("SELECT id, email, username, password, google_id, github_id, facebook_id, role FROM users WHERE username = ?", username).Scan(&user.ID, &user.Email, &user.Username, &user.Password, &googleID, &githubID, &facebookID, &user.Role)
	if err != nil {
		return user, err
	}

	if googleID.Valid {
		user.GoogleID = googleID.String
	} else {
		user.GoogleID = ""
	}

	if githubID.Valid {
		user.GitHubID = int(githubID.Int64)
	} else {
		user.GitHubID = 0
	}

	if facebookID.Valid {
		user.FacebookID = facebookID.String
	} else {
		user.FacebookID = ""
	}

	return user, nil
}

func GetThreadsByUserID(userID int) ([]models.Thread, error) {
	rows, err := DB.Query("SELECT id, title, content, category, likes, dislikes FROM threads WHERE user_id = ?", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var threads []models.Thread
	for rows.Next() {
		var thread models.Thread
		err := rows.Scan(&thread.ID, &thread.Title, &thread.Content, &thread.Category, &thread.Likes, &thread.Dislikes)
		if err != nil {
			return nil, err
		}
		threads = append(threads, thread)
	}
	return threads, nil
}

func GetCommentsByUserID(userID int) ([]models.Comment, error) {
	rows, err := DB.Query("SELECT id, thread_id, content FROM comments WHERE user_id = ?", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []models.Comment
	for rows.Next() {
		var comment models.Comment
		err := rows.Scan(&comment.ID, &comment.ThreadID, &comment.Content)
		if err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}
	return comments, nil
}

func UpdateUserRole(userID int, role string) error {
	_, err := DB.Exec("UPDATE users SET role = ? WHERE id = ?", role, userID)
	return err
}
