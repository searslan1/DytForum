// main.go
package main

import (
	"log"
	"net/http"

	"DytForum/database"
	"DytForum/handlers"

	"github.com/gorilla/mux"
)

func main() {
	database.InitDB()
	defer database.DB.Close()

	r := mux.NewRouter()
	r.HandleFunc("/", handlers.IndexHandler)
	r.HandleFunc("/register", handlers.RegisterHandler).Methods("GET", "POST")
	r.HandleFunc("/login", handlers.LoginHandler).Methods("GET", "POST")
	r.HandleFunc("/index", handlers.IndexHandler)
	r.HandleFunc("/create-thread", handlers.CreateThreadHandler).Methods("GET", "POST")
	r.HandleFunc("/profile", handlers.ProfileHandler)
	r.HandleFunc("/", handlers.ViewThreadsHandler).Methods("GET")
	fs := http.FileServer(http.Dir("./static"))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))

	log.Println("Server started on :8080")
	http.Handle("/", r)
	http.ListenAndServe(":8080", r)
}
