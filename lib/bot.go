/*
This package is the core of the slacksoc bot library. It defines the main
architecture of the Bot, event handling, plugin, and configuration interface.
*/
package lib

import "fmt"
import "os"
import "github.com/nlopes/slack"

/*
Bot contains all plugins and handlers. It manages the Slack API connection and
dispatches events as they happen.
*/
type Bot struct {
	API      *slack.Client
	RTM      *slack.RTM
	handlers map[string][]EventHandler
	plugins  map[string]Plugin
}

/*
Creates a new bot instance. This only initializes the API instance. The RTM
connection will not happen until you call RunForever() on the bot.
*/
func newBot(key string) *Bot {
	API := slack.New(key)
	API.SetDebug(true)
	return &Bot{API: API, RTM: nil, handlers: make(map[string][]EventHandler)}
}

/*
Register an EventHandler to be called whenever a specific type of event occurs.
You can register the same EventHandler to multiple events with separate calls
to this function.
*/
func (bot *Bot) OnEvent(tp string, eh EventHandler) {
	bot.handlers[tp] = append(bot.handlers[tp], eh)
}

/*
Register a MessageHandler to be called whenever a specific subtype of the
"message" event occurs: https://api.slack.com/events/message

Use an empty subType ("") for normal messages (i.e., none of those subtypes).
*/
func (bot *Bot) OnMessage(subType string, mh MessageHandler) {
	bot.OnEvent("message", func(bot *Bot, evt slack.RTMEvent) error {
		msgEvent := evt.Data.(*slack.MessageEvent)
		if msgEvent.Msg.SubType == subType {
			return mh(bot, msgEvent)
		} else {
			return nil
		}
	})
}

/*
This actually connects the bot to Slack and begins running it "forever".
*/
func (bot *Bot) RunForever() {
	bot.RTM = bot.API.NewRTM()
	go bot.RTM.ManageConnection()

	for evt := range bot.RTM.IncomingEvents {
		handlers := bot.handlers[evt.Type]
		for _, handler := range handlers {
			handler(bot, evt)
		}
	}
}

/*
Create a bot object using command line arguments. Typically, all and end-user
application should need to do is call third-party plugin registration functions,
and then call this Run() function.
*/
func Run() {
	bot := newBot(os.Args[2])
	err := bot.configure(os.Args[1])
	if err != nil {
		fmt.Println(err)
		return
	}
	bot.RunForever()
}
