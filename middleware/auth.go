package middleware

import (
	"log"
	"net/http"

	"DytForum/session"
)

// AuthMiddleware kontrol eder
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := session.Store.Get(r, "session-name")
		if err != nil {
			log.Printf("AuthMiddleware: Failed to get session: %v", err)
			http.Error(w, "Failed to get session", http.StatusInternalServerError)
			return
		}

		auth, ok := session.Values["authenticated"].(bool)
		if !ok || !auth {
			http.Error(w, "You must be logged in to access this page", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// AdminMiddleware kontrol eder
func AdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := session.Store.Get(r, "session-name")
		if err != nil {
			log.Printf("AdminMiddleware: Failed to get session: %v", err)
			http.Error(w, "Failed to get session", http.StatusInternalServerError)
			return
		}

		role, ok := session.Values["role"].(string)
		if !ok || role != "admin" {
			http.Error(w, "You must be an admin to access this page", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// ModeratorMiddleware kontrol eder
func ModeratorMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := session.Store.Get(r, "session-name")
		if err != nil {
			log.Printf("ModeratorMiddleware: Failed to get session: %v", err)
			http.Error(w, "Failed to get session", http.StatusInternalServerError)
			return
		}

		role, ok := session.Values["role"].(string)
		if !ok || role != "moderator" {
			http.Error(w, "You must be a moderator to access this page", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// GuestMiddleware ayarlar
func GuestMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _ := session.Store.Get(r, "session-name")
		if auth, ok := session.Values["authenticated"].(bool); ok && auth {
			http.Redirect(w, r, "/index", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}
