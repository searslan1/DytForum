package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

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
	//store               *sessions.CookieStore
)

func GoogleLogin(w http.ResponseWriter, r *http.Request) {
	url := googleOauthConfig.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func GoogleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	log.Printf("Google callback code: %s", code)

	// Exchange code for access token
	token, err := googleOauthConfig.Exchange(r.Context(), code)
	if err != nil {
		http.Error(w, "Failed to exchange Google token", http.StatusInternalServerError)
		log.Printf("Google token exchange error: %v", err)
		return
	}

	// Store token in session
	session, err := store.Get(r, "session-name")
	if err != nil {
		http.Error(w, "Failed to get session", http.StatusInternalServerError)
		log.Printf("Session error: %v", err)
		return
	}
	session.Values["googleAccessToken"] = token.AccessToken
	if err := session.Save(r, w); err != nil {
		http.Error(w, "Failed to save session", http.StatusInternalServerError)
		log.Printf("Session save error: %v", err)
		return
	}

	// Redirect to profile page after successful login
	http.Redirect(w, r, "/index", http.StatusSeeOther)
}

func Profile(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "session-name")
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
	session, err := store.Get(r, "session-name")
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
	log.Printf("GitHub callback code: %s", code) // Log the received code

	token, err := githubOauthConfig.Exchange(r.Context(), code)
	if err != nil {
		http.Error(w, "Failed to exchange GitHub token", http.StatusInternalServerError)
		log.Printf("GitHub token exchange error: %v", err) // Log the error
		return
	}

	// Store token in session
	session, err := store.Get(r, "session-name")
	if err != nil {
		http.Error(w, "Failed to get session", http.StatusInternalServerError)
		log.Printf("Session error: %v", err) // Log the error
		return
	}
	session.Values["githubAccessToken"] = token.AccessToken
	if err := session.Save(r, w); err != nil {
		http.Error(w, "Failed to save session", http.StatusInternalServerError)
		log.Printf("Session save error: %v", err) // Log the error
		return
	}

	// Redirect to another endpoint after successful login
	http.Redirect(w, r, "/index", http.StatusSeeOther)
}

func FacebookLogin(w http.ResponseWriter, r *http.Request) {
	url := facebookOauthConfig.AuthCodeURL("")
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func FacebookCallback(w http.ResponseWriter, r *http.Request) {
	code := r.FormValue("code")
	token, err := facebookOauthConfig.Exchange(r.Context(), code)
	if err != nil {
		log.Printf("Facebook token exchange error: %v", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	resp, err := http.Get(fmt.Sprintf("https://graph.facebook.com/me?access_token=%s&fields=id,name,email", token.AccessToken))
	if err != nil {
		log.Printf("Get: %s\n", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	defer resp.Body.Close()

	var user struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		log.Printf("Decode: %s\n", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Store the user data in the session
	session, err := store.Get(r, "session-name")
	if err != nil {
		http.Error(w, "Failed to get session", http.StatusInternalServerError)
		log.Printf("Session error: %v", err)
		return
	}
	session.Values["user"] = user
	if err := session.Save(r, w); err != nil {
		http.Error(w, "Failed to save session", http.StatusInternalServerError)
		log.Printf("Session save error: %v", err)
		return
	}

	// Redirect to profile page after successful login
	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}

func init() {
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
