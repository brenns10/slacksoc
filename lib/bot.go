/*
The lib package of slacksoc contains all the core machinery for making a bot
with several plugins work well.
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
	Handlers map[string][]EventHandler
}

/*
Creates a new bot instance. This only initializes the API instance. The RTM
connection will not happen until you call RunForever() on the bot.
*/
func newBot(key string) *Bot {
	API := slack.New(key)
	API.SetDebug(true)
	return &Bot{API: API, RTM: nil, Handlers: make(map[string][]EventHandler)}
}

/*
Register an EventHandler.
*/
func (bot *Bot) OnEvent(tp string, eh EventHandler) {
	bot.Handlers[tp] = append(bot.Handlers[tp], eh)
}

/*
Register a MessageHandler. This is still an EventHandler under the hood, but
this definitely makes client code a little prettier.
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
Run the bot forever
*/
func (bot *Bot) RunForever() {
	bot.RTM = bot.API.NewRTM()
	go bot.RTM.ManageConnection()

	for evt := range bot.RTM.IncomingEvents {
		fmt.Printf("Don't worry, I'll handle it.\n")
		handlers := bot.Handlers[evt.Type]
		for _, handler := range handlers {
			fmt.Printf("No seriously, I got this!.\n")
			handler(bot, evt)
		}
		fmt.Printf("See?\n")
	}
}

/*

 */
func Run() {
	bot := newBot(os.Args[1])
	bot.OnMessage("", func(bot *Bot, msgEvent *slack.MessageEvent) error {
		fmt.Printf("[%s]%s: %s\n", msgEvent.Msg.Channel, msgEvent.Msg.User,
			msgEvent.Msg.Text)
		return nil
	})
	bot.RunForever()
}
