package topi

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"time"

	"github.com/shurcooL/githubv4"
)

type Repositories struct {
	Nodes []struct {
		Name             string
		URL              string
		Description      string
		StargazerCount   int
		ForkCount        int
		PushedAt         time.Time
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
		} `graphql:"languages(first: 1, orderBy: {field: SIZE, direction: DESC})"`
	}
	PageInfo struct {
		EndCursor   string
		HasNextPage bool
	}
}

type Discussions struct {
	Nodes []struct {
		URL         string
		Title       string
		CreatedAt   time.Time
		Body        string
		UpvoteCount int
		Comments    struct {
			TotalCount int
			Nodes      []struct {
				Author struct {
					Login     string
					AvatarURL string
				}
				URL         string
				CreatedAt   time.Time
				UpvoteCount int
				Body        string
				Replies     struct {
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
				} `graphql:"replies(first: $replies)"`
			}
		} `graphql:"comments(first: $comments)"`
	}
	PageInfo struct {
		EndCursor   string
		HasNextPage bool
	}
}

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

func parseRepositories(repositories Repositories) []Project {
	projects := make([]Project, 0, len(repositories.Nodes))
	for _, node := range repositories.Nodes {
		var language *Language
		if len(node.Languages.Nodes) > 0 {
			lNode := node.Languages.Nodes[0]
			language = &Language{
				Name:  lNode.Name,
				Color: lNode.Color,
			}
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
			UpdatedAt:   node.PushedAt,
			Language:    language,
			Topics:      topics,
		})
	}

	return projects
}

func parseDiscussions(discussions Discussions) []Post {
	posts := make([]Post, 0, len(discussions.Nodes))
	for _, node := range discussions.Nodes {
		comments := make([]Comment, 0, len(node.Comments.Nodes))

		for _, cNode := range node.Comments.Nodes {
			replies := make([]Reply, 0, len(cNode.Replies.Nodes))
			for _, rNode := range cNode.Replies.Nodes {
				replies = append(replies, Reply{
					Author:    rNode.Author.Login,
					AvatarURL: template.URL(rNode.Author.AvatarURL),
					CreatedAt: rNode.CreatedAt,
					Body:      rNode.Body,
				})
			}
			comments = append(comments, Comment{
				Author:    cNode.Author.Login,
				AvatarURL: template.URL(cNode.Author.AvatarURL),
				CreatedAt: cNode.CreatedAt,
				Upvotes:   cNode.UpvoteCount,
				Body:      cNode.Body,
				Replies:   replies,
			})
		}

		posts = append(posts, Post{
			Title:     node.Title,
			CreatedAt: node.CreatedAt,
			URL:       template.URL(node.URL),
			Body:      node.Body,
			Upvotes:   node.UpvoteCount,
			Comments:  comments,
		})
	}
	return posts
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
			Repositories Repositories `graphql:"repositories(first: $repositories, isFork: false, privacy: PUBLIC, orderBy: {field: PUSHED_AT, direction: DESC})"`
		} `graphql:"user(login: $user)"`
		Repository struct {
			Discussions Discussions `graphql:"discussions(first: $discussions, categoryId: $category)"`
		} `graphql:"repository(owner: $user, name: $name)"`
	}
	variables := map[string]interface{}{
		"user":         githubv4.String(s.cfg.Blog.User),
		"name":         githubv4.String(s.cfg.Blog.Repository),
		"category":     s.categoryID,
		"repositories": githubv4.Int(10),
		"topics":       githubv4.Int(10),
		"discussions":  githubv4.Int(10),
		"comments":     githubv4.Int(100),
		"replies":      githubv4.Int(100),
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

	postsAfter := query.Repository.Discussions.PageInfo.EndCursor
	if !query.Repository.Discussions.PageInfo.HasNextPage {
		postsAfter = ""
	}
	projectsAfter := query.User.Repositories.PageInfo.EndCursor
	if !query.User.Repositories.PageInfo.HasNextPage {
		projectsAfter = ""
	}

	return &Variables{
		User:          user,
		Home:          home,
		Posts:         parseDiscussions(query.Repository.Discussions),
		PostsAfter:    postsAfter,
		Projects:      parseRepositories(query.User.Repositories),
		ProjectsAfter: projectsAfter,
		Dark:          true,
	}, nil
}

func (s *Server) FetchRepositories(ctx context.Context, after string) (*Variables, error) {
	var query struct {
		User struct {
			Repositories Repositories `graphql:"repositories(after: $after, first: $repositories, isFork: false, privacy: PUBLIC, orderBy: {field: PUSHED_AT, direction: DESC})"`
		} `graphql:"user(login: $user)"`
	}
	variables := map[string]interface{}{
		"user":         githubv4.String(s.cfg.Blog.User),
		"repositories": githubv4.Int(10),
		"topics":       githubv4.Int(10),
		"after":        githubv4.String(after),
	}
	if err := s.client.Query(ctx, &query, variables); err != nil {
		return nil, err
	}

	after = query.User.Repositories.PageInfo.EndCursor
	if !query.User.Repositories.PageInfo.HasNextPage {
		after = ""
	}

	return &Variables{
		Projects:      parseRepositories(query.User.Repositories),
		ProjectsAfter: after,
	}, nil
}

func (s *Server) FetchPosts(ctx context.Context, after string) (*Variables, error) {
	var query struct {
		Repository struct {
			Discussions Discussions `graphql:"discussions(after: $after, first: $discussions, categoryId: $category)"`
		} `graphql:"repository(owner: $user, name: $name)"`
	}
	variables := map[string]interface{}{
		"user":        githubv4.String(s.cfg.Blog.User),
		"name":        githubv4.String(s.cfg.Blog.Repository),
		"category":    s.categoryID,
		"discussions": githubv4.Int(10),
		"comments":    githubv4.Int(10),
		"replies":     githubv4.Int(10),
		"after":       githubv4.String(after),
	}
	if err := s.client.Query(ctx, &query, variables); err != nil {
		return nil, err
	}

	after = query.Repository.Discussions.PageInfo.EndCursor
	if !query.Repository.Discussions.PageInfo.HasNextPage {
		after = ""
	}

	return &Variables{
		Posts:      parseDiscussions(query.Repository.Discussions),
		PostsAfter: after,
	}, nil
}

func (s *Server) HighlightData(vars *Variables) error {
	buff := new(bytes.Buffer)
	if err := s.md.Convert([]byte(vars.Home.Body), buff); err != nil {
		return fmt.Errorf("failed to format home: %w", err)
	}
	vars.Home.Content = template.HTML(buff.String())
	buff.Reset()

	for i, post := range vars.Posts {
		for ii, comment := range post.Comments {
			for iii, reply := range comment.Replies {
				if err := s.md.Convert([]byte(reply.Body), buff); err != nil {
					return fmt.Errorf("failed to format reply: %w", err)
				}
				vars.Posts[i].Comments[ii].Replies[iii].Content = template.HTML(buff.String())
				buff.Reset()
			}

			if err := s.md.Convert([]byte(comment.Body), buff); err != nil {
				return fmt.Errorf("failed to format comment: %w", err)
			}
			post.Comments[ii].Content = template.HTML(buff.String())
			buff.Reset()
		}

		if err := s.md.Convert([]byte(post.Body), buff); err != nil {
			return fmt.Errorf("failed to format post: %w", err)
		}
		vars.Posts[i].Content = template.HTML(buff.String())
		buff.Reset()
	}

	vars.CSS = template.CSS(buff.String())

	return nil
}
