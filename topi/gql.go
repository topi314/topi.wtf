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

func (s *Server) FetchData(ctx context.Context) (*Variables, error) {
	var query struct {
		User struct {
			Login      string
			AvatarURL  string
			Repository struct {
				Description string
				Object      struct {
					Tree struct {
						Entries []struct {
							Name   string
							Object struct {
								Blob struct {
									Text string
								} `graphql:"... on Blob"`
							}
						}
					} `graphql:"... on Tree"`
				} `graphql:"object(expression: $expression)"`
			} `graphql:"repository(name: $user)"`
			Repositories Repositories `graphql:"repositories(first: $repositories, isFork: false, privacy: PUBLIC, orderBy: {field: PUSHED_AT, direction: DESC})"`
		} `graphql:"user(login: $user)"`
	}
	variables := map[string]interface{}{
		"user":         githubv4.String(s.cfg.GitHub.User),
		"repositories": githubv4.Int(10),
		"topics":       githubv4.Int(10),
		"expression":   githubv4.String("HEAD:"),
	}
	if err := s.githubClient.Query(ctx, &query, variables); err != nil {
		return nil, err
	}

	user := User{
		Name:      query.User.Login,
		AvatarURL: template.URL(query.User.AvatarURL),
	}

	var home Home
	for _, entry := range query.User.Repository.Object.Tree.Entries {
		if entry.Name == "README.md" {
			home.Body = entry.Object.Blob.Text
		}
	}

	projectsAfter := query.User.Repositories.PageInfo.EndCursor
	if !query.User.Repositories.PageInfo.HasNextPage {
		projectsAfter = ""
	}

	return &Variables{
		User:          user,
		Home:          home,
		Projects:      parseRepositories(query.User.Repositories),
		ProjectsAfter: projectsAfter,
		Dark:          true,
		Description:   query.User.Repository.Description,
	}, nil
}

func (s *Server) FetchRepositories(ctx context.Context, after string) (*Variables, error) {
	var query struct {
		User struct {
			Repositories Repositories `graphql:"repositories(after: $after, first: $repositories, isFork: false, privacy: PUBLIC, orderBy: {field: PUSHED_AT, direction: DESC})"`
		} `graphql:"user(login: $user)"`
	}
	variables := map[string]interface{}{
		"user":         githubv4.String(s.cfg.GitHub.User),
		"repositories": githubv4.Int(10),
		"topics":       githubv4.Int(10),
		"after":        githubv4.String(after),
	}
	if err := s.githubClient.Query(ctx, &query, variables); err != nil {
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
