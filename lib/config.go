package lib

import "encoding/gob"
import "fmt"
import "io/ioutil"
import "github.com/mitchellh/mapstructure"
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
	Token     string
	StateFile string
	SaveDelay int
	Plugins   []pluginConfigEntry
	// more configuration information will likely go here
}

func (b *Bot) initLoadState(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return nil // we will use empty state if it doesn't exist
	}

	dec := gob.NewDecoder(file)
	err = dec.Decode(&b.state)
	if err != nil {
		return err
	}
	return nil
}

/*
This loads a configuration file, sets any configuration values in the Bot, and
then initializes all plugins. To clarify, this configure() function is private
and it is for configuring the whole bot and loading the plugins.
*/
func (b *Bot) configure(filename string) error {
	var config botConfig

	// Unmarshal the bot config from YAML.
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

	// Get the bot state filename and unmarshal it.
	if config.StateFile == "" {
		config.StateFile = "state.gob"
	}
	b.stateDelay = config.SaveDelay
	b.stateFile = config.StateFile
	err = b.initLoadState(config.StateFile)
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

func contains(list []string, value string) bool {
	for _, item := range list {
		if item == value {
			return true
		}
	}
	return false
}

/*
Loads a plugin configuration into destination struct. Raise an error if the
configuration object did not contain a top-level key listed in "required". Also
raise an error if the configuration object contained any keys which were not
successfully loaded into the struct, since this is probably not intended.

Any error causes a crash, so the caller does not need to handle any errors.
*/
func (b *Bot) Configure(config PluginConfig, dest interface{}, required []string) {
	var metadata mapstructure.Metadata
	decoderConfig := &mapstructure.DecoderConfig{
		ErrorUnused: true,
		Metadata:    &metadata,
		Result:      dest,
	}
	decoder, err := mapstructure.NewDecoder(decoderConfig)
	if err != nil {
		b.Log.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("Error creating mapstructure decoder.")
	}
	err = decoder.Decode(config)
	if err != nil {
		b.Log.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("Error decoding plugin configuration.")
	}
	for _, key := range required {
		if !contains(metadata.Keys, key) {
			b.Log.WithFields(logrus.Fields{
				"key": key,
			}).Fatal("Plugin configuration missing required key.")
		}
	}
}
