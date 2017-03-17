/*
Slacksoc is a Slack bot creation framework. This package, lib, is the core of
the framework. This documentation is the main reference for plugin developers.
*/
package lib

import "fmt"
import "log"
import "os"
import "sync"

import "github.com/nlopes/slack"
import "github.com/sirupsen/logrus"

/*
Bot contains all plugins and handlers. It manages the Slack API connection and
dispatches events as they happen.
*/
type Bot struct {
	API      *slack.Client  // should not need sync
	RTM      *slack.RTM     // does not need sync
	Log      *logrus.Logger // no need to sync
	Info     *slack.Info    // DO NOT ACCESS WITHOUT LOCKING
	InfoLock sync.RWMutex   // this is the lock for ^

	// yeah plugins can't touch this
	handlers map[string][]EventHandler
	plugins  map[string]Plugin
}

/*
This handler waits for the hello message and then loads the info.
*/
func (bot *Bot) helloHandler(_ *Bot, _ slack.RTMEvent) error {
	bot.InfoLock.Lock()
	bot.Info = bot.RTM.GetInfo()
	bot.InfoLock.Unlock()
	return nil
}

/*
Creates a new bot instance. This only initializes the API instance. The RTM
connection will not happen until you call RunForever() on the bot.
*/
func newBot(key string) *Bot {
	API := slack.New(key)
	API.SetDebug(true)
	Log := logrus.New()
	Log.Level = logrus.DebugLevel
	slack.SetLogger(log.New(Log.WriterLevel(logrus.DebugLevel), "", 0))
	bot := &Bot{
		API:      API,
		RTM:      nil,
		Log:      Log,
		handlers: make(map[string][]EventHandler),
		plugins:  make(map[string]Plugin),
	}
	bot.OnEvent("hello", bot.helloHandler)
	return bot
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
			bot.Log.WithFields(logrus.Fields{
				"type": evt.Type,
			}).Info("Handling a message.")
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
