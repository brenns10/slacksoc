/*
A very, very simple command to run a Slack bot with only core plugins. If you
want external plugins, you will need to write your own small driver program.
Your best bet is to simply copy this package's code and add your additional
plugin registrations before running.

The command line interface is documented in lib.Run()
*/
package main

import "github.com/brenns10/slacksoc/lib"
import "github.com/brenns10/slacksoc/plugins"

func main() {
	plugins.Register()
	lib.Run()
}
