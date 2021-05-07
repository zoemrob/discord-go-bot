package discordgobot

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
	discord := NewDiscordContainer(botTokenEnv)
	onReady := func (s *discordgo.Session, r *discordgo.Ready) {
		for _, guild := range r.Guilds {
			channels, err := s.GuildChannels(guild.ID)
			if err != nil {
				fmt.Println("failed to get guild channels", err)
			}

			for _, channel := range channels {
				discord.Channels[channel.Name] = SimpleChannel{
					ID: channel.ID,
					Name: channel.Name,
				}
			}
		}

		if sonaChan, ok := discord.Channels["sona-dev"]; ok {
			_, err := s.ChannelMessageSend(sonaChan.ID, "Hi! I'm finally here! Talk to me with @go-bot commands")
			if err != nil {
				fmt.Println("Failed to greet", err)
			}
		}

		fmt.Println("discord struct", discord.Channels)
	}
	// add handlers
	discord.AddHandler(onReady)
	discord.AddHandler(messageCreate)
	discord.Init()
	defer discord.Close()


	// TODO: create a channel / goroutine / select to receive messages
	// from os.STDIN
	//reader := bufio.NewReader(os.Stdin)
	//for {
	//	fmt.Print("-> ")
	//	text, _ := reader.ReadString('\n')
	//
	//	text = strings.Replace(text, "\n", "", -1)
	//
	//	fmt.Println()
	//}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	goCh := make(chan os.Signal, 1)
	signal.Notify(goCh, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-goCh
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
