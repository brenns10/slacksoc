/*
This library contains a set of core plugins for the slacksoc bot. To register
these plugins, simply use the provided Register() function. Below is a list of
provided plugins (none of their implementations are publicly accessible).

Respond is a plugin which allows you to register triggers and one or more
responses to those triggers. In functionality, it is pretty much identical to
Slackbot. Its configuration is a list of "response" objects, each of which
contains "trigger" (a string), and "replies" (a list of strings).

Debug is a plugin which adds several "commands" for viewing internal state of
the bot and testing some capabilities. It allows you to view the list of
channels, users, and IDs. It also allows you to view the team metadata and test
the capability to post reactions.

Love is a CWRU Love client. It allows users to send each other love through a
simple command syntax. Its configuration object must contain two variables:
apiKey, which should be an API key generated from the admin section, and
baseUrl, which be the URL of the "api" endpoint, but without the trailing slash.
See golove/love package docs for details:
https://godoc.org/github.com/hacsoc/golove/love. See also the Yelp love repo
for even more details: https://github.com/Yelp/love
*/
package plugins

import "github.com/brenns10/slacksoc/lib"

/*
To use the core plugins, simply call this function before calling lib.Run().
*/
func Register() {
	lib.Register("Respond", newRespond)
	lib.Register("Debug", newDebug)
	lib.Register("Love", newLove)
	lib.Register("GitHub", newGitHub)
}
