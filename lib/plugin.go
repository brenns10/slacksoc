package lib

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
PluginConstructor is a function which will return a new instance of a Plugin. It
may perform a wide array of activities, including registering handlers loading
data, starting goroutines, etc. The bot parameter points to the Bot containing
the plugin. The name parameter contains the instance name of this plugin (there
may be many instances of a plugin). The config parameter contains configuration
data loaded from YAML. The recommended use of config is to parse it directly
using: https://godoc.org/github.com/mitchellh/mapstructure
*/
type PluginConstructor func(bot *Bot, name string, config PluginConfig) Plugin

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
