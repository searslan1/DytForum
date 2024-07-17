// handlers/comments.go
package handlers

import (
	"net/http"
	"strconv"

	"DytForum/database"
	"DytForum/session"
)

func CreateCommentHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := session.Store.Get(r, "session-name")
	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		http.Error(w, "You must be logged in to create a comment", http.StatusUnauthorized)
		return
	}

	if r.Method == "POST" {
		// Parse form data
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Failed to parse form data", http.StatusInternalServerError)
			return
		}

		// Extract comment details from form data
		threadID, err := strconv.Atoi(r.Form.Get("thread_id"))
		if err != nil {
			http.Error(w, "Invalid thread ID", http.StatusBadRequest)
			return
		}
		comment := r.Form.Get("comment")
		username := session.Values["username"].(string)

		var userID int
		err = database.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
		if err != nil {
			http.Error(w, "Failed to retrieve user ID", http.StatusInternalServerError)
			return
		}

		// Insert comment into database
		err = CreateComment(userID, threadID, comment, username)
		if err != nil {
			http.Error(w, "Failed to create comment", http.StatusInternalServerError)
			return
		}

		// Redirect to thread view or another appropriate page
		http.Redirect(w, r, "/thread?id="+strconv.Itoa(threadID), http.StatusSeeOther)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func CreateComment(userID, threadID int, content string, username string) error {
	_, err := database.DB.Exec("INSERT INTO comments (user_id, thread_id, content, username) VALUES (?, ?, ?, ?)", userID, threadID, content, username)
	if err != nil {
		return err
	}
	return nil
}
