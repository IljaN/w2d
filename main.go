package main

import (
	"fmt"
	"github.com/IljaN/w2d/deepl"
	"github.com/IljaN/w2d/wikipedia"
	"github.com/alexflint/go-arg"
	"io"
	"net/http"
	"os"
	"strings"
)

type Config struct {
	Article      string `arg:"required,positional"`
	TargetLang   string `arg:"-l,--" help:"target language for translation"`
	ParseOnly    bool   `arg:"-p,--" default:"false" help:"parse to markdown without translating"`
	DeeplAuthKey string `arg:"required,--,env:DEEPL_AUTH_KEY"`
}

func loadConfig() *Config {
	c := Config{}
	arg.MustParse(&c)
	c.DeeplAuthKey = strings.TrimSpace(c.DeeplAuthKey)

	return &c
}

func fetch(cfg *Config) (io.Reader, error) {
	resp, err := http.Get(cfg.Article)
	if err != nil {
		return nil, err
	}

	return resp.Body, err
}

func parse(cfg *Config, article io.Reader) (text string, err error) {
	return wikipedia.NewArticleParser().Parse(article)
}

func translate(cfg *Config, text string) (string, error) {
	dc := deepl.NewClient(cfg.DeeplAuthKey)
	translatedSentences, err := dc.Translate(text, "RU", "")
	if err != nil {
		return "", err
	}

	translatedText := ""
	for numSentence := range translatedSentences {
		translatedText = translatedText + translatedSentences[numSentence]
	}

	return translatedText, err

}

func main() {
	cfg := loadConfig()
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

	translated, err := translate(cfg, markdown)
	if err != nil {
		fmt.Printf("failed to translate: %s", err)
		os.Exit(2)
	}

	fmt.Print(translated)
	os.Exit(0)
}
