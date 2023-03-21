package topi

import (
	"context"
	"html/template"
	"time"

	"github.com/shurcooL/githubv4"
)

func (s *Server) FetchCategoryID(ctx context.Context) error {
	var query struct {
		Repository struct {
			DiscussionCategory struct {
				ID githubv4.ID
			} `graphql:"discussionCategory(slug: $category)"`
		} `graphql:"repository(name: $name, owner: $owner)"`
	}
	variables := map[string]interface{}{
		"owner":    githubv4.String(s.cfg.Blog.User),
		"name":     githubv4.String(s.cfg.Blog.Repository),
		"category": githubv4.String(s.cfg.Blog.Category),
	}
	if err := s.client.Query(ctx, &query, variables); err != nil {
		return err
	}

	s.categoryID = query.Repository.DiscussionCategory.ID
	return nil
}

func (s *Server) GetData(ctx context.Context) (*Blog, *GitHub, error) {
	var query struct {
		User struct {
			Login        string
			AvatarURL    string
			Repositories struct {
				Nodes []struct {
					Name        string
					URL         string
					Description string
					Languages   struct {
						Nodes []struct {
							Name  string
							Color string
						}
					} `graphql:"languages(first: $languages)"`
				}
			} `graphql:"repositories(first: $repositories)"`
		} `graphql:"user(login: $user)"`
		Repository struct {
			Discussions struct {
				Nodes []struct {
					URL       string
					Title     string
					CreatedAt time.Time
					Body      string
					Comments  struct {
						TotalCount int
						Nodes      []struct {
							Author struct {
								Login     string
								AvatarURL string
							}
							URL       string
							CreatedAt time.Time
							Body      string
						}
					} `graphql:"comments(first: $comments)"`
				}
			} `graphql:"discussions(first: $discussions, categoryId: $category)"`
		} `graphql:"repository(owner: $user, name: $name)"`
	}
	variables := map[string]interface{}{
		"user":         githubv4.String(s.cfg.Blog.User),
		"name":         githubv4.String(s.cfg.Blog.Repository),
		"category":     s.categoryID,
		"repositories": githubv4.Int(10),
		"languages":    githubv4.Int(10),
		"discussions":  githubv4.Int(10),
		"comments":     githubv4.Int(10),
	}
	if err := s.client.Query(ctx, &query, variables); err != nil {
		return nil, nil, err
	}

	blog := Blog{
		Posts: make([]Post, 0, len(query.Repository.Discussions.Nodes)),
	}
	for _, node := range query.Repository.Discussions.Nodes {
		comments := make([]Comment, 0, len(node.Comments.Nodes))

		for _, cNode := range node.Comments.Nodes {
			comments = append(comments, Comment{
				Author:    cNode.Author.Login,
				AvatarURL: cNode.Author.AvatarURL,
				CreatedAt: cNode.CreatedAt,
				Body:      cNode.Body,
			})
		}

		blog.Posts = append(blog.Posts, Post{
			Title:     node.Title,
			CreatedAt: node.CreatedAt,
			URL:       template.URL(node.URL),
			Body:      node.Body,
			Comments:  comments,
		})
	}

	github := GitHub{
		User: User{
			Name:      query.User.Login,
			AvatarURL: query.User.AvatarURL,
		},
		Projects: make([]Project, 0, len(query.User.Repositories.Nodes)),
	}
	for _, node := range query.User.Repositories.Nodes {
		languages := make([]Language, 0, len(node.Languages.Nodes))
		for _, lNode := range node.Languages.Nodes {
			languages = append(languages, Language{
				Name:  lNode.Name,
				Color: lNode.Color,
			})
		}
		github.Projects = append(github.Projects, Project{
			Name:        node.Name,
			Description: node.Description,
			URL:         node.URL,
			Languages:   languages,
		})
	}

	return &blog, &github, nil
}
