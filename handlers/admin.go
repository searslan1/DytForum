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

func AdminPanelHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := session.Store.Get(r, "session-name")
	role := session.Values["role"]
	if role != "admin" {
		http.Error(w, "You must be an admin to access this page", http.StatusForbidden)
		return
	}

	moderatorRequests, err := fetchModeratorRequests()
	if err != nil {
		http.Error(w, "Failed to fetch moderator requests", http.StatusInternalServerError)
		return
	}

	users, err := fetchUsers()
	if err != nil {
		http.Error(w, "Failed to fetch users", http.StatusInternalServerError)
		return
	}

	reports, err := fetchReports()
	if err != nil {
		http.Error(w, "Failed to fetch reports", http.StatusInternalServerError)
		return
	}

	tmpl := template.Must(template.ParseFiles("templates/admin_panel.html"))
	data := struct {
		ModeratorRequests []models.ModeratorRequest
		Users             []models.User
		Reports           []models.Report
	}{
		ModeratorRequests: moderatorRequests,
		Users:             users,
		Reports:           reports,
	}
	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Failed to render admin panel", http.StatusInternalServerError)
	}
}

func fetchModeratorRequests() ([]models.ModeratorRequest, error) {
	rows, err := database.DB.Query("SELECT users.id, users.username, moderator_requests.reason FROM users INNER JOIN moderator_requests ON users.id = moderator_requests.user_id WHERE moderator_requests.status = 'pending'")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []models.ModeratorRequest
	for rows.Next() {
		var request models.ModeratorRequest
		err := rows.Scan(&request.UserID, &request.Username, &request.Reason)
		if err != nil {
			return nil, err
		}
		requests = append(requests, request)
	}

	return requests, nil
}

func fetchUsers() ([]models.User, error) {
	rows, err := database.DB.Query("SELECT id, username, role FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.ID, &user.Username, &user.Role)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

func fetchReports() ([]models.Report, error) {
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

	var reports []models.Report
	for rows.Next() {
		var report models.Report
		err := rows.Scan(&report.ID, &report.ThreadID, &report.UserID, &report.Reason, &report.Username, &report.Title, &report.Content)
		if err != nil {
			return nil, err
		}
		reports = append(reports, report)
	}

	return reports, nil
}

func PromoteUserHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	_, err = database.DB.Exec("UPDATE users SET role = 'moderator' WHERE id = ?", userID)
	if err != nil {
		http.Error(w, "Failed to promote user", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/panel", http.StatusSeeOther)
}

func DemoteUserHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	_, err = database.DB.Exec("UPDATE users SET role = 'user' WHERE id = ?", userID)
	if err != nil {
		http.Error(w, "Failed to demote user", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/panel", http.StatusSeeOther)
}

func ApproveModeratorHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	_, err = database.DB.Exec("UPDATE users SET role = 'moderator' WHERE id = ?", userID)
	if err != nil {
		http.Error(w, "Failed to approve moderator request", http.StatusInternalServerError)
		return
	}

	_, err = database.DB.Exec("UPDATE moderator_requests SET status = 'approved' WHERE user_id = ?", userID)
	if err != nil {
		http.Error(w, "Failed to update moderator request status", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/panel", http.StatusSeeOther)
}

func RejectModeratorHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	_, err = database.DB.Exec("DELETE FROM moderator_requests WHERE user_id = ?", userID)
	if err != nil {
		http.Error(w, "Failed to reject moderator request", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/panel", http.StatusSeeOther)
}

func CreateCategoryHandler(w http.ResponseWriter, r *http.Request) {
	category := r.FormValue("category")
	if category == "" {
		http.Error(w, "Category name cannot be empty", http.StatusBadRequest)
		return
	}

	_, err := database.DB.Exec("INSERT INTO categories (name) VALUES (?)", category)
	if err != nil {
		http.Error(w, "Failed to create category", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/panel", http.StatusSeeOther)
}

func DeleteCategoryHandler(w http.ResponseWriter, r *http.Request) {
	category := r.FormValue("category")
	if category == "" {
		http.Error(w, "Category name cannot be empty", http.StatusBadRequest)
		return
	}

	_, err := database.DB.Exec("DELETE FROM categories WHERE name = ?", category)
	if err != nil {
		http.Error(w, "Failed to delete category", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/panel", http.StatusSeeOther)
}

func ListModeratorRequestsHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := session.Store.Get(r, "session-name")
	role := session.Values["role"]
	if role != "admin" {
		http.Error(w, "You must be an admin to access this page", http.StatusForbidden)
		return
	}

	rows, err := database.DB.Query("SELECT users.id, users.username, moderator_requests.reason FROM users INNER JOIN moderator_requests ON users.id = moderator_requests.user_id WHERE moderator_requests.status = 'pending'")
	if err != nil {
		http.Error(w, "Server error, unable to retrieve requests.", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var requests []models.ModeratorRequest
	for rows.Next() {
		var request models.ModeratorRequest
		err := rows.Scan(&request.UserID, &request.Username, &request.Reason)
		if err != nil {
			http.Error(w, "Server error, unable to process request.", http.StatusInternalServerError)
			return
		}
		requests = append(requests, request)
	}

	tmpl := template.Must(template.ParseFiles("templates/moderator_requests.html"))
	tmpl.Execute(w, struct{ Requests []models.ModeratorRequest }{Requests: requests})
}
