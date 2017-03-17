/*
This library contains a set of core plugins for the slacksoc bot. To register
these plugins, simply use the provided Register() function. Below is a list of
provided plugins (none of their implementations are publicly accessible).

Respond is a plugin which allows you to register triggers and one or more
responses to those triggers. In functionality, it is pretty much identical to
Slackbot. Its configuration is a list of "response" objects, each of which
contains "trigger" (a string), and "replies" (a list of strings).
*/
package plugins

import "github.com/brenns10/slacksoc/lib"

/*
To use the core plugins, simply call this function before calling lib.Run().
*/
func Register() {
	lib.Register("Respond", newRespond)
}
