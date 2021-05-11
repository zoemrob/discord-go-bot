package discordgobot

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/bwmarrin/discordgo"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

type BotCommand int

const (
	BotCommandHelp BotCommand = 1 << iota
	BotCommandMention
	BotCommandUnknown
	BotCommandInsult
	BotCommandMdnSearch
	BotCommandGoPkgSearch
	BotCommandGithubSearch
)

const (
	BotHelp         string = "help"
	BotInsult              = "insult me"
	BotMdnSearch           = "mdn"
	BotGoPkgSearch         = "go"
	BotGithubSearch        = "gh"
)

const sonaDevDiscordChannel string = "sona-dev"

const BotHelpMessage = `Commands:
		` + "`help`" + `
				Returns this dialogue, a list of commands.

		` + "`mdn <search terms>`" + `
				Searches Mozilla Developer Network and returns link to results

		` + "`go <search terms>`" + `
				Searches Go Pkg for relevant Go packages and returns the first 3 results

		` + "`gh <search terms>`" + `
				Searches Github for relevant git repositories and returns the first 3 results

		` + "`insult me`" + `
				If you are feeling too proud

`

type BotCommandInterface interface {
	exec() error
}

type BasicBotCommand struct {
	Command       BotCommand
	Session       *discordgo.Session
	MessageCreate *discordgo.MessageCreate
}

type BotCommandGeneric struct {
	BasicBotCommand
}

type BotCommandSearch struct {
	BasicBotCommand
	SearchTerms string
}

func (bcs BotCommandSearch) exec() error {
	if u, ok := bcs.generateURL(bcs.SearchTerms); ok {
		response, ok := bcs.getNResults(u, 3)
		if !ok {
			response = u.String()
		}

		err := bcs.respondToChannel(response)
		return err
	}
	return errors.New(fmt.Sprintf("There was a problem with %T.exec", bcs))
}

func (bcs BotCommandSearch) generateURL(search string) (url.URL, bool) {
	switch bcs.Command {
	case BotCommandMdnSearch:
		return buildUrlWithQuery(
			"developer.mozilla.org",
			"/en-US/search",
			"q",
			search,
		), true
	case BotCommandGoPkgSearch:
		return buildUrlWithQuery(
			"pkg.go.dev",
			"search",
			"q",
			search,
		), true
	case BotCommandGithubSearch:
		return buildUrlWithQuery(
			"github.com",
			"search",
			"q",
			search,
		), true
	default:
		return url.URL{}, false
	}
}

