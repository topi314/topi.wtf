package main

import (
	"context"
	"embed"
	"flag"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	chtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/dustin/go-humanize"
	"github.com/mitchellh/mapstructure"
	"github.com/shurcooL/githubv4"
	"github.com/spf13/viper"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark-emoji"
	"github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
	"go.abhg.dev/goldmark/anchor"
	"golang.org/x/oauth2"

	"github.com/topisenpai/topi.wtf/topi"
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
	log.Printf("Starting topi.wtf with version: %s (commit: %s, build time: %s)...", version, commit, buildTime)
	cfgPath := flag.String("config", "", "path to config.json")
	flag.Parse()

	if *cfgPath != "" {
		viper.SetConfigFile(*cfgPath)
	} else {
		viper.SetConfigName("config")
		viper.SetConfigType("json")
		viper.AddConfigPath(".")
		viper.AddConfigPath("/etc/topi.wtf/")
	}
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalln("Error while reading config:", err)
	}
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetEnvPrefix("topi_wtf")
	viper.AutomaticEnv()

	var cfg topi.Config
	if err := viper.Unmarshal(&cfg, func(config *mapstructure.DecoderConfig) {
		config.TagName = "cfg"
	}); err != nil {
		log.Fatalln("Error while unmarshalling config:", err)
	}
	log.Println("Config2:", cfg)

	if cfg.Debug {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	}

	var (
		tmplFunc topi.ExecuteTemplateFunc
		assets   http.FileSystem
	)

	funcs := template.FuncMap{
		"humanizeTime": humanize.Time,
	}

	if cfg.DevMode {
		log.Println("Development mode enabled")
		tmplFunc = func(wr io.Writer, name string, data any) error {
			tmpl := template.New("").Funcs(funcs)
			tmpl = template.Must(tmpl.ParseGlob("templates/*.gohtml"))
			tmpl = template.Must(tmpl.ParseGlob("templates/*/*.gohtml"))
			return tmpl.ExecuteTemplate(wr, name, data)
		}
		assets = http.Dir(".")
	} else {
		tmpl := template.New("").Funcs(funcs)
		tmpl = template.Must(tmpl.ParseFS(Templates, "templates/*.gohtml"))
		tmpl = template.Must(tmpl.ParseFS(Templates, "templates/*/*.gohtml"))
		tmplFunc = tmpl.ExecuteTemplate
		assets = http.FS(Assets)
	}

	httpClient := oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: cfg.GitHub.AccessToken},
	))
	client := githubv4.NewClient(httpClient)

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

	s := topi.NewServer(topi.FormatBuildVersion(version, commit, buildTime), cfg, client, md, assets, tmplFunc)

	if err := s.FetchCategoryID(context.TODO()); err != nil {
		log.Fatalln("Error while fetching category ID:", err)
	}
	log.Println("topi.wtf listening on:", cfg.ListenAddr)
	go s.Start()
	defer s.Close()

	si := make(chan os.Signal, 1)
	signal.Notify(si, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-si
}
