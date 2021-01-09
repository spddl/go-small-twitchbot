# go-small-twitchbot

settings.yaml

```yaml
---
user: username
oauth: abcdefghijklmnopqrstuvwxyzabcd # https://twitchapps.com/tmi/
debugmessages: true
list:
- channel: twitch # only lowercase
  mod: false      # is this user mod on this channel?
  commands:

  - cmd: "!command"
    msg: "@{{.Username}} => This is a Test Command" # https://golang.org/pkg/text/template/#hdr-Actions
    cooldown: "30s"                                 # https://golang.org/pkg/time/#ParseDuration ParseDuration parses a duration string. A duration string is a possibly signed sequence of decimal numbers, each with optional fraction and a unit suffix, such as "300ms", "-1.5h" or "2h45m". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".

  - cmd: "!jsonapi"
    msg: "@{{.Username}} => {{ .Api.title }}" # https://golang.org/pkg/text/template/#hdr-Actions
    httprequest:
      url: https://jsonplaceholder.typicode.com/todos/1
      headers:
        - field: Accept
          value: application/json

  - cmd: "!ads"
    msg: "#Ads# Text every 1h" # if repeating is set {{.Username}} will dont work
    repeating: "1h"            # https://golang.org/pkg/time/#ParseDuration ParseDuration parses a duration string. A duration string is a possibly signed sequence of decimal numbers, each with optional fraction and a unit suffix, such as "300ms", "-1.5h" or "2h45m". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
    cooldown: "30s"

- channel: twitch2 # only lowercase
# ...
```
