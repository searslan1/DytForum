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
}

type Comment struct {
	ID          int
	ThreadID    int
	UserID      int
	Content     string
	ThreadTitle string
}

type Like struct {
	ID       int
	ThreadID int
	UserID   int
	Like     int
}
