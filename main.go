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
	Translate     *translateArgs     `arg:"subcommand:translate" help:"translates a wikipedia article"`
	Markdown      *markdownArgs      `arg:"subcommand:markdown" help:"converts wikipedia article html to markdown"`
	ListLanguages *listLanguagesArgs `arg:"subcommand:list-languages" help:"retrieve a list of supported languages"`
}

func (rootArgs) Description() string {
	return "Converts a wikipedia article to markdown and translates it using the DeepL.com api.\n"
}

type authKey struct {
	DeeplAuthKey string `arg:"required,-k,--,env:W2D_DEEPL_AUTH_KEY"`
}

type translateArgs struct {
	TargetLang string `arg:"positional,required" help:"target language for translation"`
	Article    string `arg:"positional,required" help:"full url to the article or '-' for STDIN"`
	SourceLang string `arg:"-s,--" default:"" help:"source language, leave empty for autodetect"`

	authKey
}

// translateCmd fetches an article from wikipedia, parses to markdown and translates it using DeepL
func translateCmd(c *translateArgs) (string, error) {
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

type markdownArgs struct {
	Article string `arg:"positional" default:"" help:"full url to the article or '-' for STDIN"`
}

// markdownCmd fetches an article from wikipedia and converts it to markdown
func markdownCmd(c *markdownArgs) (string, error) {
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

type listLanguagesArgs struct {
	Type string `arg:"-t,--" default:"source" help:"Which type of languages to return (source or target)"`
	authKey
}

// listLanguagesCmd retrieves languages supported by the DeepL API
func listLanguagesCmd(c *listLanguagesArgs) (string, error) {
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
	var out string
	var cmdName string
	var err error
	args := rootArgs{}
	p := arg.MustParse(&args)

	switch {
	case args.Translate != nil:
		cmdName = "translate"
		out, err = translateCmd(args.Translate)
	case args.Markdown != nil:
		cmdName = "markdown"
		out, err = markdownCmd(args.Markdown)
	case args.ListLanguages != nil:
		cmdName = "list-languages"
		out, err = listLanguagesCmd(args.ListLanguages)
	}

	if err != nil {
		_ = p.FailSubcommand(err.Error(), cmdName)
		os.Exit(1)
	}

	fmt.Print(out)
	os.Exit(0)
}

// openArticle returns a reader for an article at srcUrl. If STDIN is attached srcUrl is ignored.
func openArticle(src string) (io.ReadCloser, error) {
	if stdInAttached() && src == "-" {
		return os.Stdin, nil
	}

	u, err := url.ParseRequestURI(src)
	if err != nil {
		return nil, err
	}

	resp, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

// stdInAttached returns true if stdin is connected
func stdInAttached() bool {
	stat, _ := os.Stdin.Stat()
	return (stat.Mode() & os.ModeCharDevice) == 0
}
