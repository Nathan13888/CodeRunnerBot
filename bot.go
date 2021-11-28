package main

import (
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
)

var (
	Token string
)

func init() {
	Token = os.Getenv("TOKEN")
	if Token == "" {
		panic("empty token...")
	} else if len(Token) < 10 {
		panic("token seems too short...")
	}
	log.Debug().Msg("Using token: " + Token[:10])
}

func main() {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	zerolog.TimeFieldFormat = time.RFC3339
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout}
	multi := zerolog.MultiLevelWriter(consoleWriter)
	log.Logger = zerolog.New(multi).With().Timestamp().Logger()

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + Token)
	// TODO: set activity/status
	if err != nil {
		log.Error().Err(err).Msg("error creating Discord session")
		return
	}

	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(messageCreate)

	// Just like the ping pong example, we only care about receiving message
	// events in this example.
	dg.Identify.Intents = discordgo.IntentsGuildMessages

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		log.Error().Err(err).Msg("error opening connection")
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	log.Info().Msg("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}

// TODO: ...
var allowedLanguages = []string{
	"python", "bash",
}

// TODO: check for exec time
const MAX_EXECTIME time.Duration = 60 * time.Second

// TODO: yeet this for SLASH commands
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	lines := strings.Split(m.Content, "\n")

	re := regexp.MustCompile(" +")
	split := re.Split(strings.TrimSpace(lines[0]), -1)

	cmd := split[0]
	args := split[1:]

	// TODO: add support for other commands
	// TODO: add option to specify language option
	if cmd != "!run" {
		return
	}

	log.Debug().Strs("lines", lines).Str("cmd", cmd).Strs("args", args).Msg("Received Command")

	if len(args) == 0 {
		return
	}
	// proper format: back ticks, valid language
	lang := args[0]

	var code string
	if len(lines) > 3 {
		process := lines[1:]
		if !(process[0] == "```" && process[len(process)-1] == "```") {
			sendMessage(s, m.ChannelID, "Invalid input: you must put code in code blocks")
			return
		}
		for i := 1; i < len(process)-1; i++ {
			code += process[i] + "\n"
		}
		log.Debug().Str("code", code).Msg("Running Code")
	} else if len(lines) == 3 {
		sendMessage(s, m.ChannelID, "Invalid input: do you have any code?")
		return
	} else if len(lines) < 3 {
		sendMessage(s, m.ChannelID, "Invalid input: too few lines")
		return
	}

	// exec; check max exec time
	output, err := Exec(lang, code)
	if err != nil {
		sendMessage(s, m.ChannelID, fmt.Sprintf("Encountered Error: %v", err))
		return
	}
	// TODO: crop output...
	sendMessage(s, m.ChannelID, fmt.Sprintf("Received Output:\n```\n%s\n```\n", output))
}

func sendMessage(session *discordgo.Session, channelID string, message string) error {
	_, err := session.ChannelMessageSend(channelID, message)
	if err != nil {
		log.Error().Err(err).Msg("Received error while sending message")
		return err
	}
	return nil
}
