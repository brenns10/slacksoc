package lib

import "fmt"
import "io/ioutil"
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
then initializes all plugins.
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
