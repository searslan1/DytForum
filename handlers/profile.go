package handlers

import (
	"html/template"
	"log"
	"net/http"

	"DytForum/database"
	"DytForum/models"
	"DytForum/session"
)

func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	session, err := session.Store.Get(r, "session-name")
	if err != nil {
		log.Printf("Failed to get session: %v", err)
		http.Error(w, "Failed to get session", http.StatusInternalServerError)
		return
	}

	auth, ok := session.Values["authenticated"].(bool)
	if !ok || !auth {
		http.Error(w, "You must be logged in to access this page", http.StatusUnauthorized)
		return
	}

	username := session.Values["username"].(string)

	// Kullanıcının veritabanındaki bilgilerini al
	user, err := database.GetUserByUsername(username)
	if err != nil {
		log.Printf("Failed to retrieve user: %v", err)
		http.Error(w, "Server error, unable to retrieve your profile", http.StatusInternalServerError)
		return
	}

	// Kullanıcının thread'lerini ve yorumlarını al
	threads, err := database.GetThreadsByUserID(user.ID)
	if err != nil {
		log.Printf("Failed to retrieve threads: %v", err)
		http.Error(w, "Server error, unable to retrieve your threads", http.StatusInternalServerError)
		return
	}

	comments, err := database.GetCommentsByUserID(user.ID)
	if err != nil {
		log.Printf("Failed to retrieve comments: %v", err)
		http.Error(w, "Server error, unable to retrieve your comments", http.StatusInternalServerError)
		return
	}

	data := struct {
		Username string
		Email    string
		Role     string
		Threads  []models.Thread
		Comments []models.Comment
	}{
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
		Threads:  threads,
		Comments: comments,
	}

	tmpl, err := template.ParseFiles("templates/profile.html")
	if err != nil {
		log.Printf("Failed to parse template: %v", err)
		http.Error(w, "Failed to parse template", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		log.Printf("Failed to execute template: %v", err)
		http.Error(w, "Failed to execute template", http.StatusInternalServerError)
		return
	}
}
