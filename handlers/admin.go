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

func PromoteUserHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := session.Store.Get(r, "session-name")
	role := session.Values["role"]
	if role != "admin" {
		http.Error(w, "You must be an admin to perform this action", http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	newRole := r.URL.Query().Get("role")
	if newRole != "moderator" && newRole != "user" {
		http.Error(w, "Invalid role", http.StatusBadRequest)
		return
	}

	err = database.UpdateUserRole(userID, newRole)
	if err != nil {
		http.Error(w, "Failed to update user role", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func CreateCategoryHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := session.Store.Get(r, "session-name")
	role := session.Values["role"]
	if role != "admin" {
		http.Error(w, "You must be an admin to perform this action", http.StatusForbidden)
		return
	}

	if r.Method == "POST" {
		name := r.FormValue("name")
		_, err := database.DB.Exec("INSERT INTO categories (name) VALUES (?)", name)
		if err != nil {
			http.Error(w, "Failed to create category", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/admin/categories", http.StatusSeeOther)
	} else {
		tmpl := template.Must(template.ParseFiles("templates/create_category.html"))
		tmpl.Execute(w, nil)
	}
}

func DeleteCategoryHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := session.Store.Get(r, "session-name")
	role := session.Values["role"]
	if role != "admin" {
		http.Error(w, "You must be an admin to perform this action", http.StatusForbidden)
		return
	}

	categoryID := r.URL.Query().Get("id")
	_, err := database.DB.Exec("DELETE FROM categories WHERE id = ?", categoryID)
	if err != nil {
		http.Error(w, "Failed to delete category", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/categories", http.StatusSeeOther)
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

func ApproveModeratorHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := session.Store.Get(r, "session-name")
	role := session.Values["role"]
	if role != "admin" {
		http.Error(w, "You must be an admin to perform this action", http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	_, err = database.DB.Exec("UPDATE users SET role = 'moderator' WHERE id = ?", userID)
	if err != nil {
		http.Error(w, "Failed to update user role", http.StatusInternalServerError)
		return
	}

	_, err = database.DB.Exec("UPDATE moderator_requests SET status = 'approved' WHERE user_id = ?", userID)
	if err != nil {
		http.Error(w, "Failed to update request status", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/moderator-requests", http.StatusSeeOther)
}

func RejectModeratorHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := session.Store.Get(r, "session-name")
	role := session.Values["role"]
	if role != "admin" {
		http.Error(w, "You must be an admin to perform this action", http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	_, err = database.DB.Exec("UPDATE moderator_requests SET status = 'rejected' WHERE user_id = ?", userID)
	if err != nil {
		http.Error(w, "Failed to update request status", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/moderator-requests", http.StatusSeeOther)
}
