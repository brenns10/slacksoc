/*
Slacksoc is a Slack bot creation framework. This package, lib, is th= ue core of
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
	// These are public attributes, and can be accessed with no lock.
	API  *slack.Client      // probably thread safe
	RTM  *slack.RTM         // definitely thread safe
	Log  *logrus.Logger     // definitely thread safe
	User *slack.UserDetails // read only
	Team *slack.Team        // read only

	// These private attributes require read/write locking to access safely.
	// They have helper methods so that calling code need not worry about it.
	infoLock      sync.RWMutex
	userByName    map[string]*slack.User
	userByID      map[string]*slack.User
	channelByName map[string]string
	channelByID   map[string]string

	// These private attributes should just never be accessed outside of the
	// main bot thread. They have no helper methods.
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
	Log := logrus.New()
	Log.Level = logrus.DebugLevel
	slack.SetLogger(log.New(Log.WriterLevel(logrus.DebugLevel), "", 0))
	bot := &Bot{
		API:           API,
		RTM:           nil,
		Log:           Log,
		userByName:    make(map[string]*slack.User),
		userByID:      make(map[string]*slack.User),
		channelByName: make(map[string]string),
		channelByID:   make(map[string]string),
		plugins:       make(map[string]Plugin),
		handlers:      make(map[string][]EventHandler),
	}
	bot.registerInfoHandlers()
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
