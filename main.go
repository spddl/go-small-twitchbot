package main

import (
	"bytes"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/spddl/go-twitch-ws"
)

type Bot struct {
	settings     Settings
	joinChannels []string
	isMod        map[string]bool
	commands     map[string]map[string]Command
	cooldown     CooldownContainer
	timerPool    TimerPool
	twitchBot    *twitch.Client
}

func main() {
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Lshortfile) // https://ispycode.com/GO/Logging/Setting-output-flags
	// log.SetFlags(log.Ldate | log.Lmicroseconds) // https://ispycode.com/GO/Logging/Setting-output-flags

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	bot := Bot{}
	bot.load()

	bot.joinChannels = make([]string, 0, len(bot.settings.List))
	bot.isMod = make(map[string]bool)
	bot.commands = make(map[string]map[string]Command)
	bot.cooldown = CooldownContainer{container: make(map[string]map[string]time.Time)}
	for _, value := range bot.settings.List {
		bot.joinChannels = append(bot.joinChannels, value.Channel)
		bot.isMod[value.Channel] = value.Mod
		bot.commands[value.Channel] = make(map[string]Command)
		bot.cooldown.container[value.Channel] = make(map[string]time.Time)
		for _, cmd := range value.Commands {
			bot.commands[value.Channel][cmd.Command] = cmd
		}
	}

	_twitchbot, err := twitch.NewClient(&twitch.Client{
		Server:      "wss://irc-ws.chat.twitch.tv",
		User:        bot.settings.User,
		Oauth:       bot.settings.Oauth,
		Debug:       bot.settings.Debugmessages,
		BotVerified: false,
		Channel:     bot.joinChannels,
	})
	if err != nil {
		panic(err)
	}
	bot.twitchBot = _twitchbot

	bot.twitchBot.OnPrivateMessage = func(ircMsg twitch.IRCMessage) {
		channel := string(ircMsg.Params[0][1:])
		val, found := bot.commands[channel][string(ircMsg.Params[1])]
		if found { // Commmand gefunden
			bot.Say(ircMsg, channel, val.Command)
		}
	}

	bot.twitchBot.OnNoticeMessage = func(ircMsg twitch.IRCMessage) {
		// log.Printf("OnNoticeMessage: %s\n", ircMsg) // debug
		log.Printf("%s: %s\n", ircMsg.Params[0][1:], ircMsg.Params[1])
	}

	bot.twitchBot.OnUserNoticeMessage = func(ircMsg twitch.IRCMessage) {
		// log.Printf("OnUserNoticeMessage: %s\n\n", ircMsg) // debug
		channel := string(ircMsg.Params[0][1:])
		var msg string
		if len(ircMsg.Params) > 1 {
			msg = " (" + string(ircMsg.Params[1]) + ")"
		}
		systemMsg := string(bytes.ReplaceAll(ircMsg.Tags["system-msg"], []byte{92, 115}, []byte{32}))
		log.Printf("%s: %s%s\n", channel, systemMsg, msg)
	}

	bot.twitchBot.OnJoinMessage = func(msg twitch.IRCMessage) {
		log.Printf("Join channel: %s", msg.Params[0][1:]) // to remove # from Channel Parameter
		var data []Command
		var channel = string(msg.Params[0][1:])
		for _, l := range bot.settings.List {
			if l.Channel != channel { // suche den richtigen Streamer
				continue
			}
			for _, c := range l.Commands { // Durchsuche alle Commands nach dem Wert "Repeating"
				if c.Repeating != "" {
					data = append(data, c)
				}
			}
		}
		if len(data) != 0 {
			go bot.RepeatedMessages(channel, data)
		}
	}

	bot.twitchBot.OnPartMessage = func(msg twitch.IRCMessage) { // TODO: https://golang.org/pkg/time/#NewTicker
		log.Printf("Leave channel: %s", msg.Params[0][1:]) // to remove # from Channel Parameter
		bot.timerPool.mu.Lock()
		for _, t := range bot.timerPool.TimeChan[string(msg.Params[0][1:])] {
			t.Stop()
		}
		bot.timerPool.mu.Unlock()
	}

	bot.twitchBot.OnUnknownMessage = func(ircMsg twitch.IRCMessage) {
		log.Printf("OnUnknownMessage: %s\n", ircMsg)
	}
	bot.twitchBot.Run()

	for { // ctrl - c
		<-interrupt
		bot.twitchBot.Close()
		os.Exit(0)
	}
}
