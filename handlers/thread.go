// handlers/thread.go
package handlers

import (
    "html/template"
    "net/http"
    "strconv"

    "DytForum/database"
    "DytForum/models"
)

// CreateThreadHandler handles the creation of a new thread.
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

// ViewThreadHandler handles the display of a thread with its comments.
func ViewThreadHandler(w http.ResponseWriter, r *http.Request) {
    threadIDStr := r.URL.Query().Get("id")
    if threadIDStr == "" {
        http.Error(w, "Missing thread ID", http.StatusBadRequest)
        return
    }

    threadID, err := strconv.Atoi(threadIDStr)
    if err != nil {
        http.Error(w, "Invalid thread ID", http.StatusBadRequest)
        return
    }

    // Fetch the thread details
    var thread models.Thread
    err = database.DB.QueryRow("SELECT id, title, content, category, likes, dislikes, user_id FROM threads WHERE id = ?", threadID).Scan(&thread.ID, &thread.Title, &thread.Content, &thread.Category, &thread.Likes, &thread.Dislikes, &thread.UserID)
    if err != nil {
        http.Error(w, "Failed to fetch thread details", http.StatusInternalServerError)
        return
    }

    // Fetch the comments for the thread
    rows, err := database.DB.Query("SELECT id, user_id, content, username, likes, dislikes FROM comments WHERE thread_id = ?", threadID)
    if err != nil {
        http.Error(w, "Failed to fetch comments", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var comments []models.Comment
    for rows.Next() {
        var comment models.Comment
        if err := rows.Scan(&comment.ID, &comment.UserID, &comment.Content, &comment.Username, &comment.Likes, &comment.Dislikes); err != nil {
            http.Error(w, "Failed to scan comment", http.StatusInternalServerError)
            return
        }
        comment.ThreadID = threadID
        comments = append(comments, comment)
    }

    // Fetch the username of the thread creator
    var creatorUsername string
    err = database.DB.QueryRow("SELECT username FROM users WHERE id = ?", thread.UserID).Scan(&creatorUsername)
    if err != nil {
        http.Error(w, "Failed to fetch thread creator username", http.StatusInternalServerError)
        return
    }

    // Render the thread page with comments
    data := struct {
        Thread    models.Thread
        Comments  []models.Comment
        Username  string
    }{
        Thread:   thread,
        Comments: comments,
        Username: creatorUsername,
    }
    err = renderTemplate(w, "threads.html", data)
    if err != nil {
        http.Error(w, "Failed to render template", http.StatusInternalServerError)
    }
}

// renderTemplate renders the HTML templates.
func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) error {
    templates := template.Must(template.ParseGlob("templates/*.html"))
    return templates.ExecuteTemplate(w, tmpl, data)
}
