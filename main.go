package main

import (
	"context"
	"embed"
	"flag"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	chtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/dustin/go-humanize"
	"github.com/mattn/go-colorable"
	"github.com/shurcooL/githubv4"
	"github.com/topi314/tint"
	"github.com/topi314/topi.wtf/topi"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark-emoji"
	"github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
	"go.abhg.dev/goldmark/anchor"
	"golang.org/x/oauth2"
)

var (
	version   = "unknown"
	commit    = "unknown"
	buildTime = "unknown"
)

var (
	//go:embed templates/**
	Templates embed.FS

	//go:embed assets
	Assets embed.FS
)

func main() {
	cfgPath := flag.String("config", "config.yml", "path to config file")
	flag.Parse()

	cfg, err := topi.LoadConfig(*cfgPath)
	if err != nil {
		slog.Error("failed to load config", slog.Any("error", err))
		os.Exit(-1)
	}
	setupLogger(cfg.Log)

	slog.Info("Starting topi.wtf...", slog.Any("config", cfg), slog.Any("version", version), slog.Any("commit", commit), slog.Any("buildTime", buildTime))

	var (
		tmplFunc topi.ExecuteTemplateFunc
		assets   http.FileSystem
	)

	funcs := template.FuncMap{
		"humanizeTime": humanize.Time,
	}

	if cfg.DevMode {
		slog.Info("running in dev mode")
		tmplFunc = func(wr io.Writer, name string, data any) error {
			tmpl, err := template.New("").Funcs(funcs).ParseGlob("templates/*.gohtml")
			if err != nil {
				return err
			}
			return tmpl.ExecuteTemplate(wr, name, data)
		}
		assets = http.Dir(".")
	} else {
		tmpl, err := template.New("").Funcs(funcs).ParseFS(Templates, "templates/*.gohtml")
		if err != nil {
			slog.Error("failed to parse templates", slog.Any("error", err))
			os.Exit(-1)
		}
		tmplFunc = tmpl.ExecuteTemplate
		assets = http.FS(Assets)
	}

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}
	githubClient := githubv4.NewClient(oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: cfg.GitHub.AccessToken},
	)))

	md := goldmark.New(
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
		goldmark.WithExtensions(
			extension.GFM,
			emoji.Emoji,
			&anchor.Extender{},
			highlighting.NewHighlighting(
				highlighting.WithCustomStyle(styles.Get("swapoff")),
				highlighting.WithFormatOptions(
					chtml.TabWidth(4),
					chtml.WithClasses(true),
					chtml.ClassPrefix("ch-"),
				),
			),
		),
	)

	s := topi.NewServer(topi.FormatBuildVersion(version, commit, buildTime), cfg, httpClient, githubClient, md, assets, tmplFunc)
	go s.Start()
	defer s.Close()

	slog.Info("started topi.wtf", slog.Any("listen_addr", cfg.ListenAddr))
	si := make(chan os.Signal, 1)
	signal.Notify(si, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-si
}

const (
	ansiFaint         = "\033[2m"
	ansiWhiteBold     = "\033[37;1m"
	ansiYellowBold    = "\033[33;1m"
	ansiCyanBold      = "\033[36;1m"
	ansiCyanBoldFaint = "\033[36;1;2m"
	ansiRedFaint      = "\033[31;2m"
	ansiRedBold       = "\033[31;1m"

	ansiRed     = "\033[31m"
	ansiYellow  = "\033[33m"
	ansiGreen   = "\033[32m"
	ansiMagenta = "\033[35m"
)

func setupLogger(cfg topi.LogConfig) {
	var handler slog.Handler
	switch cfg.Format {
	case "json":
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			AddSource: cfg.AddSource,
			Level:     cfg.Level,
		})

	case "text":
		handler = tint.NewHandler(colorable.NewColorable(os.Stdout), &tint.Options{
			AddSource: cfg.AddSource,
			Level:     cfg.Level,
			NoColor:   cfg.NoColor,
			LevelColors: map[slog.Level]string{
				slog.LevelDebug: ansiMagenta,
				slog.LevelInfo:  ansiGreen,
				slog.LevelWarn:  ansiYellow,
				slog.LevelError: ansiRed,
			},
			Colors: map[tint.Kind]string{
				tint.KindTime:            ansiYellowBold,
				tint.KindSourceFile:      ansiCyanBold,
				tint.KindSourceSeparator: ansiCyanBoldFaint,
				tint.KindSourceLine:      ansiCyanBold,
				tint.KindMessage:         ansiWhiteBold,
				tint.KindKey:             ansiFaint,
				tint.KindSeparator:       ansiFaint,
				tint.KindValue:           ansiWhiteBold,
				tint.KindErrorKey:        ansiRedFaint,
				tint.KindErrorSeparator:  ansiFaint,
				tint.KindErrorValue:      ansiRedBold,
			},
		})
	default:
		slog.Error("Unknown log format", slog.String("format", cfg.Format))
		os.Exit(-1)
	}
	slog.SetDefault(slog.New(handler))
}
