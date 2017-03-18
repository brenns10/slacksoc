slacksoc
========

This is a new, and maybe improved Slack bot library for Go.

Features:
- Plugin based architecture
- Configure plugins through a single YAML file
- Entire Slack API exposed to plugins

Included Plugins:
- triggers / responses, like the original Slackbot

Usage
-----

If all you would like to do is use the included plugins, usage is simple. First,
install the `slacksoc` binary:

    go get github.com/brenns10/slacksoc/slacksoc
    
Create a YAML configuration file - see [sample.yaml](sample.yaml) for an
example. Also, get a bot API token from your Slack. Then, you can run as
follows:

    slacksoc config.yaml API_KEY

### Using External Plugins

If you would like to implement your own plugins, or use a third-party plugin (if
they ever exist), you will need to write a small amount of boilerplate code.
This is because Go has no support for dynamic module loading, and therefore
plugins need to be registered with the bot before they can be used. Here is a
complete sample:

```go
package main

import "github.com/brenns10/slacksoc/lib"
import "github.com/brenns10/slacksoc/plugins"

// your plugin code here

func main() {
    plugins.Register()
    lib.Register("plugin name", PluginConstructor)
    lib.Run()
}
```

### Developing Plugins

Plugin development is rather simple. First, create a struct that implements the
`Plugin` interface. It can store any state your plugin needs. Next, implement a
constructor for your struct, which fits the signature of a `PluginConstructor`.
Your constructor will create the plugin object, and then register any event
handlers you'd like with the bot. Finally, register your plugin with
`Register()`.

Some tips:
- The entire Slack API is available to you through `bot.API`, which is an
  instance
  of [`slack.Client`](https://godoc.org/github.com/nlopes/slack#Client). You
  have an RTM connection available through `bot.RTM`, which is the easiest way
  to send a message.
- Your constructor receives the argument `config map[string]interface{}`. The
  simplest way to use this is to
  use [`mapstructure`](https://github.com/mitchellh/mapstructure) to Unmarshal
  the data directly into your plugin struct.
- If you are creating many plugins, turn them into a package and group all of
  their registration into a single function for convenience.
  
See [plugins/respond.go](plugins/respond.go) for a simple example of a plugin
with state, configuration, and an event handler.

For more information,
see
[GoDoc (lib)](https://godoc.org/github.com/brenns10/slacksoc/lib),
[GoDoc (plugins)](https://godoc.org/github.com/brenns10/slacksoc/plugins).
