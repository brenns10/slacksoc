package lib

/*
Plugin is an interface that all plugins must satisfy. The defined interface
functions mostly center around giving help information to end users.

Describe() returns a one-line string that describes this plugin in a nutshell.
Help() returns a (probably longer) string that is used when the user asks for
help on this plugin in particular. Name() returns the instance name that was
passed into the PluginConstructor.
*/
type Plugin interface {
	Name() string
	Describe() string
	Help() string
}

/*
PluginConstructor is a function which will return a new instance of a Plugin. It
may perform a wide array of activities, including registering handlers loading
data, starting goroutines, etc. The bot parameter points to the Bot containing
the plugin. The name parameter contains the instance name of this plugin (there
may be many instances of a plugin). The config parameter contains configuration
data.
TODO: The config should be a custom object, likely derived from a TOML library.
*/
type PluginConstructor func(bot *Bot, name string, config string)
