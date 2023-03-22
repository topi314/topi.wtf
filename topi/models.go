package topi

import (
	"html/template"
	"time"
)

type Variables struct {
	User     User
	Home     Home
	Posts    []Post
	Projects []Project
	Dark     bool
	CSS      template.CSS
}

type Home struct {
	Body    string
	Content template.HTML
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
	AvatarURL template.URL
	CreatedAt time.Time
	Content   template.HTML
	Body      string
}

type User struct {
	Name      string
	AvatarURL template.URL
}

type Project struct {
	Name        string
	Description string
	URL         template.URL
	Stars       int
	Forks       int
	UpdatedAt   time.Time
	Language    *Language
	Topics      []Topic
}

type Language struct {
	Name  string
	Color string
}

type Topic struct {
	Name string
	URL  string
}
