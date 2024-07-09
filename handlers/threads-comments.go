package handlers

import (
	"log"

	"DytForum/database"
	"DytForum/models"
)

// getThreadsByUser veritabanından kullanıcıya ait thread'leri çekmek için fonksiyon
func getThreadsByUser(username string) []models.Thread {
	// Kullanıcı adını kullanıcı kimliğine çevirme
	var userID int
	err := database.DB.QueryRow(`SELECT id FROM users WHERE username = ?`, username).Scan(&userID)
	if err != nil {
		log.Printf("Failed to get user ID: %v", err)
		return nil
	}

	rows, err := database.DB.Query(`SELECT id, category, title, content, likes, dislikes FROM threads WHERE user_id = ?`, userID)
	if err != nil {
		log.Printf("Failed to query threads: %v", err)
		return nil
	}
	defer rows.Close()

	var threads []models.Thread
	for rows.Next() {
		var thread models.Thread
		if err := rows.Scan(&thread.ID, &thread.Category, &thread.Title, &thread.Content, &thread.Likes, &thread.Dislikes); err != nil {
			log.Printf("Failed to scan thread: %v", err)
			continue
		}
		threads = append(threads, thread)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating over thread rows: %v", err)
	}

	return threads
}

func getCommentsByUser(username string) []models.Comment {
	// Kullanıcı adını kullanıcı kimliğine çevirme
	var userID int
	err := database.DB.QueryRow(`SELECT id FROM users WHERE username = ?`, username).Scan(&userID)
	if err != nil {
		log.Printf("Failed to get user ID: %v", err)
		return nil
	}

	rows, err := database.DB.Query(`SELECT c.id, c.thread_id, c.content FROM comments c INNER JOIN threads t ON c.thread_id = t.id WHERE t.user_id = ? AND c.user_id = ?`, userID, userID)
	if err != nil {
		log.Printf("Failed to query comments: %v", err)
		return nil
	}
	defer rows.Close()

	var comments []models.Comment
	for rows.Next() {
		var comment models.Comment
		if err := rows.Scan(&comment.ID, &comment.ThreadID, &comment.Content); err != nil {
			log.Printf("Failed to scan comment: %v", err)
			continue
		}
		comments = append(comments, comment)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating over comment rows: %v", err)
	}

	return comments
}
