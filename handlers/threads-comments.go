package handlers

import (
	"log"

	"DytForum/database"
	"DytForum/models"
)

// getThreadsByUser veritabanından kullanıcıya ait thread'leri çekmek için fonksiyon
func getThreadsByUser(username string) []models.Thread {
	// Kullanıcı adını kullanıcı kimliğine çevirme
	var user_id int
	err := database.DB.QueryRow(`SELECT id FROM users WHERE username = ?`, username).Scan(&user_id)
	if err != nil {
		log.Printf("Failed to get user ID: %v", err)
		return nil
	}

	rows, err := database.DB.Query(`SELECT id, category, title, content, likes, dislikes FROM threads WHERE user_id = ?`, user_id)
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

// getCommentsByUser veritabanından kullanıcıya ait yorumları çekmek için fonksiyon
func getCommentsByUser(username string) []models.Comment {
	// Kullanıcı adını kullanıcı kimliğine çevirme
	var user_id int
	err := database.DB.QueryRow(`SELECT id FROM users WHERE username = ?`, username).Scan(&user_id)
	if err != nil {
		log.Printf("Failed to get user ID: %v", err)
		return nil
	}

	rows, err := database.DB.Query(`SELECT id, content, thread_id FROM comments WHERE user_id = ?`, user_id)
	if err != nil {
		log.Printf("Failed to query comments: %v", err)
		return nil
	}
	defer rows.Close()

	var comments []models.Comment
	for rows.Next() {
		var comment models.Comment
		if err := rows.Scan(&comment.ID, &comment.Content, &comment.ThreadID); err != nil {
			log.Printf("Failed to scan comment: %v", err)
			continue
		}
		// Thread başlığını çekmek için başka bir sorgu
		err := database.DB.QueryRow(`SELECT title FROM threads WHERE id = ?`, comment.ThreadID).Scan(&comment.ThreadTitle)
		if err != nil {
			log.Printf("Failed to get thread title for comment: %v", err)
			continue
		}
		comments = append(comments, comment)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating over comment rows: %v", err)
	}

	return comments
}
