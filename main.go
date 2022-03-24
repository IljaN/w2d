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
	"strings"
)

type rootArgs struct {
	Translate     *translateCmd     `arg:"subcommand:translate" help:"translates a wikipedia article"`
	Markdown      *markdownCmd      `arg:"subcommand:markdown" help:"converts wikipedia article html to markdown"`
	ListLanguages *listLanguagesCmd `arg:"subcommand:list-languages" help:"retrieve a list of supported languages"`
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
func (c *translateCmd) run() (string, error) {
	html, err := openArticle(c.Article)
	if err != nil {
		return "", err
	}

	markdown, err := wikipedia.NewArticleParser().Parse(html)
	if err != nil {
		return "", fmt.Errorf("failed to parse: %s", err)
	}

	translated, err := deepl.NewClient(c.DeeplAuthKey).TranslateToString(markdown, c.TargetLang, c.SourceLang)
	if err != nil {
		return "", fmt.Errorf("failed to translate article: %s", err)
	}

	return translated, nil
}

type markdownCmd struct {
	Article string `arg:"positional" default:"" help:"full url to the article which should be converted. This arg is ignored if the article HTML is provided via STDIN"`
}

func (c *markdownCmd) run() (string, error) {
	html, err := openArticle(c.Article)
	if err != nil {
		return "", err
	}

	markdown, err := wikipedia.NewArticleParser().Parse(html)
	if err != nil {
		return "", fmt.Errorf("failed to parse: %s", err)
	}

	return markdown, nil
}

type listLanguagesCmd struct {
	Type string `arg:"-t,--" default:"source" help:"Which type of languages to return (source or target)"`
	authKey
}

// listLanguages retrieves languages supported by DeepL
func (c *listLanguagesCmd) run() (string, error) {
	dc := deepl.NewClient(c.DeeplAuthKey)
	if c.Type != "source" && c.Type != "target" {
		return "", fmt.Errorf("invalid target: %s\n", c.Type)
	}

	langs, err := dc.GetSupportedLanguages(c.Type != "source")
	if err != nil {
		return "", err
	}

	res := strings.Builder{}
	for lc := range langs {
		res.WriteString(fmt.Sprintf("%s - %s (formality_support: %t)\n", lc, langs[lc].Name, langs[lc].SupportsFormality))
	}

	return res.String(), nil

}

// w2d - translates wikipedia articles using DeepL api and renders them to markdown.
func main() {
	cmd := rootArgs{}
	p := arg.MustParse(&cmd)
	var out string
	var err error

	switch {
	case cmd.Translate != nil:
		if !stdInAttached() && cmd.Translate.Article == "" {
			err = p.FailSubcommand("article missing: Please provide the html via STDIN or pass the article URL as argument.", "translate")
		}
		out, err = cmd.Translate.run()
	case cmd.Markdown != nil:
		if !stdInAttached() && cmd.Markdown.Article == "" {
			err = p.FailSubcommand("article missing: Please provide the html via STDIN or pass the article URL as argument.", "markdown")
		}

		out, err = cmd.Markdown.run()
	case cmd.ListLanguages != nil:
		out, err = cmd.ListLanguages.run()
	}

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Print(out)
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
