package main

import (
	"errors"
	"fmt"
	"github.com/IljaN/w2d/deepl"
	"github.com/IljaN/w2d/wikipedia"
	"github.com/alexflint/go-arg"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
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

// newTranslateCmd returns cmd-function which fetches an article from wikipedia, parses to markdown and translates it using DeepL
func newTranslateCmd(parser *wikipedia.ArticleParser, deepl deepl.Client) func(articleHTML io.ReadCloser, tgtLang, srcLang string) (string, error) {
	return func(articleHTML io.ReadCloser, tgtLang, srcLang string) (string, error) {
		markdown, err := parser.Parse(articleHTML)
		if err != nil {
			return "", fmt.Errorf("failed to parse: %s", err)
		}

		translated, err := deepl.TranslateToString(markdown, tgtLang, srcLang)
		if err != nil {
			return "", fmt.Errorf("failed to translate article: %s", err)
		}

		return translated, nil
	}
}

type markdownArgs struct {
	Article string `arg:"positional" default:"" help:"full url to the article or '-' for STDIN"`
}

// newMarkdownCmd returns cmd-function witch fetches an article from wikipedia and converts it to markdown
func newMarkdownCmd(parser *wikipedia.ArticleParser) func(articleHTML io.ReadCloser) (string, error) {
	return func(articleHTML io.ReadCloser) (string, error) {
		markdown, err := parser.Parse(articleHTML)
		if err != nil {
			return "", fmt.Errorf("failed to parse: %s", err)
		}

		return markdown, nil
	}
}

type listLanguagesArgs struct {
	Type string `arg:"-t,--" default:"source" help:"Which type of languages to return (source or target)"`
	authKey
}

// listLanguagesCmd retrieves cmd-function which gets languages supported by the DeepL
func newListLanguagesCmd(deepl deepl.Client) func(langType string) (string, error) {
	return func(langType string) (string, error) {
		if langType != "source" && langType != "target" {
			return "", fmt.Errorf("invalid target: %s\n", langType)
		}

		langs, err := deepl.GetSupportedLanguages(langType != "source")
		if err != nil {
			return "", err
		}

		keys := make([]string, 0, len(langs))
		for k := range langs {
			keys = append(keys, k)
		}

		sort.Strings(keys)

		res := strings.Builder{}
		for _, k := range keys {
			res.WriteString(fmt.Sprintf("%s - %s (formality_support: %t)\n", k, langs[k].Name, langs[k].SupportsFormality))
		}

		return res.String(), nil
	}
}

// w2d - translates wikipedia articles using DeepL api and renders them to markdown.
func main() {
	var out, cmdName string
	var err error
	args := rootArgs{}
	p := arg.MustParse(&args)

	switch {
	case args.Translate != nil:
		var articleHTML io.ReadCloser
		cmdName = "translate"
		translate := newTranslateCmd(wikipedia.NewArticleParser(), deepl.NewClient(args.Translate.DeeplAuthKey))
		articleHTML, err = openArticle(args.Translate.Article)
		if err != nil {
			break
		}

		out, err = translate(articleHTML, args.Translate.TargetLang, args.Translate.SourceLang)
	case args.Markdown != nil:
		var articleHTML io.ReadCloser
		cmdName = "markdown"
		markdown := newMarkdownCmd(wikipedia.NewArticleParser())
		articleHTML, err = openArticle(args.Markdown.Article)
		if err != nil {
			break
		}

		out, err = markdown(articleHTML)
	case args.ListLanguages != nil:
		cmdName = "list-languages"
		listLanguages := newListLanguagesCmd(deepl.NewClient(args.ListLanguages.DeeplAuthKey))
		out, err = listLanguages(args.ListLanguages.Type)
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
	if src == "-" {
		if stdInAttached() {
			return os.Stdin, nil
		}

		return nil, errors.New("stdin redirection required if '-' is given")
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

func stdInAttached() bool {
	stat, _ := os.Stdin.Stat()
	return (stat.Mode() & os.ModeCharDevice) == 0
}
