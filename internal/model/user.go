package model

type User struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	IsAdmin   bool   `json:"is_admin"`
	CreatedAt string `json:"created_at"`
}
