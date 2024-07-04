// handlers/thread.go

package handlers

import (
    "html/template"
    "log"
    "net/http"

    "DytForum/database"
    "DytForum/models"
)

func CreateThreadHandler(w http.ResponseWriter, r *http.Request) {
    session, _ := store.Get(r, "session-name")
    if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
        http.Error(w, "You must be logged in to create a thread", http.StatusUnauthorized)
        return
    }

    if r.Method == "GET" {
        tmpl := template.Must(template.ParseFiles("templates/create_thread.html"))
        tmpl.Execute(w, nil)
    } else if r.Method == "POST" {
        category := r.FormValue("category")
        title := r.FormValue("title")
        content := r.FormValue("content")
        username := session.Values["username"].(string)

        var userID int
        err := database.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        _, err = database.DB.Exec("INSERT INTO threads (user_id, category, title, content) VALUES (?, ?, ?, ?)", userID, category, title, content)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        http.Redirect(w, r, "/index", http.StatusSeeOther)
    }
}

func ViewThreadHandler(w http.ResponseWriter, r *http.Request) {
    threadID := r.URL.Query().Get("id")
    if threadID == "" {
        http.Error(w, "Thread ID is required", http.StatusBadRequest)
        return
    }

    log.Printf("Fetching thread with ID: %s", threadID)

    var thread models.Thread
    var username string
    err := database.DB.QueryRow(`
        SELECT t.id, t.title, t.content, t.likes, t.dislikes, u.username, t.category 
        FROM threads t 
        JOIN users u ON t.user_id = u.id 
        WHERE t.id = ?`, threadID).Scan(&thread.ID, &thread.Title, &thread.Content, &thread.Likes, &thread.Dislikes, &username, &thread.Category)
    if err != nil {
        log.Printf("Failed to fetch thread details: %v", err)
        http.Error(w, "Failed to fetch thread", http.StatusInternalServerError)
        return
    }

    rows, err := database.DB.Query("SELECT id, content, user_id, likes, dislikes FROM comments WHERE thread_id = ?", threadID)
    if err != nil {
        log.Printf("Failed to fetch comments: %v", err)
        http.Error(w, "Failed to fetch comments", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var comments []models.Comment
    for rows.Next() {
        var comment models.Comment
        err := rows.Scan(&comment.ID, &comment.Content, &comment.UserID, &comment.Likes, &comment.Dislikes)
        if err != nil {
            log.Printf("Failed to read comment %v", err)
            http.Error(w, "Failed to read comment data", http.StatusInternalServerError)
            return
        }
        comments = append(comments, comment)
    }

    tmpl := template.Must(template.ParseFiles("templates/threads.html"))
    err = tmpl.Execute(w, map[string]interface{}{
        "Thread":   thread,
        "Username": username,
        "Comments": comments,
    })
    if err != nil {
        log.Printf("Failed to execute template: %v", err)
        http.Error(w, "Failed to render page", http.StatusInternalServerError)
    }
}
