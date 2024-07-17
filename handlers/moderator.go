package handlers

import (
	"html/template"
	"log"
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

	// Fetch pending threads and reports
	pendingThreads, err := fetchPendingThreads()
	if err != nil {
		http.Error(w, "Failed to fetch pending threads", http.StatusInternalServerError)
		return
	}

	pendingReports, err := fetchPendingReports()
	if err != nil {
		http.Error(w, "Failed to fetch pending reports", http.StatusInternalServerError)
		return
	}

	tmpl := template.Must(template.ParseFiles("templates/moderator_panel.html"))
	data := struct {
		PendingThreads []models.Thread
		PendingReports []models.Report
	}{
		PendingThreads: pendingThreads,
		PendingReports: pendingReports,
	}
	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Failed to render moderator panel", http.StatusInternalServerError)
		log.Printf("Failed to render moderator panel: %v", err)
		return
	}
}

func fetchPendingThreads() ([]models.Thread, error) {
	rows, err := database.DB.Query("SELECT id, category, title, content, likes, dislikes, user_id, approved FROM threads WHERE approved = 0")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pendingThreads []models.Thread
	for rows.Next() {
		var thread models.Thread
		err := rows.Scan(&thread.ID, &thread.Category, &thread.Title, &thread.Content, &thread.Likes, &thread.Dislikes, &thread.UserID, &thread.Approved)
		if err != nil {
			return nil, err
		}
		pendingThreads = append(pendingThreads, thread)
	}

	return pendingThreads, nil
}

func fetchPendingReports() ([]models.Report, error) {
	rows, err := database.DB.Query(`
		SELECT 
			reports.id, reports.thread_id, reports.user_id, reports.reason, 
			users.username, threads.title, threads.content 
		FROM reports 
		INNER JOIN users ON reports.user_id = users.id 
		INNER JOIN threads ON reports.thread_id = threads.id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pendingReports []models.Report
	for rows.Next() {
		var report models.Report
		err := rows.Scan(&report.ID, &report.ThreadID, &report.UserID, &report.Reason, &report.Username, &report.Title, &report.Content)
		if err != nil {
			return nil, err
		}
		pendingReports = append(pendingReports, report)
	}

	return pendingReports, nil
}

func DeleteThreadHandler(w http.ResponseWriter, r *http.Request) {
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
