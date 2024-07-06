package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"DytForum/handlers"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"
)

var (
	fbOauthConfig = &oauth2.Config{
		ClientID:     os.Getenv("509500681599762"),
		ClientSecret: os.Getenv("c34fb62a5104cff9b2451935aae454fb"),
		RedirectURL:  "http://localhost:8080/auth/facebook/callback",
		Scopes:       []string{"public_profile", "email"},
		Endpoint:     facebook.Endpoint,
	}
	db    *sql.DB
	store = sessions.NewCookieStore([]byte("something-very-secret"))
)

type UserInfo struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func main() {
	var err error
	db, err = sql.Open("sqlite3", "./forum.db")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()
	createTable := `
    CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        facebook_id TEXT,
        name TEXT,
        email TEXT
    );`
	_, err = db.Exec(createTable)
	if err != nil {
		fmt.Println(err)
		return
	}
	r := mux.NewRouter()
	r.HandleFunc("/", handlers.HomeHandler)
	r.HandleFunc("/auth/facebook", fbLoginHandler)
	r.HandleFunc("/auth/facebook/callback", fbCallbackHandler)
	http.Handle("/", r)
	r.HandleFunc("/register", handlers.RegisterHandler).Methods("GET", "POST")
	r.HandleFunc("/login", handlers.LoginHandler).Methods("GET", "POST")
	r.HandleFunc("/index", handlers.IndexHandler)
	r.HandleFunc("/create-thread", handlers.CreateThreadHandler).Methods("GET", "POST")
	r.HandleFunc("/thread", handlers.ViewThreadHandler).Methods("GET")
	r.HandleFunc("/profile", handlers.ProfileHandler).Methods("GET")
	r.HandleFunc("/logout", handlers.LogoutHandler).Methods("GET")

	// Use mux.HandleFunc for these endpoints
	r.HandleFunc("/create-comment", handlers.CreateCommentHandler).Methods("POST")
	r.HandleFunc("/like-dislike-thread", handlers.LikeThread).Methods("POST")

	fs := http.FileServer(http.Dir("./static"))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func fbLoginHandler(w http.ResponseWriter, r *http.Request) {
	url := fbOauthConfig.AuthCodeURL("randomstate")
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func fbCallbackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	token, err := fbOauthConfig.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Printf("Failed to exchange token: %v\n", err)
		http.Error(w, "Failed to get access token", http.StatusInternalServerError)
		return
	}

	response, err := http.Get("https://graph.facebook.com/me?fields=id,name,email&access_token=" + token.AccessToken)
	if err != nil {
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}
	defer response.Body.Close()

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		http.Error(w, "Failed to read response body", http.StatusInternalServerError)
		return
	}

	var userInfo UserInfo
	err = json.Unmarshal(data, &userInfo)
	if err != nil {
		http.Error(w, "Failed to unmarshal user info", http.StatusInternalServerError)
		return
	}

	var userID int
	err = db.QueryRow("SELECT id FROM users WHERE facebook_id = ?", userInfo.ID).Scan(&userID)
	if err == sql.ErrNoRows {
		res, err := db.Exec("INSERT INTO users (facebook_id, name, email) VALUES (?, ?, ?)", userInfo.ID, userInfo.Name, userInfo.Email)
		if err != nil {
			http.Error(w, "Failed to insert user into database", http.StatusInternalServerError)
			return
		}
		userID64, _ := res.LastInsertId()
		userID = int(userID64)
	} else if err != nil {
		fmt.Printf("Failed to query user from database: %v\n", err)
		http.Error(w, "Failed to query user from database", http.StatusInternalServerError)
		return
	}

	session, _ := store.Get(r, "session-name")
	session.Values["authenticated"] = true
	session.Values["user_id"] = userID
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
