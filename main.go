package main

import (
	"log"
	"net/http"

	"DytForum/database"
	"DytForum/handlers"

	"github.com/gorilla/mux"
)

func main() {
	if err := database.InitDB("forum.db"); err != nil {
		log.Fatal("Failed to initialize database: ", err)
	}
	defer database.DB.Close()

	r := mux.NewRouter()
	r.HandleFunc("/", handlers.HomeHandler)
	r.HandleFunc("/register", handlers.RegisterHandler).Methods("GET", "POST")
	r.HandleFunc("/login", handlers.LoginHandler).Methods("GET", "POST")
	r.HandleFunc("/index", handlers.IndexHandler)
	r.HandleFunc("/create-thread", handlers.CreateThreadHandler).Methods("GET", "POST")
	r.HandleFunc("/thread", handlers.ViewThreadHandler).Methods("GET")
	r.HandleFunc("/profile", handlers.ProfileHandler).Methods("GET")
	r.HandleFunc("/logout", handlers.LogoutHandler).Methods("GET")
	

	// Use mux.HandleFunc for these endpoints
	r.HandleFunc("/create-comment", handlers.CreateCommentHandler).Methods("POST")
	r.HandleFunc("/like-comment", handlers.LikeComment).Methods("POST")
	r.HandleFunc("/like-thread", handlers.LikeThread).Methods("POST")


	fs := http.FileServer(http.Dir("./static"))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))

	log.Println("Server started on :8080")
	http.Handle("/", r)
	log.Fatal(http.ListenAndServe(":8080", r))
}
