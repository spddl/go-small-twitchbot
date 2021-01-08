package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type Settings struct {
	User          string `yaml:"user"`
	Oauth         string `yaml:"oauth"`
	Debugmessages bool   `yaml:"debugmessages"`
	List          []struct {
		Channel  string    `yaml:"channel"`
		Mod      bool      `yaml:"mod"`
		Commands []Command `yaml:"commands"`
	} `yaml:"list"`
}

type Command struct {
	Command     string `yaml:"cmd"`
	Message     string `yaml:"msg"`
	Cooldown    string `yaml:"cooldown"`
	Repeating   string `yaml:"repeating"`
	Httprequest struct {
		URL     string `yaml:"url"`
		Headers []struct {
			Field string `yaml:"field"`
			Value string `yaml:"value"`
		} `yaml:"headers"`
	} `yaml:"httprequest,omitempty"`
}

func (bot *Bot) load() {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)

	content, err := ioutil.ReadFile(filepath.Join(exPath, "settings.yaml"))
	if err != nil {
		log.Fatal(err)
	}

	err = yaml.Unmarshal(content, &bot.settings)
	if err != nil {
		log.Fatalf("cannot unmarshal data: %v", err)
	}
}
