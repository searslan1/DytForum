// handlers/auth.go
package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"html/template"
	"net/http"

	"DytForum/database"

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
