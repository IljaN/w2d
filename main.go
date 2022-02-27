package main

import (
	"fmt"
	"github.com/IljaN/w2d/deepl"
	"github.com/alexflint/go-arg"
	"os"
	"os/exec"
	"strings"
)

type Config struct {
	Article       string `arg:"required,positional"`
	TargetLang    string `arg:"-l,--" help:"target language for translation"`
	WikExecutable string `arg:"-w,--" help:"Path to the wik executable" default:"/usr/local/bin/wik"`
	DeeplAuthKey  string `arg:"required,--,env:DEEPL_AUTH_KEY"`
}

func loadConfig() *Config {
	c := Config{}
	arg.MustParse(&c)
	c.DeeplAuthKey = strings.TrimSpace(c.DeeplAuthKey)

	return &c
}

func fetchArticle(cfg *Config) (text []byte, err error) {
	termFlag := fmt.Sprintf("-s %s", cfg.Article)
	cmd := exec.Command(cfg.WikExecutable, termFlag)
	return cmd.Output()
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
	txt, err := fetchArticle(cfg)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	translated, err := translate(cfg, string(txt))
	fmt.Println(translated)

	os.Exit(0)

}
