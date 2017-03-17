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
    
Create a YAML configuration file - see [etc/sample.yaml][etc/sample.yaml] for an
example. Also, get a bot API token from your Slack. Then, you can run as
follows:

    slacksoc config.yaml API_KEY

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

For more information,
see
[GoDoc (lib)](https://godoc.org/github.com/brenns10/slacksoc/lib),
[GoDoc (plugins)](https://godoc.org/github.com/brenns10/slacksoc/plugins).
