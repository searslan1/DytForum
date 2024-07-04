// database/thread.go
package database

import (
	"DytForum/models"
)

func GetThreadsByUserID(userID int) ([]models.Thread, error) {
	var threads []models.Thread

	rows, err := DB.Query("SELECT id, title, content, category FROM threads WHERE user_id = ?", userID)
	if err != nil {
		return threads, err
	}
	defer rows.Close()

	for rows.Next() {
		var thread models.Thread
		err := rows.Scan(&thread.ID, &thread.Title, &thread.Content, &thread.Category)
		if err != nil {
			return threads, err
		}
		threads = append(threads, thread)
	}

	return threads, nil
}
