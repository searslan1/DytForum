// handlers/thread.go
package handlers

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"DytForum/database"
	"DytForum/models"
	"DytForum/session"
)

// CreateThreadHandler handles the creation of a new thread.
func CreateThreadHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := session.Store.Get(r, "session-name")
	role := session.Values["role"]
	if role != "user" && role != "moderator" {
		http.Error(w, "You must be logged in to create a thread", http.StatusUnauthorized)
		return
	}

	if r.Method == "GET" {
		categories, err := fetchCategories()
		if err != nil {
			http.Error(w, "Failed to fetch categories", http.StatusInternalServerError)
			return
		}

		tmpl := template.Must(template.ParseFiles("templates/create_thread.html"))
		data := struct {
			Categories []models.Category
		}{
			Categories: categories,
		}
		tmpl.Execute(w, data)
	} else if r.Method == "POST" {
		title := r.FormValue("title")
		content := r.FormValue("content")
		categoryID, err := strconv.Atoi(r.FormValue("category"))
		if err != nil {
			http.Error(w, "Invalid category ID", http.StatusBadRequest)
			return
		}
		userID := session.Values["userID"].(int)

		_, err = database.DB.Exec("INSERT INTO threads (category, title, content, user_id) VALUES (?, ?, ?, ?)", categoryID, title, content, userID)
		if err != nil {
			http.Error(w, "Failed to create thread", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/index", http.StatusSeeOther)
	}
}

// ViewThreadHandler handles the display of a thread with its comments.

func ViewThreadHandler(w http.ResponseWriter, r *http.Request) {
	threadIDStr := r.URL.Query().Get("id")
	if threadIDStr == "" {
		http.Error(w, "Missing thread ID", http.StatusBadRequest)
		return
	}

	threadID, err := strconv.Atoi(threadIDStr)
	if err != nil {
		http.Error(w, "Invalid thread ID", http.StatusBadRequest)
		return
	}

	// Fetch the thread details
	var thread models.Thread
	err = database.DB.QueryRow("SELECT id, title, content, category, likes, dislikes, user_id, approved FROM threads WHERE id = ?", threadID).Scan(&thread.ID, &thread.Title, &thread.Content, &thread.Category, &thread.Likes, &thread.Dislikes, &thread.UserID, &thread.Approved)
	if err != nil {
		http.Error(w, "Failed to fetch thread details", http.StatusInternalServerError)
		return
	}

	// Fetch the comments for the thread
	rows, err := database.DB.Query("SELECT id, user_id, content, username, likes, dislikes FROM comments WHERE thread_id = ?", threadID)
	if err != nil {
		http.Error(w, "Failed to fetch comments", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var comments []models.Comment
	for rows.Next() {
		var comment models.Comment
		if err := rows.Scan(&comment.ID, &comment.UserID, &comment.Content, &comment.Username, &comment.Likes, &comment.Dislikes); err != nil {
			http.Error(w, "Failed to scan comment", http.StatusInternalServerError)
			return
		}
		comment.ThreadID = threadID
		comments = append(comments, comment)
	}

	// Fetch the username of the thread creator
	var creatorUsername string
	err = database.DB.QueryRow("SELECT username FROM users WHERE id = ?", thread.UserID).Scan(&creatorUsername)
	if err != nil {
		http.Error(w, "Failed to fetch thread creator username", http.StatusInternalServerError)
		return
	}

	// Construct the file path for the picture
	picturePath := fmt.Sprintf("static/uploads/%d.jpg", threadID)

	// Render the thread page with comments
	data := struct {
		Thread       models.Thread
		Comments     []models.Comment
		Username     string
		PicturePath  string
		PictureError string
	}{
		Thread:      thread,
		Comments:    comments,
		Username:    creatorUsername,
		PicturePath: picturePath,
	}
	err = renderTemplate(w, "threads.html", data)
	if err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
	}
}

// renderTemplate renders the HTML templates.
func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) error {
	templates, err := template.ParseGlob("templates/*.html")
	if err != nil {
		log.Printf("Error parsing templates: %v", err)
		return err
	}
	err = templates.ExecuteTemplate(w, tmpl, data)
	if err != nil {
		log.Printf("Error executing template: %v", err)
	}
	return err
}
