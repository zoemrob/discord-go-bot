package discordgobot

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"strings"
)

type BotCommand int

const (
	BotCommandHelp BotCommand = 1 << iota
	BotCommandMention
	BotCommandUnknown
)

const (
	BotHelp string = "help"
	sonaDevDiscordChannel string = "sona-dev"
)


type BotCommandDetails struct {
	Command BotCommand
}

func (bc *BotCommandDetails) exec(s *discordgo.Session, m *discordgo.MessageCreate) error {
	var err error
	switch bc.Command {
	case BotCommandHelp:
		_, err = s.ChannelMessageSend(m.ChannelID, "This is just testing a general response for help.")
	case BotCommandMention:
		_, err = s.ChannelMessageSend(m.ChannelID, "You called sire?\n`Mention me with a command`")
	case BotCommandUnknown:
		_, err = s.ChannelMessageSend(m.ChannelID, "Um... I didn't get that. Try something different...?")
	}

	return err
}

func parseContent(content string, botID string) BotCommandDetails {
	switch {
	case strings.Contains(content, BotHelp):
		return BotCommandDetails{
			Command: BotCommandHelp,
		}
	case strings.Trim(content, "<@!> ") == botID:
		return BotCommandDetails{
			Command: BotCommandMention,
		}
	default:
		return BotCommandDetails{
			Command: BotCommandUnknown,
		}
	}
}

// DiscordContainer : Holds discordgo session and related info
type DiscordContainer struct {
	DiscordSession *discordgo.Session
	eventHandlers []func()
	Channels map[string]SimpleChannel
}

func NewDiscordContainer(botTokenEnv string) *DiscordContainer {
	discord, err := discordgo.New("Bot " + botTokenEnv)
	if err != nil {
		fmt.Println("Error initializing discord:", err)
	}

	// initial capacity for 10 eventHandlers
	return &DiscordContainer{
		DiscordSession:           discord,
		eventHandlers: make([]func(), 0, 10),
		Channels: make(map[string]SimpleChannel),
	}
}

func (d DiscordContainer) Init() {
	err := d.DiscordSession.Open()
	if err != nil {
		fmt.Println("Error connection to discord:", err)
	}

	d.listenToIntents()
}

// AddHandler *discordgo.Session.AddHandler wrapper, adds cleanup methods
// For use of always-on intent handlers
func (d DiscordContainer) AddHandler(handler interface{}) {
	rmHandler := d.DiscordSession.AddHandler(handler)
	d.eventHandlers = append(d.eventHandlers, rmHandler)
}

// removeHandlers cleanup of added always-on handlers
func (d DiscordContainer) removeHandlers() {
	for _, rh := range d.eventHandlers {
		rh()
	}
}

// listenToIntents configures discordGo.Intents to listen for
func (d DiscordContainer) listenToIntents() {
	d.DiscordSession.Identify.Intents = discordgo.IntentsGuildMessages
}

// Close closes discordgo.Session and removes event handlers
func (d DiscordContainer) Close() {
	d.removeHandlers()
	err := d.DiscordSession.Close()
	if err != nil {
		fmt.Println("Failed to close discord Session")
	}
}

// SendToSonaDevChannel allows sending to main dev channel
func (d DiscordContainer) SendToSonaDevChannel(message string) {
	if sonaChan, ok := d.Channels[sonaDevDiscordChannel]; ok {
		_, err := d.DiscordSession.ChannelMessageSend(sonaChan.ID, message)
		if err != nil {
			fmt.Println("Failed to send message:", message, "Error:", err)
		}
	}
}

// SimpleChannel simplifies to only the data needed
type SimpleChannel struct {
	ID string
	Name string
}