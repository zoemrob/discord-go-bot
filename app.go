package discordgobot

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

var botTokenEnv string
var disablePassthrough bool

func init() {
	flag.StringVar(&botTokenEnv, "t", "", "Bot Token")
	flag.BoolVar(&disablePassthrough, "d", false, "Disable passthrough")
	flag.Parse()
}

func Start() {
	discord := NewDiscordContainer(botTokenEnv)
	onReady := func(s *discordgo.Session, r *discordgo.Ready) {
		for _, guild := range r.Guilds {
			channels, err := s.GuildChannels(guild.ID)
			if err != nil {
				fmt.Println("failed to get guild channels", err)
			}

			for _, channel := range channels {
				discord.Channels[channel.Name] = SimpleChannel{
					ID:   channel.ID,
					Name: channel.Name,
				}
			}
		}
		// disable for testing
		discord.SendToSonaDevChannel("Hi! I'm finally here! Talk to me with " + GetBotName(discord.DiscordSession) + " commands")
	}
	// add handlers
	discord.AddHandler(onReady)
	discord.AddHandler(messageCreate)
	discord.Init()
	defer discord.Close()

	stdinCh := make(chan string)
	sigCh := make(chan os.Signal, 1)
	// uniquely non blocking, signal.Notify does not block even unbuffered
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	go readStdin(stdinCh, sigCh)

	fmt.Println("Bot is now running.  Type `exit` or `^C` to quit, and type anything else to speak through me!")
	fmt.Print("-> ")
	for text := range stdinCh {
		if !disablePassthrough {
			discord.SendToSonaDevChannel(text)
		}
		fmt.Print("-> ")
	}
	fmt.Println("closing...")
}

func readStdin(stdinCh chan<- string, sigCh <-chan os.Signal) {
	// include buffer of one to handle ^C os signal
	rCh := make(chan string, 1)
	go func() {
		reader := bufio.NewReader(os.Stdin)
		for {
			// blocks until delimiter, therefore cannot reach select
			text, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println("err in goroute:", err)
				close(rCh)
				return
			}
			text = strings.Replace(text, "\n", "", -1)
			rCh <- text
		}
	}()

	for {
		select {
		case text, ok := <-rCh:
			if !ok || text == "exit" {
				close(stdinCh)
				return
			}
			stdinCh <- text
		case <-sigCh:
			close(stdinCh)
		}
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

			cmd := parseContent(s, m)

			err := cmd.exec()
			if err != nil {
				fmt.Println(err)
			}
			return
		}
	}

	// TODO: add some funny responses to random things people say
	// example, 1/3rd chance that if someone sends a message only containing an emoji
	// respond with "don't you :emoji: me"
}
