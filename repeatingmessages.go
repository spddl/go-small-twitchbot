package main

import (
	"log"
	"sort"
	"sync"
	"time"

	"github.com/spddl/go-twitch-ws"
)

type TimerPool struct {
	TimeChan map[string][]*time.Ticker
	mu       sync.RWMutex
}

type Pool struct {
	nanoseconds int
	duration    time.Duration
	cmd         Command
}

func (bot *Bot) RepeatedMessages(streamer string, data []Command) {
	var allTimers []Pool
	var sumInterval int
	for _, command := range data {
		duration, err := time.ParseDuration(command.Repeating)
		if err != nil {
			log.Println(err) // läuft auf Fehler wenn der Wert zu hoch ist
		}

		DiN := int(duration.Nanoseconds())
		sumInterval += DiN
		allTimers = append(allTimers, Pool{nanoseconds: DiN, duration: duration, cmd: command})
	}
	sort.Slice(allTimers, func(i, j int) bool { return allTimers[i].nanoseconds < allTimers[j].nanoseconds }) // shortest first

	var lenIntervals = len(allTimers)
	for i, p := range allTimers {
		bot.timerPool.TimeChan = make(map[string][]*time.Ticker)
		go func(i int, p Pool, sumInterval int, channel string) {
			if i == 0 {
				log.Printf("'%s' every %s", p.cmd.Message, p.duration)
				bot.Say(twitch.IRCMessage{}, channel, p.cmd.Command)
			} else {
				var delay int
				if lenIntervals == sumInterval/allTimers[i].nanoseconds && sumInterval%allTimers[i].nanoseconds == 0 {
					delay = allTimers[i].nanoseconds / lenIntervals * i
				} else {
					delay = sumInterval / lenIntervals * i
				}

				for delay > allTimers[i].nanoseconds {
					delay -= allTimers[i].nanoseconds
				}

				delayDuration := time.Duration(delay) * time.Nanosecond
				log.Printf("'%s' starts in %s an then every %s", p.cmd.Message, delayDuration.Truncate(time.Second), p.duration)
				time.Sleep(delayDuration)
			}
			ticker := time.NewTicker(p.duration)
			bot.timerPool.mu.Lock()
			bot.timerPool.TimeChan[channel] = append(bot.timerPool.TimeChan[channel], ticker)
			bot.timerPool.mu.Unlock()
			for {
				<-ticker.C
				bot.Say(twitch.IRCMessage{}, channel, p.cmd.Command)
			}
		}(i, p, sumInterval, streamer)
	}
}
