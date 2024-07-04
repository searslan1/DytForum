// database/user.go
package database

import (
	"DytForum/models"
)

func GetUserByUsername(username string) (models.User, error) {
	var user models.User

	err := DB.QueryRow("SELECT id, username, email FROM users WHERE username = ?", username).Scan(&user.ID, &user.Username, &user.Email)
	if err != nil {
		return user, err
	}

	return user, nil
}
