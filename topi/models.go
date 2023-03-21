package topi

import (
	"html/template"
	"time"
)

type Variables struct {
	Blog   Blog
	GitHub GitHub
	Dark   bool
	CSS    template.CSS
}

type Blog struct {
	Posts []Post
}

type Post struct {
	Title     string
	CreatedAt time.Time
	URL       template.URL
	Content   template.HTML
	Body      string
	Comments  []Comment
}

type Comment struct {
	Author    string
	AvatarURL string
	CreatedAt time.Time
	Content   template.HTML
	Body      string
}

type GitHub struct {
	User     User
	Projects []Project
}

type User struct {
	Name      string
	AvatarURL string
}

type Project struct {
	Name        string
	Description string
	URL         string
	Languages   []Language
}

type Language struct {
	Name  string
	Color string
}
