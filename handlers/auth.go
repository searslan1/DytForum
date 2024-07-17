package handlers

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"

	"DytForum/database"
	"DytForum/models"
	"DytForum/session"

	"golang.org/x/crypto/bcrypt"
)

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		tmpl := template.Must(template.ParseFiles("templates/register.html"))
		tmpl.Execute(w, nil)
	} else if r.Method == "POST" {
		username := r.FormValue("username")
		email := r.FormValue("email")
		password := r.FormValue("password")
		isModerator := r.FormValue("moderator") == "on"
		isAdmin := r.FormValue("admin") == "on" // Admin kaydı için

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Server error, unable to create your account.", http.StatusInternalServerError)
			return
		}

		role := "user"
		if isAdmin {
			role = "admin"
		}

		_, err = database.DB.Exec("INSERT INTO users (username, email, password, role) VALUES (?, ?, ?, ?)", username, email, hashedPassword, role)
		if err != nil {
			log.Printf("RegisterHandler: Failed to insert user: %v", err)
			http.Error(w, "Server error, unable to create your account.", http.StatusInternalServerError)
			return
		}

		if isModerator && !isAdmin { // Sadece kullanıcı ve moderatör kaydı için
			var userID int
			err = database.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
			if err != nil {
				http.Error(w, "Server error, unable to process your request.", http.StatusInternalServerError)
				return
			}

			_, err = database.DB.Exec("INSERT INTO moderator_requests (user_id, reason, status) VALUES (?, ?, 'pending')", userID, "Applied during registration")
			if err != nil {
				http.Error(w, "Server error, unable to submit your moderator request.", http.StatusInternalServerError)
				return
			}
		}

		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		tmpl := template.Must(template.ParseFiles("templates/login.html"))
		tmpl.Execute(w, nil)
	} else if r.Method == "POST" {
		username := r.FormValue("username")
		password := r.FormValue("password")
		var storedPassword, role string
		var userID int
		err := database.DB.QueryRow("SELECT id, password, role FROM users WHERE username = ?", username).Scan(&userID, &storedPassword, &role)
		if err != nil {
			http.Error(w, "Invalid username or password", http.StatusUnauthorized)
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(password)); err != nil {
			http.Error(w, "Invalid username or password", http.StatusUnauthorized)
			return
		}

		session, _ := session.Store.Get(r, "session-name")
		session.Values["authenticated"] = true
		session.Values["username"] = username
		session.Values["userID"] = userID
		session.Values["role"] = role
		log.Printf("Login successful for user: %s with userID: %d and role: %s", username, userID, role)
		err = session.Save(r, w)
		if err != nil {
			log.Printf("Failed to save session: %v", err)
			http.Error(w, "Failed to save session", http.StatusInternalServerError)
			return
		}

		if role == "admin" {
			http.Redirect(w, r, "/admin/panel", http.StatusSeeOther)
		} else {
			http.Redirect(w, r, "/index", http.StatusSeeOther)
		}
	}
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")

	var rows *sql.Rows
	var err error
	if category != "" {
		rows, err = database.DB.Query("SELECT id, title, content, category FROM threads WHERE category = ?", category)
	} else {
		rows, err = database.DB.Query("SELECT id, title, content, category FROM threads")
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var threads []models.Thread
	for rows.Next() {
		var thread models.Thread
		err := rows.Scan(&thread.ID, &thread.Title, &thread.Content, &thread.Category)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		threads = append(threads, thread)
	}

	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	tmpl.Execute(w, struct {
		Threads []models.Thread
	}{
		Threads: threads,
	})
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := session.Store.Get(r, "session-name")

	// Revoke user's authentication by clearing session values
	session.Values["authenticated"] = false
	session.Values["username"] = ""
	session.Save(r, w)

	// Redirect the user to the login page or any other appropriate page
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// DebugSessionHandler to print session values for debugging
func DebugSessionHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := session.Store.Get(r, "session-name")
	log.Printf("Session values: %v", session.Values)
	w.Write([]byte("Check server logs for session values"))
}
