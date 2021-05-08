package discordgobot

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"os"
	"strings"
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

		discord.SendToSonaDevChannel("Hi! I'm finally here! Talk to me with @go-bot commands")
	}
	// add handlers
	discord.AddHandler(onReady)
	discord.AddHandler(messageCreate)
	discord.Init()
	defer discord.Close()

	stdinCh := make(chan string)
	go readStdin(stdinCh)

	fmt.Println("Bot is now running.  Type `exit` to quit, and type anything else to speak through me!")
	for {
		fmt.Print("-> ")
		text, ok := <- stdinCh
		if !ok {
			break
		}
		discord.SendToSonaDevChannel(text)
	}
}

func readStdin (stdinCh chan string) {
	reader := bufio.NewReader(os.Stdin)
	for {
		text, _ := reader.ReadString('\n')
		text = strings.Replace(text, "\n", "", -1)

		if text == "exit" {
			close(stdinCh)
			return
		}

		stdinCh <- text
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