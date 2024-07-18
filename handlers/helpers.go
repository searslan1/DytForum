package handlers

import (
	"DytForum/database"
	"DytForum/models"
)

func fetchCategories() ([]models.Category, error) {
	rows, err := database.DB.Query("SELECT id, name FROM categories")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var category models.Category
		err := rows.Scan(&category.ID, &category.Name)
		if err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	return categories, nil
}
