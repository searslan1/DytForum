package handlers

import (
	"html/template"
	"net/http"

	"DytForum/database"
	"DytForum/models"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	categories, err := fetchCategories()
	if err != nil {
		http.Error(w, "Failed to fetch categories", http.StatusInternalServerError)
		return
	}

	threads, err := fetchThreads()
	if err != nil {
		http.Error(w, "Failed to fetch threads", http.StatusInternalServerError)
		return
	}

	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	data := struct {
		Categories []models.Category
		Threads    []models.Thread
	}{
		Categories: categories,
		Threads:    threads,
	}
	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
	}
}

func fetchThreads() ([]models.Thread, error) {
	rows, err := database.DB.Query("SELECT id, category, title, content FROM threads WHERE approved = 1")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var threads []models.Thread
	for rows.Next() {
		var thread models.Thread
		err := rows.Scan(&thread.ID, &thread.Category, &thread.Title, &thread.Content)
		if err != nil {
			return nil, err
		}
		threads = append(threads, thread)
	}

	return threads, nil
}
