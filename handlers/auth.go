package handlers

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"html/template"
	"log"
	"net/http"

	"DytForum/database"
	"DytForum/models"

	"github.com/gorilla/sessions"
)

//var store = sessions.NewCookieStore([]byte("something-very-secret"))

var store *sessions.CookieStore

func init() {
	store = sessions.NewCookieStore(
		[]byte("something-very-secret"),
	)

	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,
	}
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		tmpl := template.Must(template.ParseFiles("templates/register.html"))
		tmpl.Execute(w, nil)
	} else if r.Method == "POST" {
		email := r.FormValue("email")
		username := r.FormValue("username")
		password := r.FormValue("password")

		// E-posta adresinin zaten var olup olmadığını kontrol edin
		var existingEmail string
		err := database.DB.QueryRow("SELECT email FROM users WHERE email = ?", email).Scan(&existingEmail)
		if err == nil {
			http.Error(w, "Email address already in use", http.StatusBadRequest)
			return
		}

		hash := sha256.New()
		hash.Write([]byte(password))
		hashedPassword := hex.EncodeToString(hash.Sum(nil))

		_, err = database.DB.Exec("INSERT INTO users (email, username, password) VALUES (?, ?, ?)", email, username, hashedPassword)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
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
		var storedPassword string
		var userID int
		err := database.DB.QueryRow("SELECT id, password FROM users WHERE username = ?", username).Scan(&userID, &storedPassword)
		if err != nil {
			http.Error(w, "Invalid username or password", http.StatusUnauthorized)
			return
		}

		hash := sha256.New()
		hash.Write([]byte(password))
		hashedPassword := hex.EncodeToString(hash.Sum(nil))

		if hashedPassword != storedPassword {
			http.Error(w, "Invalid username or password", http.StatusUnauthorized)
			return
		}

		session, _ := store.Get(r, "session-name")
		session.Values["authenticated"] = true
		session.Values["username"] = username
		session.Values["userID"] = userID
		log.Printf("Login successful for user: %s with userID: %d", username, userID) // Added logging
		err = session.Save(r, w)
		//debug
		if err != nil {
			log.Printf("Failed to save session: %v", err) // Added logging
			http.Error(w, "Failed to save session", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/index", http.StatusSeeOther)
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
	session, _ := store.Get(r, "session-name")

	// Revoke user's authentication by clearing session values
	session.Values["authenticated"] = false
	session.Values["username"] = ""
	session.Save(r, w)

	// Redirect the user to the login page or any other appropriate page
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// DebugSessionHandler to print session values for debugging
func DebugSessionHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")
	log.Printf("Session values: %v", session.Values)
	w.Write([]byte("Check server logs for session values"))
}
