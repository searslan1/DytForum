package handlers

import (
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

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Server error, unable to create your account.", http.StatusInternalServerError)
			return
		}

		_, err = database.DB.Exec("INSERT INTO users (username, email, password, role) VALUES (?, ?, ?, 'user')", username, email, hashedPassword)
		if err != nil {
			log.Printf("RegisterHandler: Failed to insert user: %v", err)
			http.Error(w, "Server error, unable to create your account.", http.StatusInternalServerError)
			return
		}

		if isModerator {
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

func AdminLoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		tmpl := template.Must(template.ParseFiles("templates/admin.html"))
		tmpl.Execute(w, nil)
	} else if r.Method == "POST" {
		username := r.FormValue("username")
		password := r.FormValue("password")

		var storedPassword string
		err := database.DB.QueryRow("SELECT password FROM users WHERE username = ? AND role = 'admin'", username).Scan(&storedPassword)
		if err != nil {
			http.Error(w, "Invalid username or password", http.StatusUnauthorized)
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(password))
		if err != nil {
			http.Error(w, "Invalid username or password", http.StatusUnauthorized)
			return
		}

		// Create a session to store admin role information
		session, _ := session.Store.Get(r, "session-name")
		session.Values["role"] = "admin"
		session.Save(r, w)

		http.Redirect(w, r, "/admin/panel", http.StatusSeeOther)
	}
}

func AdminLogoutHandler(w http.ResponseWriter, r *http.Request) {
	// Çıkış işlemleri yapılabilir (session temizleme vb.)
	// Örneğin:
	session, _ := session.Store.Get(r, "session-name")
	session.Options.MaxAge = -1 // Session'ı sonlandır
	err := session.Save(r, w)
	if err != nil {
		http.Error(w, "Failed to logout", http.StatusInternalServerError)
		return
	}

	// Çıkış sonrası yönlendirme
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}
