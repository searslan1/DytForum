package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"DytForum/database"
	"DytForum/models"
)

func GoogleRegisterCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	token, err := googleOauthConfig.Exchange(r.Context(), code)
	if err != nil {
		http.Error(w, "Failed to exchange Google token", http.StatusInternalServerError)
		log.Printf("Google token exchange error: %v", err)
		return
	}

	client := googleOauthConfig.Client(r.Context(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		http.Error(w, "Failed to get Google user info", http.StatusInternalServerError)
		log.Printf("Google user info error: %v", err)
		return
	}
	defer resp.Body.Close()

	var googleUserInfo map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&googleUserInfo); err != nil {
		http.Error(w, "Failed to decode Google user info", http.StatusInternalServerError)
		log.Printf("Google user info decode error: %v", err)
		return
	}

	// Google'dan gelen kullanıcı bilgilerini al
	user := models.GoogleUserInfo{
		Name:  googleUserInfo["name"].(string),
		Email: googleUserInfo["email"].(string),
	}

	// Eğer bu e-posta ile zaten bir kullanıcı varsa hata ver
	var existingEmail string
	err = database.DB.QueryRow("SELECT email FROM users WHERE email = ?", user.Email).Scan(&existingEmail)
	if err == nil {
		http.Error(w, "Email address already in use", http.StatusBadRequest)
		return
	}

	// Kullanıcıyı veritabanına kaydet
	_, err = database.DB.Exec("INSERT INTO users (email, username, password) VALUES (?, ?, 'oauth')", user.Email, user.Name)
	if err != nil {
		http.Error(w, "Failed to insert user", http.StatusInternalServerError)
		log.Printf("Database error: %v", err)
		return
	}

	// Oturumu başlat ve kullanıcıyı yönlendir
	session, err := store.Get(r, "session-name")
	if err != nil {
		http.Error(w, "Failed to get session", http.StatusInternalServerError)
		log.Printf("Session error: %v", err)
		return
	}
	session.Values["username"] = user.Name
	session.Values["authenticated"] = true
	if err := session.Save(r, w); err != nil {
		http.Error(w, "Failed to save session", http.StatusInternalServerError)
		log.Printf("Session save error: %v", err)
		return
	}

	http.Redirect(w, r, "/index", http.StatusSeeOther)
}

func GitHubRegisterCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	token, err := githubOauthConfig.Exchange(r.Context(), code)
	if err != nil {
		http.Error(w, "Failed to exchange GitHub token", http.StatusInternalServerError)
		log.Printf("GitHub token exchange error: %v", err)
		return
	}

	client := githubOauthConfig.Client(r.Context(), token)
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		http.Error(w, "Failed to get GitHub user info", http.StatusInternalServerError)
		log.Printf("GitHub user info error: %v", err)
		return
	}
	defer resp.Body.Close()

	var githubUserInfo map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&githubUserInfo); err != nil {
		http.Error(w, "Failed to decode GitHub user info", http.StatusInternalServerError)
		log.Printf("GitHub user info decode error: %v", err)
		return
	}

	// GitHub'dan gelen kullanıcı bilgilerini al
	user := models.GitHubUserInfo{
		Login: githubUserInfo["login"].(string),
		Email: "",
	}
	if email, ok := githubUserInfo["email"].(string); ok {
		user.Email = email
	}

	// Eğer bu e-posta ile zaten bir kullanıcı varsa hata ver
	var existingEmail string
	err = database.DB.QueryRow("SELECT email FROM users WHERE email = ?", user.Email).Scan(&existingEmail)
	if err == nil {
		http.Error(w, "Email address already in use", http.StatusBadRequest)
		return
	}

	// Kullanıcıyı veritabanına kaydet
	_, err = database.DB.Exec("INSERT INTO users (email, username, password) VALUES (?, ?, 'oauth')", user.Email, user.Login)
	if err != nil {
		http.Error(w, "Failed to insert user", http.StatusInternalServerError)
		log.Printf("Database error: %v", err)
		return
	}

	// Oturumu başlat ve kullanıcıyı yönlendir
	session, err := store.Get(r, "session-name")
	if err != nil {
		http.Error(w, "Failed to get session", http.StatusInternalServerError)
		log.Printf("Session error: %v", err)
		return
	}
	session.Values["username"] = user.Login
	session.Values["authenticated"] = true
	if err := session.Save(r, w); err != nil {
		http.Error(w, "Failed to save session", http.StatusInternalServerError)
		log.Printf("Session save error: %v", err)
		return
	}

	http.Redirect(w, r, "/index", http.StatusSeeOther)
}

func FacebookRegisterCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	token, err := facebookOauthConfig.Exchange(r.Context(), code)
	if err != nil {
		http.Error(w, "Failed to exchange Facebook token", http.StatusInternalServerError)
		log.Printf("Facebook token exchange error: %v", err)
		return
	}

	client := facebookOauthConfig.Client(r.Context(), token)
	resp, err := client.Get("https://graph.facebook.com/me?fields=id,name,email")
	if err != nil {
		http.Error(w, "Failed to get Facebook user info", http.StatusInternalServerError)
		log.Printf("Facebook user info error: %v", err)
		return
	}
	defer resp.Body.Close()

	var facebookUserInfo map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&facebookUserInfo); err != nil {
		http.Error(w, "Failed to decode Facebook user info", http.StatusInternalServerError)
		log.Printf("Facebook user info decode error: %v", err)
		return
	}

	// Facebook'tan gelen kullanıcı bilgilerini al
	user := models.FacebookUserInfo{
		Name:  facebookUserInfo["name"].(string),
		Email: facebookUserInfo["email"].(string),
	}

	// Eğer bu e-posta ile zaten bir kullanıcı varsa hata ver
	var existingEmail string
	err = database.DB.QueryRow("SELECT email FROM users WHERE email = ?", user.Email).Scan(&existingEmail)
	if err == nil {
		http.Error(w, "Email address already in use", http.StatusBadRequest)
		return
	}

	// Kullanıcıyı veritabanına kaydet
	_, err = database.DB.Exec("INSERT INTO users (email, username, password) VALUES (?, ?, 'oauth')", user.Email, user.Name)
	if err != nil {
		http.Error(w, "Failed to insert user", http.StatusInternalServerError)
		log.Printf("Database error: %v", err)
		return
	}

	// Oturumu başlat ve kullanıcıyı yönlendir
	session, err := store.Get(r, "session-name")
	if err != nil {
		http.Error(w, "Failed to get session", http.StatusInternalServerError)
		log.Printf("Session error: %v", err)
		return
	}
	session.Values["username"] = user.Name
	session.Values["authenticated"] = true
	if err := session.Save(r, w); err != nil {
		http.Error(w, "Failed to save session", http.StatusInternalServerError)
		log.Printf("Session save error: %v", err)
		return
	}

	http.Redirect(w, r, "/index", http.StatusSeeOther)
}
