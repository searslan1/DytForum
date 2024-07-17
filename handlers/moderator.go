package handlers

import (
	"html/template"
	"net/http"
	"strconv"

	"DytForum/database"
	"DytForum/models"
	"DytForum/session"

	"github.com/gorilla/mux"
)

func ApproveThreadHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	threadID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid thread ID", http.StatusBadRequest)
		return
	}

	_, err = database.DB.Exec("UPDATE threads SET approved = 1 WHERE id = ?", threadID)
	if err != nil {
		http.Error(w, "Failed to approve thread", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/moderator/panel", http.StatusSeeOther)
}

func RejectThreadHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	threadID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid thread ID", http.StatusBadRequest)
		return
	}

	_, err = database.DB.Exec("DELETE FROM threads WHERE id = ?", threadID)
	if err != nil {
		http.Error(w, "Failed to delete thread", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/moderator/panel", http.StatusSeeOther)
}

func ApproveReportHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	reportID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid report ID", http.StatusBadRequest)
		return
	}

	// Example: Update report status in database or perform other actions
	// For example purposes, let's assume we update the status in reports table
	_, err = database.DB.Exec("UPDATE reports SET status = 'approved' WHERE id = ?", reportID)
	if err != nil {
		http.Error(w, "Failed to approve report", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/moderator/panel", http.StatusSeeOther)
}

func RejectReportHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	reportID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid report ID", http.StatusBadRequest)
		return
	}

	// Example: Delete report from database or mark it as rejected
	// For example purposes, let's assume we delete the report from reports table
	_, err = database.DB.Exec("DELETE FROM reports WHERE id = ?", reportID)
	if err != nil {
		http.Error(w, "Failed to reject report", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/moderator/panel", http.StatusSeeOther)
}

func ModeratorRequestHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := session.Store.Get(r, "session-name")
	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		http.Error(w, "You must be logged in to request to become a moderator", http.StatusUnauthorized)
		return
	}

	if r.Method == "GET" {
		tmpl := template.Must(template.ParseFiles("templates/moderator_request.html"))
		tmpl.Execute(w, nil)
	} else if r.Method == "POST" {
		reason := r.FormValue("reason")
		userID := session.Values["userID"].(int)

		_, err := database.DB.Exec("INSERT INTO moderator_requests (user_id, reason, status) VALUES (?, ?, 'pending')", userID, reason)
		if err != nil {
			http.Error(w, "Server error, unable to submit your request.", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/profile", http.StatusSeeOther)
	}
}

func ModeratorPanelHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := session.Store.Get(r, "session-name")
	role := session.Values["role"]
	if role != "moderator" {
		http.Error(w, "You must be a moderator to access this page", http.StatusForbidden)
		return
	}

	rows, err := database.DB.Query("SELECT id, title FROM threads WHERE approved = 0")
	if err != nil {
		http.Error(w, "Server error, unable to retrieve threads.", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var pendingThreads []models.Thread
	for rows.Next() {
		var thread models.Thread
		err := rows.Scan(&thread.ID, &thread.Title)
		if err != nil {
			http.Error(w, "Server error, unable to process threads.", http.StatusInternalServerError)
			return
		}
		pendingThreads = append(pendingThreads, thread)
	}

	tmpl := template.Must(template.ParseFiles("templates/moderator_panel.html"))
	tmpl.Execute(w, struct{ PendingThreads []models.Thread }{PendingThreads: pendingThreads})
}
