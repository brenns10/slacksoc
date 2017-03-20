/*
Slacksoc is a Slack bot creation framework. This package, lib, is th= ue core of
the framework. This documentation is the main reference for plugin developers.
*/
package lib

import "fmt"
import "os"
import "regexp"
import "sync"

import "github.com/nlopes/slack"
import "github.com/sirupsen/logrus"
import "github.com/google/shlex"

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
Creates a new bot instance. This initializes the internal data structures, as
well as the bot Logger. However, the API object is not initialized until the bot
is configured. The RTM object is not initialized until the bot starts its "run
forever" loop. The User and Team objects are not initialized until the bot
receives the server's "hello" event.
*/
func newBot() *Bot {
	Log := logrus.New()
	Log.Level = logrus.DebugLevel
	bot := &Bot{
		API:           nil,
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
	bot.OnCommand("help", helpCommand)
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
Register a MessageHandler to be called when a message comes in (subtype "") and
it is addressed to the bot. The definition of "addressed" depends on the
situation. In a channel or group, this is a message that begins with the bot
username or an @mention of the bot, followed by an optional colon and whitespace.
In a direct message, any message is considered "addressed" to the bot.

Unlike a regular handler registered with OnMessage(), the event.Msg.Text field
is modified so that it only includes the text "after" the part that "addresses"
the message to the bot. So a message like "@slacksoc: hello there" would become
"hello there" for a handler registered with OnAddressed()
*/
func (bot *Bot) OnAddressed(mh MessageHandler) {
	bot.OnMessage("", func(bot *Bot, evt *slack.MessageEvent) error {
		// We need to compile the regex *here* because when plugins register
		// their handlers, the User/Team fields have not been initialized yet.
		// Could optimize this by placing the compiled regex into a struct field
		// which is initialized by the hello message handler.
		re := regexp.MustCompile(fmt.Sprintf(
			`\s*(<@%s(\|\w+)?>|@?%s):?\s+`, bot.User.ID, bot.User.Name,
		))
		match := re.FindAllStringIndex(evt.Msg.Text, 1)
		if match != nil && match[0][0] == 0 {
			// replace Msg.Text, but restore it after
			oldText := evt.Msg.Text
			evt.Msg.Text = evt.Msg.Text[match[0][1]:]
			rv := mh(bot, evt)
			evt.Msg.Text = oldText
			return rv
		}
		if IsDM(evt.Msg.Channel) {
			return mh(bot, evt)
		}
		return nil
	})
}

/*
Register a MessageHandler to be called whenever a message (subtype "") matches a
regular expression.
*/
func (bot *Bot) OnMatch(regex string, mh MessageHandler) {
	bot.OnMatchExpr(regexp.MustCompile(regex), mh)
}

/*
Same as Bot.OnMatch, but takes a compiled regex.
*/
func (bot *Bot) OnMatchExpr(expr *regexp.Regexp, mh MessageHandler) {
	bot.OnAddressed(func(bot *Bot, evt *slack.MessageEvent) error {
		match := expr.FindStringIndex(evt.Msg.Text)
		if len(match) > 0 {
			return mh(bot, evt)
		}
		return nil
	})
}

/*
Register a CommandHandler to be called when a message addressed to the bot is a
particular command.
*/
func (bot *Bot) OnCommand(cmd string, ch CommandHandler) {
	bot.OnAddressed(func(bot *Bot, evt *slack.MessageEvent) error {
		args, err := shlex.Split(evt.Msg.Text)
		if err != nil {
			return nil // bad command line syntax is not an error :)
		}
		if args[0] == cmd {
			return ch(bot, evt, args)
		}
		return nil
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
		bot.Log.WithFields(logrus.Fields{
			"type": evt.Type,
		}).Info("Handling a message.")
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
	if len(os.Args) < 2 {
		fmt.Printf("usage: %s CONFIG\n", os.Args[0])
		return
	}
	bot := newBot()
	err := bot.configure(os.Args[1])
	if err != nil {
		fmt.Println(err)
		return
	}
	bot.RunForever()
}
