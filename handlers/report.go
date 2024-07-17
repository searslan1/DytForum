package handlers

import (
	"html/template"
	"log"
	"net/http"
	"strconv"

	"DytForum/database"
	"DytForum/models"
	"DytForum/session"
)

func ReportThreadHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := session.Store.Get(r, "session-name")
	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		http.Error(w, "You must be logged in to report a thread", http.StatusUnauthorized)
		return
	}

	userID := session.Values["userID"].(int)

	// Parse thread_id from form value
	threadID, err := strconv.Atoi(r.FormValue("thread_id"))
	if err != nil {
		http.Error(w, "Invalid thread ID", http.StatusBadRequest)
		return
	}

	// Validate reason field
	reason := r.FormValue("reason")
	if reason == "" {
		http.Error(w, "Reason cannot be empty", http.StatusBadRequest)
		return
	}

	// Insert report into database
	_, err = database.DB.Exec("INSERT INTO reports (thread_id, user_id, reason) VALUES (?, ?, ?)", threadID, userID, reason)
	if err != nil {
		http.Error(w, "Failed to report thread", http.StatusInternalServerError)
		return
	}

	// Redirect to the index page after reporting
	http.Redirect(w, r, "/index", http.StatusSeeOther)
}

func ListReportsHandler(w http.ResponseWriter, r *http.Request) {
	session, err := session.Store.Get(r, "session-name")
	if err != nil {
		log.Printf("ListReportsHandler: Failed to get session: %v", err)
		http.Error(w, "Failed to get session", http.StatusInternalServerError)
		return
	}

	role, ok := session.Values["role"].(string)
	if !ok || role != "moderator" {
		http.Error(w, "You must be a moderator to access this page", http.StatusForbidden)
		return
	}

	rows, err := database.DB.Query(`
		SELECT 
			reports.id, reports.thread_id, reports.user_id, reports.reason, 
			users.username, threads.title, threads.content 
		FROM reports 
		INNER JOIN users ON reports.user_id = users.id 
		INNER JOIN threads ON reports.thread_id = threads.id
	`)
	if err != nil {
		log.Printf("ListReportsHandler: Failed to retrieve reports: %v", err)
		http.Error(w, "Failed to retrieve reports", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var reports []struct {
		ID       int
		ThreadID int
		UserID   int
		Reason   string
		Username string
		Title    string
		Content  string
	}
	for rows.Next() {
		var report struct {
			ID       int
			ThreadID int
			UserID   int
			Reason   string
			Username string
			Title    string
			Content  string
		}
		err := rows.Scan(&report.ID, &report.ThreadID, &report.UserID, &report.Reason, &report.Username, &report.Title, &report.Content)
		if err != nil {
			log.Printf("ListReportsHandler: Failed to scan report: %v", err)
			http.Error(w, "Failed to scan report", http.StatusInternalServerError)
			return
		}
		reports = append(reports, report)
	}

	tmpl := template.Must(template.ParseFiles("templates/moderator_panel.html"))
	var pendingThreads []models.Thread

	data := struct {
		PendingThreads []models.Thread
		Reports        []struct {
			ID       int
			ThreadID int
			UserID   int
			Reason   string
			Username string
			Title    string
			Content  string
		}
	}{
		PendingThreads: pendingThreads, // Assuming you have pendingThreads defined elsewhere
		Reports:        reports,
	}
	err = tmpl.Execute(w, data)
	if err != nil {
		log.Printf("ListReportsHandler: Failed to execute template: %v", err)
		http.Error(w, "Failed to execute template", http.StatusInternalServerError)
	}
}
