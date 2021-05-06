package discordgobot

import (
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
