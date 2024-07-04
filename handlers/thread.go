// handlers/thread.go
package handlers

import (
	"html/template"
	"net/http"

	"DytForum/database"
)

func CreateThreadHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")
	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		http.Error(w, "You must be logged in to create a thread", http.StatusUnauthorized)
		return
	}

	if r.Method == "GET" {
		tmpl := template.Must(template.ParseFiles("templates/create_thread.html"))
		tmpl.Execute(w, nil)
	} else if r.Method == "POST" {
		category := r.FormValue("category")
		title := r.FormValue("title")
		content := r.FormValue("content")
		username := session.Values["username"].(string)

		var userID int
		err := database.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		_, err = database.DB.Exec("INSERT INTO threads (user_id, category, title, content) VALUES (?, ?, ?, ?)", userID, category, title, content)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/index", http.StatusSeeOther)
	}
}
