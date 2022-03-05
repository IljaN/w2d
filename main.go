package main

import (
	"fmt"
	"github.com/IljaN/w2d/deepl"
	"github.com/IljaN/w2d/wikipedia"
	"github.com/alexflint/go-arg"
	"io"
	"net/http"
	"os"
)

type Config struct {
	Translate    *translateCmd `arg:"subcommand:translate" help:"translates a wikipedia article"`
	ListLangs    *listLangsCmd `arg:"subcommand:list-languages" help:"retrieve a list of supported languages"`
	DeeplAuthKey string        `arg:"required,-k,--,env:W2D_DEEPL_AUTH_KEY"`
}

func (Config) Description() string {
	return "Converts a wikipedia article to markdown and translates it using the DeepL.com api.\n"
}

type translateCmd struct {
	Article    string `arg:"positional,required" help:"full url to the article which should be translated"`
	TargetLang string `arg:"-t,--" default:"en" help:"target language for translation"`
	SourceLang string `arg:"-s,--" default:"" help:"source language, leave empty for autodetect"`
	ParseOnly  bool   `arg:"-p,--" default:"false" help:"parse to markdown without translating"`
	authKey    string
}

type listLangsCmd struct {
	Type    string `arg:"-t,--" default:"source" help:"Which type of languages to return (source or target)"`
	authKey string
}

// w2d - translates wikipedia articles using DeepL api and renders them to markdown.
func main() {
	c := Config{}
	arg.MustParse(&c)
	switch {
	case c.ListLangs != nil:
		c.ListLangs.authKey = c.DeeplAuthKey
		listLanguages(c.ListLangs)
	case c.Translate != nil:
		c.Translate.authKey = c.DeeplAuthKey
		translate(c.Translate)
	}

	os.Exit(0)
}

// translate fetches an article from wikipedia, parses to markdown and translates it using DeepL
func translate(cfg *translateCmd) {
	article, err := fetch(cfg)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	markdown, err := parse(cfg, article)
	if err != nil {
		fmt.Printf("failed to parse: %s", err)
		os.Exit(2)
	}

	if cfg.ParseOnly {
		fmt.Print(markdown)
		os.Exit(0)
	}

	translated, err := translateArticle(cfg, markdown)
	if err != nil {
		fmt.Printf("failed to translateArticle: %s", err)
		os.Exit(2)
	}

	fmt.Print(translated)
	os.Exit(0)

}

// fetch downloads html from a given url
func fetch(cfg *translateCmd) (io.Reader, error) {
	resp, err := http.Get(cfg.Article)
	if err != nil {
		return nil, err
	}

	return resp.Body, err
}

// parse converts article to markdown
func parse(cfg *translateCmd, article io.Reader) (text string, err error) {
	return wikipedia.NewArticleParser().Parse(article)
}

// translateArticle translates text using DeepL API
func translateArticle(cfg *translateCmd, text string) (string, error) {
	dc := deepl.NewClient(cfg.authKey)
	translatedSentences, err := dc.Translate(text, cfg.TargetLang, "")
	if err != nil {
		return "", err
	}

	translatedText := ""
	for numSentence := range translatedSentences {
		translatedText = translatedText + translatedSentences[numSentence]
	}

	return translatedText, err

}

// listLanguages retrieves languages supported by DeepL
func listLanguages(cfg *listLangsCmd) {
	dc := deepl.NewClient(cfg.authKey)
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
