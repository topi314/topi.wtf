package topi

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/shurcooL/githubv4"
	"github.com/yuin/goldmark"
)

type ExecuteTemplateFunc func(wr io.Writer, name string, data any) error

func NewServer(version string, cfg Config, httpClient *http.Client, githubClient *githubv4.Client, md goldmark.Markdown, assets http.FileSystem, tmpl ExecuteTemplateFunc) *Server {
	s := &Server{
		version:      version,
		cfg:          cfg,
		httpClient:   httpClient,
		githubClient: githubClient,
		md:           md,
		assets:       assets,
		tmpl:         tmpl,
	}

	s.server = &http.Server{
		Addr:    cfg.ListenAddr,
		Handler: s.Routes(),
	}

	return s
}

type Server struct {
	version      string
	cfg          Config
	httpClient   *http.Client
	githubClient *githubv4.Client
	server       *http.Server
	md           goldmark.Markdown
	assets       http.FileSystem
	tmpl         ExecuteTemplateFunc
}

func (s *Server) Start() {
	if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("Error while listening", slog.Any("err", err))
		os.Exit(-1)
	}
}

func (s *Server) Close() {
	if err := s.server.Close(); err != nil {
		slog.Error("Error while closing server", slog.Any("err", err))
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
