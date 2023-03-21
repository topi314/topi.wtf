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

	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/mitchellh/mapstructure"
	"github.com/shurcooL/githubv4"
	"github.com/spf13/viper"
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
	log.Println("Config:", cfg)

	if cfg.Debug {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	}

	var (
		tmplFunc topi.ExecuteTemplateFunc
		assets   http.FileSystem
	)
	if cfg.DevMode {
		log.Println("Development mode enabled")
		tmplFunc = func(wr io.Writer, name string, data any) error {
			tmpl := template.New("")
			tmpl = template.Must(tmpl.ParseGlob("templates/*.gohtml"))
			tmpl = template.Must(tmpl.ParseGlob("templates/*/*.gohtml"))
			return tmpl.ExecuteTemplate(wr, name, data)
		}
		assets = http.Dir(".")
	} else {
		tmpl := template.New("")
		tmpl = template.Must(tmpl.ParseFS(Templates, "templates/*.gohtml"))
		tmpl = template.Must(tmpl.ParseFS(Templates, "templates/*/*.gohtml"))
		tmplFunc = tmpl.ExecuteTemplate
		assets = http.FS(Assets)
	}

	httpClient := oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: cfg.GitHub.AccessToken},
	))
	client := githubv4.NewClient(httpClient)

	htmlFormatter := html.New(
		html.WithClasses(true),
		html.WithAllClasses(true),
		html.ClassPrefix("ch-"),
		html.Standalone(false),
		html.InlineCode(false),
		html.WithNopPreWrapper(),
		html.TabWidth(4),
	)

	s := topi.NewServer(topi.FormatBuildVersion(version, commit, buildTime), cfg, client, htmlFormatter, assets, tmplFunc)

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
