package main

import (
	"bytes"
	"encoding/json"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/spddl/go-twitch-ws"
)

func (bot *Bot) Say(ircMsg twitch.IRCMessage, channel, command string) {
	if bot.commands[channel][command].Cooldown != "" && !bot.cooldown.Test(channel, command) {
		return
	}

	var jsonApi interface{}
	if bot.commands[channel][command].Httprequest.URL != "" {
		c := make(chan interface{})
		go fetch(twitch.IRCMessage{}, bot.commands[channel][command], c)
		jsonApi = <-c
	}
	tmpl := generateMessage(twitch.IRCMessage{}, bot.commands[channel][command], jsonApi)

	bot.twitchBot.Say(channel, tmpl, bot.isMod[channel])
	if bot.commands[channel][command].Cooldown != "" {
		bot.cooldown.Set(channel, bot.commands[channel][command].Command, bot.commands[channel][command].Cooldown)
	}
}

func generateMessage(ircMsg twitch.IRCMessage, command Command, jsonParsed interface{}) string {
	tmpl := template.Must(template.New("").Funcs(template.FuncMap{
		"parseRFC3339": parseRFC3339,
	}).Parse(command.Message))

	buf := new(bytes.Buffer)
	err := tmpl.Execute(buf, struct {
		IrcMsg   twitch.IRCMessage
		Api      interface{}
		Username string
	}{
		ircMsg,
		jsonParsed,
		string(ircMsg.Tags["display-name"]),
	})
	if err != nil {
		log.Println(err)
		return ""
	}

	return buf.String()
}

func parseRFC3339(timestamp string) (string, error) { // https://golang.org/pkg/html/template/#Template.Funcs
	t, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return "", err
	}

	return time.Since(t).String(), nil
}

func fetch(ircMsg twitch.IRCMessage, command Command, c chan interface{}) {
	req, err := http.NewRequest("GET", command.Httprequest.URL, nil)
	if err != nil {
		log.Println(err)
	}

	for _, v := range command.Httprequest.Headers {
		req.Header.Set(v.Field, v.Value)
	}

	client := http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
	}

	// log.Println("body:", string(body)) // debug

	var jsonParsed interface{}
	if err := json.Unmarshal(body, &jsonParsed); err != nil {
		log.Println(err)
	}
	resp.Body.Close()

	c <- jsonParsed
}
