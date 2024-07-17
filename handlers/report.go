package handlers

import (
	"html/template"
	"log"
	"net/http"
	"strconv"

	"DytForum/database"
	"DytForum/session"
)

func ReportThreadHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := session.Store.Get(r, "session-name")
	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		http.Error(w, "You must be logged in to report a thread", http.StatusUnauthorized)
		return
	}

	userID := session.Values["userID"].(int)
	threadID, err := strconv.Atoi(r.FormValue("thread_id"))
	if err != nil {
		http.Error(w, "Invalid thread ID", http.StatusBadRequest)
		return
	}
	reason := r.FormValue("reason")

	_, err = database.DB.Exec("INSERT INTO reports (thread_id, user_id, reason) VALUES (?, ?, ?)", threadID, userID, reason)
	if err != nil {
		http.Error(w, "Failed to report thread", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/thread?id="+strconv.Itoa(threadID), http.StatusSeeOther)
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

	rows, err := database.DB.Query("SELECT reports.id, reports.thread_id, reports.user_id, reports.reason, users.username FROM reports INNER JOIN users ON reports.user_id = users.id")
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
	}
	for rows.Next() {
		var report struct {
			ID       int
			ThreadID int
			UserID   int
			Reason   string
			Username string
		}
		err := rows.Scan(&report.ID, &report.ThreadID, &report.UserID, &report.Reason, &report.Username)
		if err != nil {
			log.Printf("ListReportsHandler: Failed to scan report: %v", err)
			http.Error(w, "Failed to scan report", http.StatusInternalServerError)
			return
		}
		reports = append(reports, report)
	}

	tmpl := template.Must(template.ParseFiles("templates/reports.html"))
	err = tmpl.Execute(w, struct {
		Reports []struct {
			ID       int
			ThreadID int
			UserID   int
			Reason   string
			Username string
		}
	}{Reports: reports})
	if err != nil {
		log.Printf("ListReportsHandler: Failed to execute template: %v", err)
		http.Error(w, "Failed to execute template", http.StatusInternalServerError)
	}
}
