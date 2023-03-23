package topi

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"runtime"
	"time"

	"github.com/shurcooL/githubv4"
	"github.com/yuin/goldmark"
)

type ExecuteTemplateFunc func(wr io.Writer, name string, data any) error

func NewServer(version string, cfg Config, client *githubv4.Client, md goldmark.Markdown, assets http.FileSystem, tmpl ExecuteTemplateFunc) *Server {
	s := &Server{
		version: version,
		cfg:     cfg,
		client:  client,
		md:      md,
		assets:  assets,
		tmpl:    tmpl,
	}

	s.server = &http.Server{
		Addr:    cfg.ListenAddr,
		Handler: s.Routes(),
	}

	return s
}

type Server struct {
	version    string
	cfg        Config
	categoryID githubv4.ID
	client     *githubv4.Client
	server     *http.Server
	md         goldmark.Markdown
	assets     http.FileSystem
	tmpl       ExecuteTemplateFunc
}

func (s *Server) Start() {
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalln("Error while listening:", err)
	}
}

func (s *Server) Close() {
	if err := s.server.Close(); err != nil {
		log.Println("Error while closing server:", err)
	}
}

func FormatBuildVersion(version string, commit string, buildTime string) string {
	if len(commit) > 7 {
		commit = commit[:7]
	}

	buildTimeStr := "unknown"
	if buildTime != "unknown" {
		parsedTime, _ := time.Parse(time.RFC3339, buildTime)
		if !parsedTime.IsZero() {
			buildTimeStr = parsedTime.Format(time.ANSIC)
		}
	}
	return fmt.Sprintf("Go Version: %s\nVersion: %s\nCommit: %s\nBuild Time: %s\nOS/Arch: %s/%s\n", runtime.Version(), version, commit, buildTimeStr, runtime.GOOS, runtime.GOARCH)
}
