package lib

import "fmt"
import "io/ioutil"
import "github.com/sirupsen/logrus"
import "github.com/mitchellh/mapstructure"
import "os"
import "gopkg.in/yaml.v2"

/*
PluginConfig is the type that is used to configure plugins. It is a map which
can be converted fairly easily to any struct type using mapstructure (as you
would for YAML or JSON).
*/
type PluginConfig map[string]interface{}

/*
The plugins section of the bot config file is just a list of these.
*/
type pluginConfigEntry struct {
	Name   string
	Config PluginConfig `yaml:",omitempty"`
}

/*
This structure represents the configuration file used to configure the bot.
*/
type botConfig struct {
	Plugins []pluginConfigEntry
	// more configuration information will likely go here
}

/*
This loads a configuration file, sets any configuration values in the Bot, and
then initializes all plugins. To clarify, this configure() function is private
and it is for configuring the whole bot and loading the plugins.
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
