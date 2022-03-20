package main

import (
	"fmt"
	"github.com/IljaN/w2d/deepl"
	"github.com/IljaN/w2d/wikipedia"
	"github.com/alexflint/go-arg"
	"io"
	"net/http"
	"net/url"
	"os"
)

type Config struct {
	Translate *translateCmd `arg:"subcommand:translate" help:"translates a wikipedia article"`
	Markdown  *markdownCmd  `arg:"subcommand:markdown" help:"converts wikipedia article html to markdown"`
	ListLangs *listLangsCmd `arg:"subcommand:list-languages" help:"retrieve a list of supported languages"`
}

func (Config) Description() string {
	return "Converts a wikipedia article to markdown and translates it using the DeepL.com api.\n"
}

type translateCmd struct {
	Article    string `arg:"positional" default:"" help:"full url to the article which should be translated. This arg is ignored if the article HTML is provided via STDIN"`
	TargetLang string `arg:"-t,--" default:"en" help:"target language for translation"`
	SourceLang string `arg:"-s,--" default:"" help:"source language, leave empty for autodetect"`

	authKey
}

type markdownCmd struct {
	Article string `arg:"positional" default:"" help:"full url to the article which should be converted. This arg is ignored if the article HTML is provided via STDIN"`
}

type listLangsCmd struct {
	Type string `arg:"-t,--" default:"source" help:"Which type of languages to return (source or target)"`
	authKey
}

type authKey struct {
	DeeplAuthKey string `arg:"required,-k,--,env:W2D_DEEPL_AUTH_KEY"`
}

// w2d - translates wikipedia articles using DeepL api and renders them to markdown.
func main() {
	c := Config{}
	p := arg.MustParse(&c)

	switch {
	case c.Translate != nil:
		if !stdInAttached() && c.Translate.Article == "" {
			err := p.FailSubcommand("article missing: Please provide the html via STDIN or pass the article URL as argument.", "translate")
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}
		translate(c.Translate)
	case c.Markdown != nil:
		if !stdInAttached() && c.Markdown.Article == "" {
			err := p.FailSubcommand("article missing: Please provide the html via STDIN or pass the article URL as argument.", "markdown")
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}

		convertToMarkdown(c.Markdown)
	case c.ListLangs != nil:
		listLanguages(c.ListLangs)
	}

	os.Exit(0)
}

// translate fetches an article from wikipedia, parses to markdown and translates it using DeepL
func translate(cfg *translateCmd) {
	html, err := openArticle(cfg.Article)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	markdown, err := wikipedia.NewArticleParser().Parse(html)
	if err != nil {
		fmt.Printf("failed to parse: %s", err)
		os.Exit(2)
	}

	translated, err := deepl.NewClient(cfg.DeeplAuthKey).TranslateToString(markdown, cfg.TargetLang, cfg.SourceLang)
	if err != nil {
		fmt.Printf("failed to translateArticle: %s", err)
		os.Exit(2)
	}

	fmt.Print(translated)
	os.Exit(0)
}

func convertToMarkdown(cfg *markdownCmd) {
	html, err := openArticle(cfg.Article)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	markdown, err := wikipedia.NewArticleParser().Parse(html)
	if err != nil {
		fmt.Printf("failed to parse: %s", err)
		os.Exit(2)
	}

	fmt.Print(markdown)
}

// listLanguages retrieves languages supported by DeepL
func listLanguages(cfg *listLangsCmd) {
	dc := deepl.NewClient(cfg.DeeplAuthKey)
	if cfg.Type != "source" && cfg.Type != "target" {
		fmt.Printf("invalid target: %s\n", cfg.Type)
		os.Exit(1)
	}

	langs, err := dc.GetSupportedLanguages(cfg.Type != "source")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	for lc := range langs {
		fmt.Printf("%s - %s (formality_support: %t)\n", lc, langs[lc].Name, langs[lc].SupportsFormality)
	}

	os.Exit(0)
}

// openArticle returns a reader for an article at srcUrl. If STDIN is attached srcUrl is ignored.
func openArticle(srcURL string) (io.ReadCloser, error) {
	if stdInAttached() {
		return os.Stdin, nil
	}

	u, err := url.ParseRequestURI(srcURL)
	if err != nil {
		return nil, err
	}

	resp, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

func stdInAttached() bool {
	stat, _ := os.Stdin.Stat()
	return (stat.Mode() & os.ModeCharDevice) == 0
}
