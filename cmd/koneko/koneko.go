// Example invocation:
// koneko -source hello_world.md -out /tmp/koneko
//
// koneko -source hello_world.md,goodbye_moon.md -out /tmp/koneko
//
// koneko -source hello_world.md -source goodbye_moon.md -out /tmp/koneko
package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"

	//. "github.com/cvanloo/blog-go/assert"
	"github.com/cvanloo/blog-go/config"
	"github.com/cvanloo/blog-go/markup"
	"github.com/cvanloo/blog-go/page"
)

type ArrayFlag []string

func (af *ArrayFlag) String() string {
	return fmt.Sprintf("%v", *af)
}

func (af *ArrayFlag) Set(value string) error {
	values := strings.FieldsFunc(value, func(r rune) bool {
		return r == ',' || r == ' '
	})
	*af = append(*af, values...)
	return nil
}

var (
	source ArrayFlag
)

type SiteConfig struct {
	Address        string `cfg:"mandatory=true"`
	Sitename       string `cfg:"mandatory=true"`
	Birthday       string `cfg:"mandatory=true"`
	DefaultTagline string `cfg:"mandatory=true;name=DEFAULT_TAGLINE"`
	Relme          string
	Fedicreator    string
	Author         string `cfg:"mandatory=true"`
	Email          string
	Lang           string `cfg:"default=en"`
	Extensions     string `cfg:"default=.md,.ᗢ"`
}

func main() {
	os.Exit(app())
}

func app() int {
	fmt.Println("こんにちは、子猫ちゃん")
	switch {
	case len(os.Args) >= 2 && os.Args[1] == "make-assets":
		os.Setenv("MAKE_ASSETS", "1")
		argSet := flag.NewFlagSet("make-assets", flag.ExitOnError)
		argSet.Var(&source, "source", "Input files. If given a directory, it will be processed recursively. A hyphen (the default) will read from stdin.")
		out := argSet.String("out", ".", "Directory to write static sites to.")
		envPath := argSet.String("env", ".env", "Path to the environment file.")
		argSet.Parse(os.Args[2:])
		if err := godotenv.Load(*envPath); err != nil {
			log.Println(err)
			return -1
		}
		var cfg SiteConfig
		if err := config.Load(&cfg); err != nil {
			log.Println(err)
			return -1
		}
		siteInfo, err := initializeSite(cfg)
		if err != nil {
			log.Println(err)
			return -1
		}
		m := markup.New(
			markup.SiteInfo(siteInfo),
			markup.IncludeExtensions(strings.Split(cfg.Extensions, ",")...),
			markup.SourcePaths(source),
			markup.OutDir(*out),
		)
		if err := m.MakeAssets(); err != nil {
			fmt.Println(err)
			return 1
		}
	default:
		argSet := flag.NewFlagSet("generate-blog", flag.ExitOnError)
		argSet.Var(&source, "source", "Input files. If given a directory, it will be processed recursively. A hyphen (the default) will read from stdin.")
		out := argSet.String("out", ".", "Directory to write static sites to.")
		envPath := argSet.String("env", ".env", "Path to the environment file.")
		argSet.Parse(os.Args[1:])
		if err := godotenv.Load(*envPath); err != nil {
			log.Println(err)
			return -1
		}
		var cfg SiteConfig
		if err := config.Load(&cfg); err != nil {
			log.Println(err)
			return -1
		}
		siteInfo, err := initializeSite(cfg)
		if err != nil {
			log.Println(err)
			return -1
		}
		fi, err := os.Stat(*out)
		if err != nil {
			log.Println(err)
			return -1
		}
		if !fi.IsDir() {
			log.Printf("%s is not a directory", *out)
			return -1
		}
		m := markup.New(
			markup.SiteInfo(siteInfo),
			markup.IncludeExtensions(strings.Split(cfg.Extensions, ",")...),
			markup.SourcePaths(source),
			markup.OutDir(*out),
		)
		if err := m.Run(); err != nil {
			fmt.Println(err)
			return 1
		}
	}
	return 0
}

func initializeSite(cfg SiteConfig) (siteInfo page.Site, err error) {
	addr, err := url.Parse(cfg.Address)
	if err != nil {
		return siteInfo, err
	}
	if addr.Scheme != "http" && addr.Scheme != "https" {
		return siteInfo, fmt.Errorf("invalid scheme for site address: %s", addr)
	}
	if addr.Path != "" && addr.Path != "/" {
		return siteInfo, fmt.Errorf("@todo: sites not rooted at the base address are not supported right now: %s", addr)
	}
	siteInfo.Address = addr

	siteInfo.Name = cfg.Sitename

	taglineParts := strings.FieldsFunc(cfg.DefaultTagline, func(r rune) bool {
		return r == '<' || r == '>'
	})
	var tagline page.StringOnlyContent
	modeWeird := false
	for _, part := range taglineParts {
		if modeWeird {
			tagline = append(tagline, page.Weird(part))
		} else {
			tagline = append(tagline, page.Text(part))
		}
		modeWeird = !modeWeird
	}
	siteInfo.DefaultTagline = tagline

	siteInfo.RelMe = cfg.Relme
	siteInfo.FediCreator = cfg.Fedicreator

	// @todo: actually run lexer and parser on this (requires some major changes to both the lexer and parser though)
	//        ...or we just add extra functions to lex and parse individual text
	siteInfo.Owner = page.EscapedString(strings.ReplaceAll(cfg.Author, "~", "&nbsp;"))

	siteInfo.Email = cfg.Email
	siteInfo.Birthday, err = time.Parse("2006", cfg.Birthday)
	if err != nil {
		return siteInfo, fmt.Errorf("invalid birthyear: %w", err)
	}

	// @todo: cfg.Lang
	return siteInfo, nil
}
