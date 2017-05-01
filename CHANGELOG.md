Changelog
=========

All notable changes to this project will be documented in this file.
The format is based on [Keep a Changelog](http://keepachangelog.com/).
Versions are updated as follows:
- Major version incremented for changes which would break existing deployments
  or plugins without a config/code change.
- Minor version incremented for changes which are compatible with old
  deployments and plugins, but may add features which won't be available without
  a change.
- Patch version incremented for bugfixes, or changes which shouldn't impact old
  deployments or plugins.
  
Since this project is in Go, "release" is an arbitrary term, but it refers to a
git tag. When you `go get` this package, you get the master branch, but you are
welcome to pin particular git tags within your `vendor/` directory and update
manually.

## [Unreleased]

## [1.1.2] - 2017-04-30

- **Fixed:** incorrect bold markup in HotPotato plugin
- **Fixed:** incorrect time formatting in HotPotato plugin

## [1.1.1] - 2017-04-30

- **Fixed:** segfault in GitHub plugin due to new logging code

## [1.1.0] - 2017-04-30

- **Fixed:** Debug plugin's "state" command outputs as 0, not as {0}
- **Changed:** commas may be used when addressing slacksoc
- **Added:** HotPotato plugin
- **Fixed:** incorrect bold markup in GitHub plugin help text
- **Added:** "version" command in Debug plugin, to see bot version

## [1.0.0] - 2017-04-25

Initial release. This version is marked by the first deployment to the Hacker
Society slack: [hacsoc/slacksoc][].


[hacsoc/slack]: https://github.com/hacsoc/slacksoc
