package lib

import "bytes"
import "encoding/gob"
import "fmt"

import "github.com/nlopes/slack"
import "github.com/sirupsen/logrus"

/*
Plugin is an interface that all plugins must satisfy. The defined interface
functions mostly center around giving help information to end users.

Describe() returns a one-line string that describes this plugin in a nutshell.
Help() returns a (probably longer) string that is used when the user asks for
help on this plugin in particular.

In order to use your plugin, you need to register its constructor with the
Register function, and then add an entry for it in the bot config.
*/
type Plugin interface {
	Describe() string
	Help() string
}

/*
PluginConstructor is a function which will return a new instance of a Plugin.
The constructor must have been registered with the bot, and it will be called
during bot startup, if the plugin is requested by the configuration.

This function is called with a pointer to the bot, as well as the plugin's name
(as registered) and configuration data. See PluginConfig docs for more
information on that.

A PluginConstructor may perform a wide array of activities. Typical activities
are registering handlers and loading configuration. More complex plugins may
wish to start goroutines here, connect to APIs etc. This function must complete
before the bot starts up, so blocking could be a concern, but not nearly as much
concern as in an event handler.

A few things are off limits within the constructor. The bot is not yet connected
to Slack at this stage. As a direct result, the RTM field of the bot may not be
used. More importantly, the functions which get users and channels may not be
used either. If you wish to use those, consider registering a handler for the
"hello" event.

The API field of the bot is initialized at this point, so constructors may use
that freely.
*/
type PluginConstructor func(bot *Bot, name string, config PluginConfig) Plugin

/*
pluginStateEvent is used when plugins update their internal state and need to be
saved. Plugin state is simply bytes (marshalled from their state data
structures) and it is associated with a name.
*/
type pluginStateEvent struct {
	Type   string
	Plugin string
	State  []byte
}

/*
Internal registry of plugin constructors.
*/
var plugins = make(map[string]PluginConstructor)

/*
This function will register a plugin constructor with the slacksoc library. Your
name should be unique among all plugins, so one option could be the fully
qualified Go import name.
*/
func Register(name string, ctor PluginConstructor) {
	plugins[name] = ctor
}

/*
Load the plugin's saved state into dest. Will not do anything if there was no
saved state.
*/
func (bot *Bot) GetState(plugin string, dest interface{}) {
	state, ok := bot.state[plugin]
	if !ok {
		return
	}
	dec := gob.NewDecoder(bytes.NewReader(state))
	err := dec.Decode(dest)
	if err != nil {
		bot.Log.WithFields(logrus.Fields{
			"plugin": plugin,
		}).Error("Failed to GetState for plugin. Continuing.")
	}
}

/*
Update your plugin's state. Plugins should call this function whenever their
persisted state has changed. Their state will be saved to disk based on the bot
timeout policy.
*/
func (bot *Bot) UpdateState(plugin string, state interface{}) {
	var buf bytes.Buffer
	var event pluginStateEvent
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(state)
	if err != nil {
		bot.Log.WithFields(logrus.Fields{
			"plugin": plugin,
			"error":  err,
		}).Error("Failed to UpdateState for plugin. Continuing.")
		return
	}
	event.Type = "update"
	event.Plugin = plugin
	event.State = buf.Bytes()
	bot.stateChan <- event
}

/*
This handler is for showing help on all commands.
*/
func helpCommand(bot *Bot, evt *slack.MessageEvent, args []string) error {
	var msg bytes.Buffer
	if len(args) <= 1 {
		msg.WriteString(fmt.Sprintf(
			"I am %s. I have many plugins:\n\n", bot.User.Name,
		))
		for name, plugin := range bot.plugins {
			msg.WriteString("*" + name + ":* ")
			msg.WriteString(plugin.Describe())
			msg.WriteString("\n")
		}
		msg.WriteString("\nUse `help PLUGIN` for more information on a plugin.")
		bot.Reply(evt, msg.String())
	} else {
		plugin, ok := bot.plugins[args[1]]
		if ok {
			bot.Reply(evt, plugin.Help())
		} else {
			bot.Reply(evt, fmt.Sprintf(
				"Sorry, I couldn't find the plugin \"%s\"", args[1],
			))
		}
	}
	return nil
}
