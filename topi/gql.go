package topi

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"time"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
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

func (s *Server) FetchData(ctx context.Context) (*Variables, error) {
	var query struct {
		User struct {
			Login      string
			AvatarURL  string
			Repository struct {
				Object struct {
					Blob struct {
						Text string
					} `graphql:"... on Blob"`
				} `graphql:"object(expression: $expression)"`
			} `graphql:"repository(name: $user)"`
			Repositories struct {
				Nodes []struct {
					Name             string
					URL              string
					Description      string
					StargazerCount   int
					ForkCount        int
					RepositoryTopics struct {
						Nodes []struct {
							Topic struct {
								Name string
							}
							URL string
						}
					} `graphql:"repositoryTopics(first: $topics)"`
					Languages struct {
						Nodes []struct {
							Name  string
							Color string
						}
					} `graphql:"languages(first: $languages)"`
				}
			} `graphql:"repositories(first: $repositories, orderBy: {field: STARGAZERS, direction: DESC})"`
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
		"topics":       githubv4.Int(10),
		"discussions":  githubv4.Int(10),
		"comments":     githubv4.Int(10),
		"expression":   githubv4.String("HEAD:README.md"),
	}
	if err := s.client.Query(ctx, &query, variables); err != nil {
		return nil, err
	}

	user := User{
		Name:      query.User.Login,
		AvatarURL: template.URL(query.User.AvatarURL),
	}

	home := Home{
		Body: query.User.Repository.Object.Blob.Text,
	}

	posts := make([]Post, 0, len(query.Repository.Discussions.Nodes))
	for _, node := range query.Repository.Discussions.Nodes {
		comments := make([]Comment, 0, len(node.Comments.Nodes))

		for _, cNode := range node.Comments.Nodes {
			comments = append(comments, Comment{
				Author:    cNode.Author.Login,
				AvatarURL: template.URL(cNode.Author.AvatarURL),
				CreatedAt: cNode.CreatedAt,
				Body:      cNode.Body,
			})
		}

		posts = append(posts, Post{
			Title:     node.Title,
			CreatedAt: node.CreatedAt,
			URL:       template.URL(node.URL),
			Body:      node.Body,
			Comments:  comments,
		})
	}

	projects := make([]Project, 0, len(query.User.Repositories.Nodes))
	for _, node := range query.User.Repositories.Nodes {
		languages := make([]Language, 0, len(node.Languages.Nodes))
		for _, lNode := range node.Languages.Nodes {
			languages = append(languages, Language{
				Name:  lNode.Name,
				Color: lNode.Color,
			})
		}

		topics := make([]Topic, 0, len(node.RepositoryTopics.Nodes))
		for _, tNode := range node.RepositoryTopics.Nodes {
			topics = append(topics, Topic{
				Name: tNode.Topic.Name,
				URL:  tNode.URL,
			})
		}

		projects = append(projects, Project{
			Name:        node.Name,
			Description: node.Description,
			URL:         template.URL(node.URL),
			Stars:       node.StargazerCount,
			Forks:       node.ForkCount,
			Languages:   languages,
			Topics:      topics,
		})
	}

	return &Variables{
		User:     user,
		Home:     home,
		Posts:    posts,
		Projects: projects,
		Dark:     true,
	}, nil
}

func (s *Server) HighlightData(vars *Variables, style *chroma.Style) error {
	buff := new(bytes.Buffer)

	iterator, err := lexers.Markdown.Tokenise(nil, vars.Home.Body)
	if err != nil {
		return fmt.Errorf("failed to tokenize home: %w", err)
	}
	if err = s.htmlFormatter.Format(buff, style, iterator); err != nil {
		return fmt.Errorf("failed to format home: %w", err)
	}
	vars.Home.Content = template.HTML(buff.String())
	buff.Reset()

	for i, post := range vars.Posts {
		for ii, comment := range post.Comments {
			iterator, err = lexers.Markdown.Tokenise(nil, comment.Body)
			if err != nil {
				return fmt.Errorf("failed to tokenize comment: %w", err)
			}
			if err = s.htmlFormatter.Format(buff, style, iterator); err != nil {
				return fmt.Errorf("failed to format comment: %w", err)
			}
			post.Comments[ii].Content = template.HTML(buff.String())
			buff.Reset()
		}

		iterator, err = lexers.Markdown.Tokenise(nil, post.Body)
		if err != nil {
			return fmt.Errorf("failed to tokenize post: %w", err)
		}
		if err = s.htmlFormatter.Format(buff, style, iterator); err != nil {
			return fmt.Errorf("failed to format post: %w", err)
		}
		vars.Posts[i].Content = template.HTML(buff.String())
		buff.Reset()
	}

	if err = s.htmlFormatter.WriteCSS(buff, style); err != nil {
		return fmt.Errorf("failed to write css: %w", err)
	}
	vars.CSS = template.CSS(buff.String())

	return nil
}
