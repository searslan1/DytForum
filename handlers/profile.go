package handlers

import (
	"html/template"
	"net/http"

	"DytForum/models"
)

func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "session-name")
	if err != nil {
		http.Error(w, "Failed to get session", http.StatusInternalServerError)
		return
	}

	googleUserInfo, googleOk := session.Values["googleUserInfo"].(models.GoogleUserInfo)
	githubUserInfo, githubOk := session.Values["githubUserInfo"].(models.GitHubUserInfo)
	facebookUserInfo, facebookOk := session.Values["facebookUserInfo"].(models.FacebookUserInfo)

	var username string
	if googleOk {
		username = googleUserInfo.Name
	} else if githubOk {
		username = githubUserInfo.Login
	} else if facebookOk {
		username = facebookUserInfo.Name
	} else {
		http.Error(w, "User not logged in", http.StatusUnauthorized)
		return
	}
	// Kullanıcının thread'lerini ve yorumlarını al
	threads := getThreadsByUser(username)
	comments := getCommentsByUser(username)

	data := struct {
		Username         string
		GoogleUserInfo   models.GoogleUserInfo
		GitHubUserInfo   models.GitHubUserInfo
		FacebookUserInfo models.FacebookUserInfo
		Threads          []models.Thread
		Comments         []models.Comment
	}{
		Username:         username,
		GoogleUserInfo:   googleUserInfo,
		GitHubUserInfo:   githubUserInfo,
		FacebookUserInfo: facebookUserInfo,
		Threads:          threads,
		Comments:         comments,
	}

	tmpl, err := template.ParseFiles("templates/profile.html")
	if err != nil {
		http.Error(w, "Failed to parse template", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Failed to execute template", http.StatusInternalServerError)
		return
	}
}
