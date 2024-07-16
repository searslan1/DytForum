package main

import (
	"encoding/gob"
	"log"
	"net/http"

	"DytForum/database"
	"DytForum/handlers"
	"DytForum/middleware"
	"DytForum/models"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func init() {
	// models paketinizdeki struct'ları gob'a kaydedin
	gob.Register(models.GoogleUserInfo{})
	gob.Register(models.GitHubUserInfo{})
	gob.Register(models.FacebookUserInfo{})
}

func main() {
	if err := database.InitDB("forum.db"); err != nil {
		log.Fatal("Failed to initialize database: ", err)
	}
	defer database.DB.Close()

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/", handlers.HomeHandler)
	// Auth routes
	r.HandleFunc("/auth/github/login", handlers.GitHubLogin)
	r.HandleFunc("/auth/github/callback", handlers.GitHubCallback)
	r.HandleFunc("/auth/google/login", handlers.GoogleLogin)
	r.HandleFunc("/auth/google/callback", handlers.GoogleCallback)
	r.HandleFunc("/auth/facebook", handlers.FacebookLogin)
	r.HandleFunc("/auth/facebook/callback", handlers.FacebookCallback)

	// Protected endpoints
	protected := r.NewRoute().Subrouter()
	protected.Use(middleware.AuthMiddleware)
	protected.HandleFunc("/profile", handlers.ProfileHandler)
	protected.HandleFunc("/create-thread", handlers.CreateThreadHandler).Methods("GET", "POST")
	protected.HandleFunc("/create-comment", handlers.CreateCommentHandler).Methods("POST")
	protected.HandleFunc("/like-thread", handlers.LikeThread).Methods("POST")
	protected.HandleFunc("/like-dislike-thread", handlers.LikeThread).Methods("POST")

	// Public endpoints
	public := r.NewRoute().Subrouter()
	public.HandleFunc("/index", handlers.IndexHandler)
	public.HandleFunc("/thread", handlers.ViewThreadHandler).Methods("GET")
	public.HandleFunc("/profile", handlers.ProfileHandler).Methods("GET")
	public.HandleFunc("/logout", handlers.LogoutHandler).Methods("GET")

	// Static files
	fs := http.FileServer(http.Dir("./static"))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))

	// Start server
	log.Println("Server started on :8080")
	http.Handle("/", r)
	log.Fatal(http.ListenAndServe(":8080", r))
}
