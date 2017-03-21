package lib

import "fmt"
import "io/ioutil"
import "log"
import "os"
import "gopkg.in/yaml.v2"
import "github.com/sirupsen/logrus"
import "github.com/nlopes/slack"

/*
PluginConfig is the type that is used to configure plugins. It is a very generic
map. The recommended route is to use the mapstructure library to parse it. This
works like normal unmarshalling in Go. For example:

    type myConfig struct {
        ApiKey    string
        RandomInt int
    }

You can parse a config into this struct like so:

    var cfg myConfig
    err := mapstructure.Decode(pluginConfig, &cfg)
    if err != nil {
        return err
    }
    // cfg now contains data from the PluginConfig

The data in the plugin config comes from the bot's YAML configuration file.
Specifically, everything in your plugin's configuration object that isn't the
name field gets included in the map. So, this configuration file would be
suitable for loading the above struct:

    token: foo
    plugins:
      - name: YourPluginName
        apiKey: bar
        randomInt: 4 # chosen by fair dice roll. guaranteed to be random

Notice that the initial characters are lower case. Mapstructure will only load
public fields, and it will load them from lower-cased fields in the map. Keep
this quirk in mind while working with the library. For more info, refer to the
mapstructure docs: https://godoc.org/github.com/mitchellh/mapstructure
*/
type PluginConfig map[string]interface{}

/*
The plugins section of the bot config file is just a list of these.
*/
type pluginConfigEntry struct {
	Name   string
	Config PluginConfig `yaml:",omitempty,inline"`
}

/*
This structure represents the configuration file used to configure the bot.
*/
type botConfig struct {
	Token   string
	Plugins []pluginConfigEntry
	// more configuration information will likely go here
}

/*
This loads a configuration file, sets any configuration values in the Bot, and
then initializes all plugins.

TODO: at some point this could become public. Some doc will be migrated from
Run() at that point.
*/
func (b *Bot) configure(filename string) error {
	var config botConfig
	file, err := os.Open(filename)
	if err != nil {
		return err
	}

	arr, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(arr, &config)
	if err != nil {
		return err
	}

	API := slack.New(config.Token)
	API.SetDebug(true)
	slack.SetLogger(log.New(b.Log.WriterLevel(logrus.DebugLevel), "", 0))
	b.API = API

	for _, entry := range config.Plugins {
		ctor, ok := plugins[entry.Name]
		if !ok {
			return fmt.Errorf("config error: plugin %s not found", entry.Name)
		}
		plugin := ctor(b, entry.Name, entry.Config)
		if plugin == nil {
			return fmt.Errorf("error loading plugin %s", entry.Name)
		}
		b.plugins[entry.Name] = plugin
	}
	return nil
}
