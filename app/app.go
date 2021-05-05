package app

import (
	"flag"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"os"
	"os/signal"
	"syscall"
)

var botTokenEnv string

func init() {
	flag.StringVar(&botTokenEnv, "t", "", "Bot Token")
	flag.Parse()
}

func Start() {
	discord, err := discordgo.New("Bot " + botTokenEnv)
	if err != nil {
		fmt.Println("Error initializing discord:", err)
	}

	defer closeSession(discord)

	err = discord.Open()
	if err != nil {
		fmt.Println("Error connection to discord:", err)
	}

	rm := discord.AddHandler(messageCreate)
	defer rm()

	discord.Identify.Intents = discordgo.IntentsGuildMessages

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	goCh := make(chan os.Signal, 1)
	signal.Notify(goCh, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-goCh
}

func closeSession(discord *discordgo.Session) {
	err := discord.Close()
	if err != nil {
		fmt.Println("Failed to close discord Session", err)
	}
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself

	if m.Author.ID == s.State.User.ID {
		return
	}

	for _, u := range m.Mentions {
		if u.ID == s.State.User.ID {
			fmt.Println(m.ContentWithMentionsReplaced())

			cmd := parseContent(m.Content, s.State.User.ID)

			err := cmd.exec(s, m)
			if err != nil {
				fmt.Println(err)
			}
		}
	}

}
