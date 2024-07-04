package handlers

import (
	"html/template"
	"net/http"

	"DytForum/database"
	"DytForum/models"
)

func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")
	username, ok := session.Values["username"].(string)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	var user models.User
	err := database.DB.QueryRow("SELECT id, username FROM users WHERE username = ?", username).Scan(&user.ID, &user.Username)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	rows, err := database.DB.Query("SELECT id, title, content FROM threads WHERE user_id = ?", user.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var threads []models.Thread
	for rows.Next() {
		var thread models.Thread
		err := rows.Scan(&thread.ID, &thread.Title, &thread.Content)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		threads = append(threads, thread)
	}

	rows, err = database.DB.Query(`
		SELECT c.id, c.content, t.title 
		FROM comments c 
		JOIN threads t ON c.thread_id = t.id 
		WHERE c.user_id = ?`, user.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var comments []models.Comment
	for rows.Next() {
		var comment models.Comment
		err := rows.Scan(&comment.ID, &comment.Content, &comment.ThreadTitle)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		comments = append(comments, comment)
	}

	data := struct {
		Username string
		Threads  []models.Thread
		Comments []models.Comment
	}{
		Username: user.Username,
		Threads:  threads,
		Comments: comments,
	}

	tmpl := template.Must(template.ParseFiles("templates/profile.html"))
	tmpl.Execute(w, data)
}
