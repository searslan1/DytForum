// handlers/view.go
package handlers

import (
	"database/sql"
	"html/template"
	"net/http"

	"DytForum/database"

	"DytForum/models"
)

func ViewThreadsHandler(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")

	var rows *sql.Rows
	var err error
	if category != "" {
		rows, err = database.DB.Query("SELECT id, title, content FROM threads WHERE category = ?", category)
	} else {
		rows, err = database.DB.Query("SELECT id, title, content FROM threads")
	}
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

	tmpl := template.Must(template.ParseFiles("templates/threads.html"))
	tmpl.Execute(w, threads)
}
