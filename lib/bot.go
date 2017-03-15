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
	Handlers map[string][]Handler
}

/*
Creates a new bot instance by connecting to Slack. This doesn't do any plugin
loading or initialization.
*/
func newBot(key string) *Bot {
	API := slack.New(key)
	API.SetDebug(true)
	return &Bot{API: API, RTM: nil, Handlers: make(map[string][]Handler)}
}

/*
Register a Handler.
*/
func (bot *Bot) Register(tp string, handler Handler) {
	var handlerList []Handler
	var ok bool
	handlerList, ok = bot.Handlers[tp]
	if !ok {
		handlerList = make([]Handler, 0, 5)
	}
	handlerList = append(handlerList, handler)
	bot.Handlers[tp] = handlerList
}

/*
Register a callback handler.
*/
func (bot *Bot) RegisterF(tp string, plugin *Plugin, cb HandlerCallback) {
	bot.Register(tp, NewHandlerFunc(plugin, cb))
}

/*
Run the bot forever
*/
func (bot *Bot) RunForever() {
	bot.RTM = bot.API.NewRTM()
	go bot.RTM.ManageConnection()

	fmt.Printf("Handlers\n")
	fmt.Printf("%+v\n", bot.Handlers)
	for evt := range bot.RTM.IncomingEvents {
		handlers, ok := bot.Handlers[evt.Type]
		if !ok {
			continue
		}
		fmt.Printf("handlers for event\n")
		fmt.Printf("%+v\n", handlers)
		for _, handler := range handlers {
			handler.Handle(bot, evt)
		}
	}
}

func Run() {
	bot := newBot(os.Args[1])
	bot.RegisterF("message", nil, func(bot *Bot, evt slack.RTMEvent, data interface{}) {
		fmt.Println(evt)
	})
	bot.RunForever()
}
