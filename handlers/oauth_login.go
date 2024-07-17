package handlers

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"DytForum/database"
	"DytForum/models"
	"DytForum/session"

	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

var (
	googleOauthConfig   *oauth2.Config
	githubOauthConfig   *oauth2.Config
	facebookOauthConfig *oauth2.Config
)

func GoogleLogin(w http.ResponseWriter, r *http.Request) {
	url := googleOauthConfig.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func GoogleCallback(w http.ResponseWriter, r *http.Request) {
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

	user := models.GoogleUserInfo{
		Name:  googleUserInfo["name"].(string),
		Email: googleUserInfo["email"].(string),
	}

	session, err := session.Store.Get(r, "session-name")
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

	// Kullanıcıyı veritabanında kontrol et ve ekle/güncelle
	_, err = database.DB.Exec(`INSERT INTO users (username, email, password) VALUES (?, ?, 'oauth')
                               ON CONFLICT(email) DO UPDATE SET username=excluded.username`, user.Name, user.Email)
	if err != nil {
		http.Error(w, "Failed to insert/update user", http.StatusInternalServerError)
		log.Printf("Database error: %v", err)
		return
	}

	http.Redirect(w, r, "/index", http.StatusSeeOther)
}

func Profile(w http.ResponseWriter, r *http.Request) {
	session, err := session.Store.Get(r, "session-name")
	if err != nil {
		http.Error(w, "Failed to get session", http.StatusInternalServerError)
		log.Printf("Session error: %v", err)
		return
	}

	// Check for Google access token in session
	googleAccessToken, googleOK := session.Values["googleAccessToken"].(string)

	// Check for GitHub access token in session
	githubAccessToken, githubOK := session.Values["githubAccessToken"].(string)

	// Example: Using Google access token
	if googleOK {
		// Use googleAccessToken to fetch user profile data from Google APIs if needed
		fmt.Fprintf(w, "Google Profile Page\nAccess Token: %s", googleAccessToken)
		return
	}

	// Example: Using GitHub access token
	if githubOK {
		// Use githubAccessToken to fetch user profile data from GitHub APIs if needed
		fmt.Fprintf(w, "GitHub Profile Page\nAccess Token: %s", githubAccessToken)
		return
	}

	http.Error(w, "Access token not found in session", http.StatusInternalServerError)
	log.Println("Access token not found in session")
}

func ProtectedEndpoint(w http.ResponseWriter, r *http.Request) {
	session, err := session.Store.Get(r, "session-name")
	if err != nil {
		http.Error(w, "Failed to get session", http.StatusInternalServerError)
		log.Printf("Error getting session: %v", err) // Log the error
		return
	}

	accessToken, ok := session.Values["accessToken"].(string)
	if !ok {
		http.Error(w, "Access token not found in session", http.StatusInternalServerError)
		log.Println("Access token not found in session") // Log the error
		return
	}

	// Use accessToken to make authenticated requests or perform actions
	fmt.Fprintf(w, "Access Token: %s", accessToken)
}

func GitHubLogin(w http.ResponseWriter, r *http.Request) {
	url := githubOauthConfig.AuthCodeURL("state-token")
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func GitHubCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Code not found", http.StatusBadRequest)
		log.Println("Code not found in URL")
		return
	}
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

	user := models.GitHubUserInfo{
		Login: githubUserInfo["login"].(string),
		Email: "",
	}
	if email, ok := githubUserInfo["email"].(string); ok {
		user.Email = email
	}
	session, err := session.Store.Get(r, "session-name")
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

	// Kullanıcıyı veritabanında kontrol et ve ekle/güncelle
	_, err = database.DB.Exec(`INSERT INTO users (username, email, password) VALUES (?, ?, 'oauth')
                               ON CONFLICT(email) DO UPDATE SET username=excluded.username`, user.Login, user.Email)
	if err != nil {
		http.Error(w, "Failed to insert/update user", http.StatusInternalServerError)
		log.Printf("Database error: %v", err)
		return
	}

	http.Redirect(w, r, "/index", http.StatusSeeOther)
}

func FacebookLogin(w http.ResponseWriter, r *http.Request) {
	url := facebookOauthConfig.AuthCodeURL("")
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func FacebookCallback(w http.ResponseWriter, r *http.Request) {
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

	user := models.FacebookUserInfo{
		Name:  facebookUserInfo["name"].(string),
		Email: facebookUserInfo["email"].(string),
	}

	session, err := session.Store.Get(r, "session-name")
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

	// Kullanıcıyı veritabanında kontrol et ve ekle/güncelle
	_, err = database.DB.Exec(`INSERT INTO users (username, email, password) VALUES (?, ?, 'oauth')
                               ON CONFLICT(email) DO UPDATE SET username=excluded.username`, user.Name, user.Email)
	if err != nil {
		http.Error(w, "Failed to insert/update user", http.StatusInternalServerError)
		log.Printf("Database error: %v", err)
		return
	}

	http.Redirect(w, r, "/index", http.StatusSeeOther)
}

func init() {
	gob.Register(models.User{})

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file %v", err)
	}

	googleOauthConfig = &oauth2.Config{
		RedirectURL:  "http://localhost:8080/auth/google/callback",
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		Scopes:       []string{"profile", "email"},
		Endpoint:     google.Endpoint,
	}
	githubOauthConfig = &oauth2.Config{
		RedirectURL:  "http://localhost:8080/auth/github/callback",
		ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		Scopes:       []string{"user:email"},
		Endpoint:     github.Endpoint,
	}
	facebookOauthConfig = &oauth2.Config{
		ClientID:     os.Getenv("FACEBOOK_KEY"),
		ClientSecret: os.Getenv("FACEBOOK_SECRET"),
		RedirectURL:  "http://localhost:8080/auth/facebook/callback",
		Endpoint:     facebook.Endpoint,
		Scopes:       []string{"email"},
	}
	// SESSION_SECRET'ın doğru yüklendiğini kontrol edelim
	sessionSecret := os.Getenv("SESSION_SECRET")
	if sessionSecret == "" {
		log.Fatal("SESSION_SECRET is not set")
	}
}
