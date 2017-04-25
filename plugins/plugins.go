/*
This library contains a set of core plugins for the slacksoc bot. To register
these plugins, simply use the provided Register() function. Below is a list of
provided plugins (none of their implementations are publicly accessible).

Respond Plugin

Respond is a plugin which allows you to register triggers and one or more
responses to those triggers. In functionality, it is pretty much a superset of
Slackbot, since it allows regular expressions and reactions. Here is its sample
configuration:

  - name: Respond
    # This is a list of objects that define a response. (As of now) only one
    # response can ever fire at a time--the first match.
    #
    # * trigger is a regular expression that uses the syntax of RE2:
    #   https://golang.org/s/re2syntax
    #   + For case insensitive matching, put (?i) at the front of your regex
    #   + The regular expression need only match within the string, not
    #     necessarily the whole string.
    #   + Use ^ and $ to match beginning/end of the string
    #   + If you have trouble with YAML messing up your regex, check this:
    #     http://stackoverflow.com/questions/6915756/uninterpreted-strings-in-yaml
    #
    # * replies - A list. One is randomly selected and sent to the channel on
    #   match.
    #
    # * reacts - A list. One is randomly selected and added to the message on
    #   match. Don't include colons. This will happen in addition to the reply.
    responses:
      - trigger: ^(yo|hey|hi|hello|sup),? slacksoc$
        replies: ["hello", "wassup", "yo"]
      - trigger: ^((good)?bye|adios),? slacksoc$
        replies: ["goodbye"]
      - trigger: (?i)i love you
        reacts: ["heart"]


Debug Plugin

Debug is a plugin which adds several "commands" for viewing internal state of
the bot and testing some capabilities. No configuration is required (beyond the
name of the plugin). Use `slacksoc help Debug` for more information on its
"functionality".

Love Plugin

Love is a CWRU Love client. It allows users to send each other love through a
simple command syntax. Its configuration object must contain two variables:
apiKey, which should be an API key generated from the admin section, and
baseUrl, which be the URL of the "api" endpoint, but without the trailing slash.
See golove/love package docs for details:
https://godoc.org/github.com/hacsoc/golove/love. See also the Yelp love repo
for even more details: https://github.com/Yelp/love

The ApiKey may be provided through the LOVE_API_KEY environment variable
instead.

Sample configuration:

  - name: Love
    # You'll need to get this from the Admin section of CWRU love. You can
    # provide this token via the config file, or the LOVE_API_KEY environment
    # variable.
    apiKey: LOVE_API_KEY
    baseUrl: https://cwrulove.appspot.com/api

GitHub Plugin

GitHub is a plugin which allows you to post a GitHub issue. See "slacksoc help
GitHub" for usage instructions. In its config object, you will need to set the
fields clientID, clientSecret, and accessToken. Or, you can specify them via
environment variables.

  - name: GitHub
    # clientID and clientSecret should be created by registering an app
    # https://github.com/settings/applications/new
    clientID: GITHUB_CLIENT_ID
    clientSecret: GITHUB_CLIENT_SECRET
    # accessToken is the authorization for your application to act on behalf
    # of a particular user. Log into this user on GitHub and go here:
    #
    # https://github.com/login/oauth/authorize?scope=repo&client_id=$CLIENT_ID
    #
    # Then, take the code appended to the URL and put it into this curl:
    #
    # curl -X POST -F 'client_id=$CLIENT_ID' \
    #              -F 'client_secret=$CLIENT_SECRET' \
    #              -F 'code=$CODE' \
    #      https://github.com/login/oauth/access_token
    #
    # The accessToken will be in the response.
    accessToken: GITHUB_ACCESS_TOKEN
    # You may provide the above GitHub tokens via the correspondingly named
    # environment variables instead (e.g. GITHUB_ACCESS_TOKEN)

RealName Plugin

RealName is a plugin that politely asks people to set their "Real Name" fields
so that other team members know their name. It sends a direct message to the
user when they join a particular channel with empty Real Name fields. Typically,
you'll want to run this on the #general channel so that having real names set is
a policy for the whole team. However, you could run it on another channel, and
it will still work.

  - name: RealName
    channel: general

Plugin Library Design

This plugin library demonstrates what I believe to be the best way to publish
plugins. Make the type name and constructor private. Collect all your plugins
into a single package, and then expose only a single public function to Register
them. Finally, document each plugin at the package level, with configuration
samples.

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
	lib.Register("RealName", newRealName)
}
