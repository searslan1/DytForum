package main

import (
	"log"
	"net/http"

	"DytForum/database"
	"DytForum/handlers"
	"DytForum/middleware"
	"DytForum/session"

	"github.com/gorilla/mux"
)

func main() {
	if err := database.InitDB("forum.db"); err != nil {
		log.Fatal("Failed to initialize database: ", err)
	}
	defer database.DB.Close()

	session.Init()

	r := mux.NewRouter()
	r.HandleFunc("/", handlers.HomeHandler)

	// Auth routes
	r.HandleFunc("/auth/github/login", handlers.GitHubLogin)
	r.HandleFunc("/auth/github/callback", handlers.GitHubCallback)
	r.HandleFunc("/auth/google/login", handlers.GoogleLogin)
	r.HandleFunc("/auth/google/callback", handlers.GoogleCallback)
	r.HandleFunc("/auth/facebook", handlers.FacebookLogin)
	r.HandleFunc("/auth/facebook/callback", handlers.FacebookCallback)

	// Login and Register routes
	r.HandleFunc("/register", handlers.RegisterHandler).Methods("GET", "POST")
	r.HandleFunc("/login", handlers.LoginHandler).Methods("GET", "POST")

	// Moderator request route
	r.HandleFunc("/moderator-request", handlers.ModeratorRequestHandler).Methods("GET", "POST")

	// Protected endpoints
	protected := r.NewRoute().Subrouter()
	protected.Use(middleware.AuthMiddleware)
	protected.HandleFunc("/profile", handlers.ProfileHandler)
	protected.HandleFunc("/create-thread", handlers.CreateThreadHandler).Methods("GET", "POST")
	protected.HandleFunc("/create-comment", handlers.CreateCommentHandler).Methods("POST")
	protected.HandleFunc("/like-thread", handlers.LikeThread).Methods("POST")
	protected.HandleFunc("/like-dislike-thread", handlers.LikeThread).Methods("POST")
	protected.HandleFunc("/report-thread", handlers.ReportThreadHandler).Methods("POST")

	// Moderator endpoints
	moderator := r.NewRoute().Subrouter()
	moderator.Use(middleware.ModeratorMiddleware)
	moderator.HandleFunc("/moderator/panel", handlers.ModeratorPanelHandler).Methods("GET")
	moderator.HandleFunc("/moderator/approve-thread/{id:[0-9]+}", handlers.ApproveThreadHandler).Methods("GET")
	moderator.HandleFunc("/moderator/reject-thread/{id:[0-9]+}", handlers.RejectThreadHandler).Methods("GET")
	moderator.HandleFunc("/moderator/reports", handlers.ListReportsHandler).Methods("GET")
	moderator.HandleFunc("/moderator/approve-report/{id:[0-9]+}", handlers.ApproveReportHandler).Methods("GET")
	moderator.HandleFunc("/moderator/reject-report/{id:[0-9]+}", handlers.RejectReportHandler).Methods("GET")
	moderator.HandleFunc("/moderator/delete-thread/{id:[0-9]+}", handlers.DeleteThreadHandler).Methods("GET")

	// Admin endpoints
	admin := r.NewRoute().Subrouter()
	admin.Use(middleware.AdminMiddleware)
	r.HandleFunc("/admin", handlers.AdminLoginHandler).Methods("GET", "POST")
	admin.HandleFunc("/admin/panel", handlers.AdminPanelHandler).Methods("GET")
	admin.HandleFunc("/admin/moderator-requests", handlers.ListModeratorRequestsHandler).Methods("GET")
	admin.HandleFunc("/admin/approve-moderator/{id:[0-9]+}", handlers.ApproveModeratorHandler).Methods("GET")
	admin.HandleFunc("/admin/reject-moderator/{id:[0-9]+}", handlers.RejectModeratorHandler).Methods("GET")
	admin.HandleFunc("/admin/promote-user/{id:[0-9]+}", handlers.PromoteUserHandler).Methods("GET")
	admin.HandleFunc("/admin/demote-user/{id:[0-9]+}", handlers.DemoteUserHandler).Methods("GET")
	admin.HandleFunc("/admin/create-category", handlers.CreateCategoryHandler).Methods("POST")
	admin.HandleFunc("/admin/delete-category", handlers.DeleteCategoryHandler).Methods("POST")
	admin.HandleFunc("/logout", handlers.AdminLogoutHandler).Methods("GET")

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
