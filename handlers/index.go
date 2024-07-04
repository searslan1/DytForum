// handlers/index.go
package handlers

import (
	"database/sql"
	"html/template"
	"net/http"

	"DytForum/database"
	"DytForum/models"
)

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")

	var rows *sql.Rows
	var err error
	if category != "" {
		rows, err = database.DB.Query("SELECT id, title, content, category FROM threads WHERE category = ?", category)
	} else {
		rows, err = database.DB.Query("SELECT id, title, content, category FROM threads")
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var threads []models.Thread
	for rows.Next() {
		var thread models.Thread
		err := rows.Scan(&thread.ID, &thread.Title, &thread.Content, &thread.Category)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		threads = append(threads, thread)
	}

	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	tmpl.Execute(w, struct {
		Threads []models.Thread
	}{
		Threads: threads,
	})
}

func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")
	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	username, ok := session.Values["username"].(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := database.GetUserByUsername(username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	threads, err := database.GetThreadsByUserID(user.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl := template.Must(template.ParseFiles("templates/profile.html"))
	tmpl.Execute(w, struct {
		User    models.User
		Threads []models.Thread
	}{
		User:    user,
		Threads: threads,
	})
}
