// models/models.go
package models

type User struct {
	ID       int
	Username string
	Email    string
	Password string
}

type Thread struct {
	ID       int
	UserID   int
	Category string
	Title    string
	Content  string
	Likes     int
	Dislikes  int
	Username  string
	Comments  []Comment
}

type Comment struct {
	ID          int
	ThreadID    int
	UserID      int
	Content     string
	ThreadTitle string // gerekli mi bu ?
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
