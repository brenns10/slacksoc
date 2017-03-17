package main

import "github.com/brenns10/slacksoc/lib"
import "github.com/brenns10/slacksoc/plugins"

func main() {
	plugins.Register()
	lib.Run()
}