// BotCommandSearch takes a url to query and a number of results to return
func (bcs BotCommandSearch) getNResults(u url.URL, n int) (string, bool) {
	if bcs.Command == BotCommandMdnSearch {
		return "", false
	}
	res, err := http.Get(u.String())
	if err != nil {
		fmt.Printf("Error in %T.getNResults\nParams:%v\nError:%v", bcs, u, err)
	}

	defer closeReader(res.Body)

	if res.StatusCode != 200 {
		fmt.Printf("status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		fmt.Println(err)
	}

	results := make([]string, 0, n+1)
	switch bcs.Command {
	case BotCommandGoPkgSearch:
		results = append(results, "Here's what I found at pkg.go.dev...")
		doc.Find(".SearchResults .LegacySearchSnippet").Each(func(i int, s *goquery.Selection) {
			if i > 2 {
				return
			}

			var rs string
			if link, exists := s.Find("a").First().Attr("href"); exists {
				rs += formatUrlForMessage(u, link)
				rs += s.Find("p.SearchSnippet-synopsis").First().Text()
			}

			if rs != "" {
				results = append(results, rs)
			}
		})
	case BotCommandGithubSearch:
		results = append(results, "Here's what I found at the ol' Github...")
		doc.Find(".repo-list-item").Each(func(i int, s *goquery.Selection) {
			if i > 2 {
				return
			}

			var rs string
			if link, exists := s.Find("a").First().Attr("href"); exists {
				rs += formatUrlForMessage(u, link)
				raw := s.Find("p").First().Text()
				trimmed := strings.Replace(raw, "<em>", "", -1)
				trimmed = strings.Replace(trimmed, "</em>", "", -1)
				trimmed = strings.Replace(trimmed, "\n", "", -1)
				rs += trimmed
			}

			if rs != "" {
				results = append(results, rs)
			}
		})
	}

	if len(results) == 0 {
		results = append(results, "Seems like there wasn't anything to find. Sorry.")
	}

	return strings.Join(results, "\n\n"), true
}

func closeReader(Body io.ReadCloser) {
	err := Body.Close()
	if err != nil {
		fmt.Printf("Error closing %T\nError:%v", Body, err)
	}
}

func (bbc BasicBotCommand) respondToChannel(message string) error {
	_, err := bbc.Session.ChannelMessageSend(bbc.MessageCreate.ChannelID, message)
	return err
}

func (bcg BotCommandGeneric) exec() error {
	var err error
	switch bcg.Command {
	case BotCommandHelp:
		err = bcg.respondToChannel(BotHelpMessage)
	case BotCommandMention:
		err = bcg.respondToChannel("You called sire? Mention me with " + GetBotName(bcg.Session))
	case BotCommandInsult:
		err = bcg.respondToChannel(getRandomInsult())
	case BotCommandUnknown:
		err = bcg.respondToChannel("Um... I didn't get that. Try `" + GetBotName(bcg.Session) + " help`")
	}

	return err
}

func getRandomInsult() string {
	insults := []string{
		"You are basically a dog with a human body",
		"I wish you were just less of a turd",
		"Yeah, I mean, uh... you kinda suck",
		"Your mother was a hamster, and your father smelt of elderberries",
		"Sorry, I don't feel like it",
		"Um. No?",
		"You look like you got yeeted in the face by a lobster",
		"I didn't even know you were smart enough to speak!",
		"*sigh...* Why are you such a glutton for punishment?",
		"When you die in Warzone, your squad buys self-revive instead",
	}

	return insults[rand.Intn(len(insults)-1)]
}

func parseContent(s *discordgo.Session, m *discordgo.MessageCreate) BotCommandInterface {
	botID := s.State.User.ID
	botMention := fmt.Sprintf("<@!%v>", botID)
	content := m.Content

	regexPrefix := botMention + `\s*`

	botMentionRegex := regexp.MustCompile(regexPrefix + `$`)
	botHelpRegex := regexp.MustCompile(regexPrefix + BotHelp + `\s*`)
	botInsultRegex := regexp.MustCompile(regexPrefix + BotInsult + `\s*`)
	botMdnSearchRegex := regexp.MustCompile(regexPrefix + BotMdnSearch + `\s*`)
	botGoPkgSearchRegex := regexp.MustCompile(regexPrefix + BotGoPkgSearch + `\s*`)
	botGithubSearchRegex := regexp.MustCompile(regexPrefix + BotGithubSearch + `\s*`)

	switch {
	// @bot
	case botMentionRegex.MatchString(content):
		return &BotCommandGeneric{
			BasicBotCommand{BotCommandMention, s, m},
		}
	// @bot help
	case botHelpRegex.MatchString(content):
		return &BotCommandGeneric{
			BasicBotCommand{BotCommandHelp, s, m},
		}
	// @bot insult me
	case botInsultRegex.MatchString(content):
		return &BotCommandGeneric{
			BasicBotCommand{BotCommandInsult, s, m},
		}
	// @bot mdn <search terms>
	case botMdnSearchRegex.MatchString(content):
		return &BotCommandSearch{
			BasicBotCommand: BasicBotCommand{BotCommandMdnSearch, s, m},
			SearchTerms:     botMdnSearchRegex.ReplaceAllString(content, ""),
		}
	// @bot go <search terms>
	case botGoPkgSearchRegex.MatchString(content):
		return &BotCommandSearch{
			BasicBotCommand: BasicBotCommand{BotCommandGoPkgSearch, s, m},
			SearchTerms:     botGoPkgSearchRegex.ReplaceAllString(content, ""),
		}
	// @bot gh <search terms>
	case botGithubSearchRegex.MatchString(content):
		return &BotCommandSearch{
			BasicBotCommand: BasicBotCommand{BotCommandGithubSearch, s, m},
			SearchTerms:     botGithubSearchRegex.ReplaceAllString(content, ""),
		}
	// TODO: // BotCommandShadowUser, which accepts another user mention, and then go-bot mimics them in an annoying way
	// command is not found
	default:
		return &BotCommandGeneric{
			BasicBotCommand{BotCommandUnknown, s, m},
		}
	}
}

// DiscordContainer : Holds discordgo session and related info
type DiscordContainer struct {
	DiscordSession *discordgo.Session
	eventHandlers  []func()
	Channels       map[string]SimpleChannel
}

func NewDiscordContainer(botTokenEnv string) *DiscordContainer {
	discord, err := discordgo.New("Bot " + botTokenEnv)
	if err != nil {
		fmt.Println("Error initializing discord:", err)
	}

	// initial capacity for 10 eventHandlers
	return &DiscordContainer{
		DiscordSession: discord,
		eventHandlers:  make([]func(), 0, 10),
		Channels:       make(map[string]SimpleChannel),
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

// GetBotName returns the server bot name
func GetBotName(s *discordgo.Session) string {
	return "@" + s.State.User.Username
}

// SimpleChannel simplifies to only the data needed
type SimpleChannel struct {
	ID   string
	Name string
}

// formatUrlForMessage returns a properly formatted string for sending to client
func formatUrlForMessage(u url.URL, path string) string {
	return u.Scheme + "://" + u.Host + path + "\n\t"
}

// buildUrl shortens the repetition of adding query params
func buildUrlWithQuery(host, path, qparam, search string) url.URL {
	u := url.URL{
		Scheme: "https",
		Host:   host,
		Path:   path,
	}

	q := u.Query()
	q.Add(qparam, search)
	u.RawQuery = q.Encode()

	return u
}
