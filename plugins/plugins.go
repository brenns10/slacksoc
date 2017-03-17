/*
This library contains a set of core plugins for the slacksoc bot. To register
these plugins, simply use the provided Register() function.
*/
package plugins

import "github.com/brenns10/slacksoc/lib"

/*
To use the core plugins, simply call this function before calling lib.Run().
*/
func Register() {
	lib.Register("Respond", newRespond)
}
