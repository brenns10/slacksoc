/*
Slacksoc is a Slack bot creation framework. This package, lib, is the core of
the framework. This documentation should serve as a good API reference. A guided
introduction to the framework is available on the GitHub wiki:
https://github.com/brenns10/slacksoc/wiki
*/
package lib

import "encoding/gob"
import "fmt"
import "os"
import "regexp"
import "sync"
import "time"

import "github.com/nlopes/slack"
import "github.com/sirupsen/logrus"
import "github.com/google/shlex"

/*
Bot is the publicly exported type that contains most of the Slacksoc framework.
Most use of the Bot will occur within plugins, since they receive pointers to
the bot in constructors and event handlers. Public attributes may be accessed
without synchronization, but should not be modified.

No public constructor exists for the Bot as of now. Users of the library should
only need to call the library's Run function (see its docs for more info).
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

	// This stuff is for plugin state and saving.
	state      map[string][]byte
	stateDelay int
	stateFile  string
	stateDirty bool
	stateChan  chan pluginStateEvent

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
		state:         make(map[string][]byte),
		stateChan:     make(chan pluginStateEvent, 100),
		plugins:       make(map[string]Plugin),
		handlers:      make(map[string][]EventHandler),
	}
	bot.registerInfoHandlers()
	bot.OnCommand("help", helpCommand)
	return bot
}

/*
Register an EventHandler to be called whenever a specific type of Slack RTM
event occurs. You can register the same EventHandler to multiple events with
separate calls to this function.

This is the most basic event handler in the bot - all other types of handlers
and their associated registration functions eventually call this. Here is the
hierarchy of handlers and registration functions.

    EventHandler, OnEvent()
    -> MessageHandler, OnMessage()
       -> MessageHandler, OnAddressed()
          -> MessageHandler, OnAddressedMatch()
          -> MessageHandler, OnAddressedMatchExpr()
          -> CommandHandler, OnCommand()
       -> MessageHandler, OnMatch()
       -> MessageHandler, OnMatchExpr()
*/
func (bot *Bot) OnEvent(type_ string, eh EventHandler) {
	bot.handlers[type_] = append(bot.handlers[type_], eh)
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
is modified so that it only includes the text after the part that "addresses"
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
regular expression. The message need not be addressed to the bot.
*/
func (bot *Bot) OnMatch(regex string, mh MessageHandler) {
	bot.OnMessage("", IfMatch(regex, mh))
}

/*
Same as Bot.OnMatch, but takes a compiled regex.
*/
func (bot *Bot) OnMatchExpr(expr *regexp.Regexp, mh MessageHandler) {
	bot.OnMessage("", IfMatchExpr(expr, mh))
}

/*
Register a MessageHandler to be called whenever a message (subtype "") addressed
to the bot matches a regular expression.
*/
func (bot *Bot) OnAddressedMatch(regex string, mh MessageHandler) {
	bot.OnAddressed(IfMatch(regex, mh))
}

/*
Same as Bot.OnAddressedMatch, but takes a compiled regex.
*/
func (bot *Bot) OnAddressedMatchExpr(expr *regexp.Regexp, mh MessageHandler) {
	bot.OnAddressed(IfMatchExpr(expr, mh))
}

/*
Register a CommandHandler to be called when a message addressed to the bot is a
particular command. The handler receives parsed arguments, assuming that the
first argument is cmd. See the documentation for CommandHandler for more details.
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
This function saves state if necessary.
*/
func (bot *Bot) saveState() {
	file, err := os.Create(bot.stateFile)
	if err != nil {
		bot.Log.WithFields(logrus.Fields{
			"error":    err,
			"filename": bot.stateFile,
		}).Error("Error opening statefile for save. Continuing.")
		return
	}

	enc := gob.NewEncoder(file)
	enc.Encode(bot.state)
	file.Close()
	bot.stateDirty = false
}

/*
This function starts the Slack RTM connection and runs the bot "forever".
*/
func (bot *Bot) runForever() {
	bot.RTM = bot.API.NewRTM()
	go bot.RTM.ManageConnection()

	for {
		select {
		case evt := <-bot.RTM.IncomingEvents:
			handlers := bot.handlers[evt.Type]
			bot.Log.WithFields(logrus.Fields{
				"type": evt.Type,
			}).Info("Handling a message.")
			for _, handler := range handlers {
				handler(bot, evt)
			}
			break
		case state := <-bot.stateChan:
			if state.Type == "save" {
				bot.Log.Info("Saving state...")
				bot.saveState()
			} else if state.Type == "update" {
				bot.Log.WithFields(logrus.Fields{
					"plugin": state.Plugin,
				}).Info("Received state update")
				bot.state[state.Plugin] = state.State
				if !bot.stateDirty {
					// only queue a new save when the state /becomes/ dirty
					go func() {
						time.Sleep(time.Duration(bot.stateDelay) * time.Second)
						bot.stateChan <- pluginStateEvent{Type: "save"}
					}()
				}
				bot.stateDirty = true
			} else {
				bot.Log.WithFields(logrus.Fields{
					"type": state.Type,
				}).Warn("Unknown state event encountered.")
			}
			break
		}
	}
}

/*
Create and run a bot object using command line arguments. Currently, one command
line argument is expected: the path of a YAML configuration file. The
configuration file should at least contain the following:

    token: SLACK API TOKEN
    plugins:
      - name: PluginName
        # put any additional plugin configuration here
      - name: NextPluginName

Since Go does not allow dynamic loading, all Plugins must be registered before
this function is invoked. If you use only the core plugins provided, the
slacksoc binary is good enough. If you are creating your own bot, you will need
to register your plugins (as well as the core plugins) before calling run. As
an example:

    package main
    import "github.com/brenns10/slacksoc/lib"
    import "github.com/brenns10/slacksoc/plugins"

    func main() {
        plugins.Register()
        lib.Register("MyPlugin", newMyPlugin)
        lib.Run()
    }
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
	bot.runForever()
}
