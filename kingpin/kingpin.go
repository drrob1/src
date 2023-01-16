package main

import (
	"fmt"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
)

func start1() { // main for first example
	var verbose = kingpin.Flag("verbose", "Enable verbose mode.").Short('v').Bool()
	var timeout = kingpin.Flag("timeout", "Set timeout value, default is 5s").Default("5s").Envar("TEST_TIMEOUT").Short('t').Duration()
	var ip = kingpin.Arg("ip", "IP address to ping.").Required().IP()
	var count = kingpin.Arg("count", "Count of packets to send for ping").Int()

	kingpin.Version("0.0.1")
	kingpin.Parse()
	fmt.Printf(" Would ping %s with timeout %s and count %d\n", *ip, *timeout, *count)
	if *verbose {
		fmt.Printf(" Verbose mode is on.")
	}

	// -------------------------------------------------------------------------------------------------------------
	// next example
	// -------------------------------------------------------------------------------------------------------------

	var app = kingpin.New("chat", "A CLI chat app")
	var debug = kingpin.Flag("debug", "Enable debug mode").Short('d').Bool()
	var serverIP = kingpin.Flag("server", "Server address").Default("127.0.0.1").IPVar

	var register = app.Command("register", "Register a new user.")
	var registerNick = register.Arg("nick", "Nickname for user.").Required().String()
	var registerName = register.Arg("name", "Name of user.").Required().String()

	var post = app.Command("post", "Post a message to a channel.")
	var postImage = post.Flag("image", "Image to post.").File()
	var postChannel = post.Arg("channel", "Channel to which to post.")
	var postText = post.Arg("text", "Text to post.")

	if *debug {
		fmt.Printf(" server IP: %s\n", *ip)
		serverIP(ip)
	}

	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case register.FullCommand():
		fmt.Printf(" Full Registered name: %s, and nickname: %s\n", *registerName, *registerNick)

	case post.FullCommand():
		if *postImage != nil {

		}
		text := *postText.String()
		text = text + " "
		fmt.Printf(" Text: %s, channel: %s\n", text, *postChannel.String())
	}
}

// -------------------------------------------------------------------------------------------------------------
// next example
// -------------------------------------------------------------------------------------------------------------
// modular example

type LsCmd struct {
	All bool
}

func (l *LsCmd) run(c *kingpin.ParseContext) error {
	fmt.Printf("all=%#v\n", l.All)
	return nil
}

func configureLsCmd(app *kingpin.Application) {
	c := &LsCmd{}
	ls := app.Command("ls", "List files.").Action(c.run)
	ls.Flag("all", "List all files.").Short('a').BoolVar(&c.All)

}

func start2() { // main for next example called modular
	app := kingpin.New("modular", "My modular app.")
	configureLsCmd(app)
	kingpin.MustParse(app.Parse(os.Args[1:]))
}

func main() { // I can only have one main()
	start1()
	start2()
}

