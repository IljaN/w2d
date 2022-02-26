package main

import (
	"fmt"
	"github.com/alexflint/go-arg"
	"os"
	"os/exec"
)

type Config struct {
	Article       string `arg:"required,positional"`
	TargetLang    string `arg:"-l,--" help:"target language for translation"`
	WikExecutable string `arg:"-w,--" help:"Path to the wik executable" default:"/usr/local/bin/wik"`
	DeeplAuthKey  string `arg:"env:DEEPL_AUTH_KEY"`
}

func loadConfig() *Config {
	c := Config{}
	arg.MustParse(&c)
	return &c
}

func fetchArticle(cfg *Config) (text []byte, err error) {
	termFlag := fmt.Sprintf("-s %s", cfg.Article)
	cmd := exec.Command(cfg.WikExecutable, termFlag)
	return cmd.Output()
}

func main() {
	cfg := loadConfig()
	txt, err := fetchArticle(cfg)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println(txt)
	os.Exit(0)

}
