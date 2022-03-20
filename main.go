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

type rootArgs struct {
	Translate *translateCmd `arg:"subcommand:translate" help:"translates a wikipedia article"`
	Markdown  *markdownCmd  `arg:"subcommand:markdown" help:"converts wikipedia article html to markdown"`
	ListLangs *listLangsCmd `arg:"subcommand:list-languages" help:"retrieve a list of supported languages"`
}

func (rootArgs) Description() string {
	return "Converts a wikipedia article to markdown and translates it using the DeepL.com api.\n"
}

type authKey struct {
	DeeplAuthKey string `arg:"required,-k,--,env:W2D_DEEPL_AUTH_KEY"`
}

type translateCmd struct {
	TargetLang string `arg:"positional,required" help:"target language for translation"`
	Article    string `arg:"positional" default:"" help:"full url to the article which should be translated. This arg is ignored if the article HTML is provided via STDIN"`
	SourceLang string `arg:"-s,--" default:"" help:"source language, leave empty for autodetect"`

	authKey
}

// translate fetches an article from wikipedia, parses to markdown and translates it using DeepL
func (c *translateCmd) run() {
	html, err := openArticle(c.Article)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	markdown, err := wikipedia.NewArticleParser().Parse(html)
	if err != nil {
		fmt.Printf("failed to parse: %s", err)
		os.Exit(2)
	}

	translated, err := deepl.NewClient(c.DeeplAuthKey).TranslateToString(markdown, c.TargetLang, c.SourceLang)
	if err != nil {
		fmt.Printf("failed to translateArticle: %s", err)
		os.Exit(2)
	}

	fmt.Print(translated)
	os.Exit(0)
}

type markdownCmd struct {
	Article string `arg:"positional" default:"" help:"full url to the article which should be converted. This arg is ignored if the article HTML is provided via STDIN"`
}

func (c *markdownCmd) run() {
	html, err := openArticle(c.Article)
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

type listLangsCmd struct {
	Type string `arg:"-t,--" default:"source" help:"Which type of languages to return (source or target)"`
	authKey
}

// listLanguages retrieves languages supported by DeepL
func (c *listLangsCmd) run() {
	dc := deepl.NewClient(c.DeeplAuthKey)
	if c.Type != "source" && c.Type != "target" {
		fmt.Printf("invalid target: %s\n", c.Type)
		os.Exit(1)
	}

	langs, err := dc.GetSupportedLanguages(c.Type != "source")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	for lc := range langs {
		fmt.Printf("%s - %s (formality_support: %t)\n", lc, langs[lc].Name, langs[lc].SupportsFormality)
	}

	os.Exit(0)
}

// w2d - translates wikipedia articles using DeepL api and renders them to markdown.
func main() {
	cmd := rootArgs{}
	p := arg.MustParse(&cmd)

	switch {
	case cmd.Translate != nil:
		if !stdInAttached() && cmd.Translate.Article == "" {
			err := p.FailSubcommand("article missing: Please provide the html via STDIN or pass the article URL as argument.", "translate")
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}
		cmd.Translate.run()
	case cmd.Markdown != nil:
		if !stdInAttached() && cmd.Markdown.Article == "" {
			err := p.FailSubcommand("article missing: Please provide the html via STDIN or pass the article URL as argument.", "markdown")
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}

		cmd.Markdown.run()
	case cmd.ListLangs != nil:
		cmd.ListLangs.run()
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
