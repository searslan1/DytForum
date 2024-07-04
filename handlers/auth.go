package handlers

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"html/template"
	"net/http"

	"DytForum/database"
	"DytForum/models"

	"github.com/gorilla/sessions"
	_ "github.com/gorilla/sessions"
)

var store = sessions.NewCookieStore([]byte("something-very-secret"))

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		tmpl := template.Must(template.ParseFiles("templates/register.html"))
		tmpl.Execute(w, nil)
	} else if r.Method == "POST" {
		email := r.FormValue("email")
		username := r.FormValue("username")
		password := r.FormValue("password")

		hash := sha256.New()
		hash.Write([]byte(password))
		hashedPassword := hex.EncodeToString(hash.Sum(nil))

		_, err := database.DB.Exec("INSERT INTO users (email, username, password) VALUES (?, ?, ?)", email, username, hashedPassword)
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
		err := database.DB.QueryRow("SELECT password FROM users WHERE username = ?", username).Scan(&storedPassword)
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
		session.Save(r, w)

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

func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")
	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	username, ok := session.Values["username"].(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := database.GetUserByUsername(username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	threads, err := database.GetThreadsByUserID(user.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl := template.Must(template.ParseFiles("templates/profile.html"))
	tmpl.Execute(w, struct {
		User    models.User
		Threads []models.Thread
	}{
		User:    user,
		Threads: threads,
	})
}
