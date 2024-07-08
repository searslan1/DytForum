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
		Threads:          getThreadsByUser(username),
		Comments:         getCommentsByUser(username),
	}

	tmpl, err := template.ParseFiles("templates/profile.html")
	if err != nil {
		http.Error(w, "Failed to parse template", http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, data)
}

func getThreadsByUser(username string) []models.Thread {
	// Veritabanından kullanıcıya ait thread'leri çekmek için gerekli işlemleri yapın
	// Bu örnekte, boş bir slice döndürüyoruz
	return []models.Thread{
		{Title: "Thread 1", Content: "Content of thread 1"},
		{Title: "Thread 2", Content: "Content of thread 2"},
	}
}

func getCommentsByUser(username string) []models.Comment {
	// Veritabanından kullanıcıya ait yorumları çekmek için gerekli işlemleri yapın
	// Bu örnekte, boş bir slice döndürüyoruz
	return []models.Comment{
		{Content: "Comment 1", ThreadTitle: "Thread 1"},
		{Content: "Comment 2", ThreadTitle: "Thread 2"},
	}
}
