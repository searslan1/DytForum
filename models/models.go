// models/models.go
package models

type User struct {
	ID         int
	Username   string
	Email      string
	Password   string
	GoogleID   string
	GitHubID   int
	FacebookID string
	Role       string
}

type Thread struct {
	ID       int
	UserID   int
	Category string
	Title    string
	Content  string
	Likes    int
	Dislikes int
	Username string
	Comments []Comment
	Approved int
}

type Comment struct {
	ID       int
	ThreadID int
	UserID   int
	Content  string
	Likes    int
	Dislikes int
	Username string
}

type Like struct {
	ID       int
	ThreadID int
	UserID   int
	Like     int
}
type GoogleUserInfo struct {
	ID    string
	Name  string
	Email string
}

type GitHubUserInfo struct {
	ID    int
	Login string
	Email string
}

type FacebookUserInfo struct {
	ID    string
	Name  string
	Email string
}
type ModeratorRequest struct {
	ID       int
	UserID   int
	Username string
	Reason   string
	Status   string
}
type ProfileData struct {
	Username         string
	Email            string
	Role             string
	GoogleUserInfo   GoogleUserInfo
	GitHubUserInfo   GitHubUserInfo
	FacebookUserInfo FacebookUserInfo
	Threads          []Thread
	Comments         []Comment
}
type Report struct {
	ID       int
	ThreadID int
	UserID   int
	Reason   string
}
