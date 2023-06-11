package topi

import (
	"html/template"
	"time"
)

type Variables struct {
	User          User
	Home          Home
	Posts         []Post
	PostsAfter    string
	Projects      []Project
	ProjectsAfter string
	Dark          bool
	Description   string
	CSS           template.CSS
}

type Home struct {
	Body    string
	Content template.HTML
	LastFM  LastFM
}

type Post struct {
	Title     string
	CreatedAt time.Time
	URL       template.URL
	Content   template.HTML
	Body      string
	Upvotes   int
	Comments  []Comment
}

type Comment struct {
	Author    string
	AvatarURL template.URL
	CreatedAt time.Time
	Upvotes   int
	Content   template.HTML
	Body      string
	Replies   []Reply
}

type Reply struct {
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
