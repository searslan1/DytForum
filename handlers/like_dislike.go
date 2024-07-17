package handlers

import (
	"log"
	"net/http"
	"strconv"

	"DytForum/database"
	"DytForum/session"
)

// LikeThread handles HTTP requests to like or dislike a thread.
func LikeThread(w http.ResponseWriter, r *http.Request) {
	session, _ := session.Store.Get(r, "session-name")
	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		http.Error(w, "You must be logged in to like or dislike a thread", http.StatusUnauthorized)
		return
	}

	userID, ok := session.Values["userID"].(int)
	if !ok {
		log.Println("User ID not found in session")
		http.Error(w, "User ID not found in session", http.StatusInternalServerError)
		return
	}

	log.Printf("User ID: %d", userID)

	threadID := r.FormValue("thread_id")
	likeStatus := r.FormValue("like_status")

	if threadID == "" || likeStatus == "" {
		http.Error(w, "Thread ID and like/dislike status are required", http.StatusBadRequest)
		return
	}

	likeStatusInt, err := strconv.Atoi(likeStatus)
	if err != nil || (likeStatusInt != 1 && likeStatusInt != -1) {
		http.Error(w, "Invalid like/dislike status", http.StatusBadRequest)
		return
	}

	// Check if the user has already liked or disliked the thread
	var existingLikeStatus int
	err = database.DB.QueryRow("SELECT like_status FROM likes WHERE thread_id = ? AND user_id = ?", threadID, userID).Scan(&existingLikeStatus)
	if err == nil {
		if existingLikeStatus == likeStatusInt {
			// User is trying to like or dislike the thread again with the same status, do nothing
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther) // Redirect back to the referring page
			return
		} else {
			// User is changing their like/dislike status
			if existingLikeStatus != 0 {
				// Delete the previous like/dislike if it exists
				_, err := database.DB.Exec("DELETE FROM likes WHERE thread_id = ? AND user_id = ?", threadID, userID)
				if err != nil {
					log.Printf("Failed to delete previous like/dislike: %v", err)
					http.Error(w, "Failed to delete previous like/dislike", http.StatusInternalServerError)
					return
				}

				// Update the thread's likes/dislikes count based on the previous status
				if existingLikeStatus == 1 {
					_, err = database.DB.Exec("UPDATE threads SET likes = likes - 1 WHERE id = ?", threadID)
				} else {
					_, err = database.DB.Exec("UPDATE threads SET dislikes = dislikes - 1 WHERE id = ?", threadID)
				}
				if err != nil {
					log.Printf("Failed to update thread likes/dislikes: %v", err)
					http.Error(w, "Failed to update thread likes/dislikes", http.StatusInternalServerError)
					return
				}
			}
		}
	}

	// Insert new like/dislike
	_, err = database.DB.Exec("INSERT INTO likes (thread_id, user_id, like_status) VALUES (?, ?, ?)", threadID, userID, likeStatusInt)
	if err != nil {
		log.Printf("Failed to create like/dislike: %v", err)
		http.Error(w, "Failed to create like/dislike", http.StatusInternalServerError)
		return
	}

	// Update the thread's likes/dislikes count based on the new status
	if likeStatusInt == 1 {
		_, err = database.DB.Exec("UPDATE threads SET likes = likes + 1 WHERE id = ?", threadID)
	} else {
		_, err = database.DB.Exec("UPDATE threads SET dislikes = dislikes + 1 WHERE id = ?", threadID)
	}
	if err != nil {
		log.Printf("Failed to update thread likes/dislikes: %v", err)
		http.Error(w, "Failed to update thread likes/dislikes", http.StatusInternalServerError)
		return
	}

	// Redirect back to the referring page after successful like/dislike
	http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
}

// LikeComment handles HTTP requests to like or dislike a comment.
func LikeComment(w http.ResponseWriter, r *http.Request) {
	session, _ := session.Store.Get(r, "session-name")
	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		http.Error(w, "You must be logged in to like or dislike a comment", http.StatusUnauthorized)
		return
	}

	userID, ok := session.Values["userID"].(int)
	if !ok {
		log.Println("User ID not found in session") // Added logging
		http.Error(w, "User ID not found in session", http.StatusInternalServerError)
		return
	}

	log.Printf("User ID: %d", userID) // Added logging

	commentID := r.FormValue("comment_id")
	likeStatus := r.FormValue("like_status")

	if commentID == "" || likeStatus == "" {
		http.Error(w, "Comment ID and like/dislike status are required", http.StatusBadRequest)
		return
	}

	likeStatusInt, err := strconv.Atoi(likeStatus)
	if err != nil || (likeStatusInt != 1 && likeStatusInt != -1) {
		http.Error(w, "Invalid like/dislike status", http.StatusBadRequest)
		return
	}

	// Check if the user has already liked or disliked the comment
	var existingLikeStatus int
	err = database.DB.QueryRow("SELECT like_status FROM likes WHERE comment_id = ? AND user_id = ?", commentID, userID).Scan(&existingLikeStatus)
	if err == nil {
		if existingLikeStatus == likeStatusInt {
			// User is trying to like or dislike the comment again with the same status, do nothing
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther) // Redirect back to the referring page
			return
		} else {
			// User is changing their like/dislike status
			_, err := database.DB.Exec("UPDATE likes SET like_status = ? WHERE comment_id = ? AND user_id = ?", likeStatusInt, commentID, userID)
			if err != nil {
				log.Printf("Failed to update like/dislike: %v", err)
				http.Error(w, "Failed to update like/dislike", http.StatusInternalServerError)
				return
			}

			// Delete the previous like/dislike if it exists
			if existingLikeStatus != 0 {
				var updateCount int
				if existingLikeStatus == 1 {
					updateCount = -1
				} else {
					updateCount = 1
				}
				_, err := database.DB.Exec("UPDATE comments SET likes = likes + ?, dislikes = dislikes - ? WHERE id = ?", updateCount, updateCount, commentID)
				if err != nil {
					log.Printf("Failed to update comment likes/dislikes: %v", err)
					http.Error(w, "Failed to update comment likes/dislikes", http.StatusInternalServerError)
					return
				}
			}
		}
	} else {
		// User hasn't liked or disliked the comment yet, insert new like/dislike
		_, err := database.DB.Exec("INSERT INTO likes (comment_id, user_id, like_status) VALUES (?, ?, ?)", commentID, userID, likeStatusInt)
		if err != nil {
			log.Printf("Failed to create like/dislike: %v", err)
			http.Error(w, "Failed to create like/dislike", http.StatusInternalServerError)
			return
		}

		// Update the comment's likes/dislikes count based on the new status
		if likeStatusInt == 1 {
			_, err = database.DB.Exec("UPDATE comments SET likes = likes + 1 WHERE id = ?", commentID)
		} else {
			_, err = database.DB.Exec("UPDATE comments SET dislikes = dislikes + 1 WHERE id = ?", commentID)
		}
		if err != nil {
			log.Printf("Failed to update comment likes/dislikes: %v", err)
			http.Error(w, "Failed to update comment likes/dislikes", http.StatusInternalServerError)
			return
		}
	}

	// Redirect back to the referring page after successful like/dislike
	http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
}
