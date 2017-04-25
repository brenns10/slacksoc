slacksoc [![GoDoc](https://godoc.org/github.com/brenns10/slacksoc?status.svg)](https://godoc.org/github.com/brenns10/slacksoc)
========

This is a new, and maybe improved Slack bot library for Go.

Features:
- Plugin based architecture, providing built-in documentation for users
- Configure plugins through a single YAML file
- Entire Slack API exposed to plugins
- Bot operations are thread-safe, allowing plugins to leverage concurrency

Included Plugins:
- Triggers / responses, like the original Slackbot
- Send [CWRU/Yelp Love](https://github.com/hacsoc/love)
- Create GitHub issues

Usage
-----

If all you would like to do is use the included plugins, usage is simple. First,
install the `slacksoc` binary:

    go get github.com/brenns10/slacksoc/slacksoc
    
Create a YAML configuration file - see [sample.yaml](sample.yaml) for an
example. Be sure that, at a minimum, the config contains your Slack API token,
and an entry with appropriate configuration for each plugin you want to use.
Finally, run the bot like this:

    slacksoc config.yaml

### Using External Plugins

If you would like to implement your own plugins, or use a third-party plugin (if
they ever exist), you will need to write a small amount of boilerplate code.
This is because Go has no support for dynamic module loading, and therefore
plugins need to be registered with the bot before they can be used. Here is a
sample:

```go
package main

import "github.com/brenns10/slacksoc/lib"
import "github.com/brenns10/slacksoc/plugins"

func main() {
    plugins.Register()
    lib.Register("plugin name", PluginConstructor)
    lib.Run()
}
```

### Developing Plugins

This slack bot implementation focuses on providing a simple and powerful
experience for plugin developers. Currently, the plugin interface is still in
flux, but it will soon stabilize. Documentation on plugin development can be
found in the [Wiki](https://github.com/brenns10/slacksoc/wiki).

- [GoDoc (lib)](https://godoc.org/github.com/brenns10/slacksoc/lib)
- [GoDoc (plugins)](https://godoc.org/github.com/brenns10/slacksoc/plugins)
