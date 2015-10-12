Shigebot is a basic twitch IRC bot I wrote for 
[shigetora / cookiezi](http://www.twitch.tv/shigetora).

Features
================================================================================
- [x] Runs on Windows, Linux, Mac OS X, FreeBSD and every other os supported by 
      the Go programming language.
- [x] Connects and operates on multiple channels at once.
- [x] Simple text commands, separate for each channel and manageable by mods 
      through !cmdadd, !cmdedit and !cmdremove. The commands are saved and 
      restored on start-up.
- [x] Text commands can be restricted to moderators only through !modonly
- [x] Keeps track of the message count to avoid hitting the twitch message 
      rate cap.
- [x] Supports non-moderator accounts by randomizing messages and using a lower 
      message rate limit.
- [x] Uses a github account to commit and update the command list as a markdown  
      gist and links it instead of displaying a huge command list in chat.
- [x] Can be used as a library to develop your own bot.
- [x] Togglable case sensitivity.
- [x] Configurable ignore list to prevent conflicts with other bots on the 
      channel.

Usage
================================================================================
* Download the binaries from the 
  [releases section](https://github.com/Francesco149/shigebot/releases/).
* Obtain a twitch oauth token from [here](https://twitchapps.com/tmi/).
* Obtain a gist oauth toekn by using the provided gist-token utility.
* Enter the two oauth tokens in config.json and customize the other settings 
  to your liking.
* Run shigebot.

How to compile
================================================================================
* [Install go](https://golang.org/doc/install)
```
go get github.com/thoj/go-ircevent
go get -tags purego github.com/cznic/ql
go get github.com/MaximeD/gost
go get github.com/Francesco149/shigebot
go install github.com/Francesco149/shigebot/...
```
* Your binaries will be in GOPATH/bin

Known issues
================================================================================
* If the bot randomly disconnects after some time with an EOF error, configure 
  it to join its own channel (for example, if your bot is called mybot, you 
  should add #mybot to the channels in config.json). I have no idea why this 
  fixes it or why the random disconnects happen on some accounts, but I will 
  look into it.
* Recognizing the mods might take a while after the bot first joins a channel, 
  so mod-only commands will only start working a few minutes after the bot joins. 
  This is because twitch takes some time to send the operator modeset messages.

Using the shige package to make your own twitch bot
================================================================================
If you wish to customize the behaviour of the bot any further than what the 
settings allow, you can use the shige package as a library to make your own bot.

```
package main
import (
	"fmt"
	"strings"
	"github.com/Francesco149/shigebot/shige"
	"github.com/thoj/go-ircevent"
)

func main() {
	bot, err := shige.Init(
		"myuser", // twitch user
		"oauth:xxxxxxxxxxxxxxxxxxxxxxxxxxxxxx", // twitch oauth
		"xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx", // gist oauth
		[]string{
			"#mychannel1", 
			"#mychannel2", 
		}, 
		true, // moderator?
		false, // case sensitive
	)
	if err != nil {
		panic(err)
	}

	bot.AddCommand("test", func(c *shige.CommandData) {
		c.Channel.Privmsgf("You are %s and this is a custom built-in command." +
			" You called this command with these params: %v", c.Nick, c.Args)
	})

	bot.AddCommand("test2", func(c *shige.CommandData) {
		c.Channel.Privmsgf("This command should never execute")
	})

	bot.OnPrivmsg = func(event *irc.Event) bool {
		fmt.Println("Hi from the custom PRIVMSG handler, event=", event)

		if strings.HasPrefix(event.Message(), "!test2") {
			fmt.Println("Ignoring test2 command")
			return false
		}

		return true
	}

	bot.Run()
}
```
