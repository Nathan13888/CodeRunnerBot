package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
)

var Token string

func init() {
	Token = os.Getenv("TOKEN")
	if Token == "" {
		panic("empty token...")
	} else if len(Token) < 10 {
		panic("token seems too short...")
	}

	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	zerolog.TimeFieldFormat = time.RFC3339
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout}
	multi := zerolog.MultiLevelWriter(consoleWriter)
	log.Logger = zerolog.New(multi).With().Timestamp().Logger()

	log.Debug().Msg("Using token: " + Token[:10])

	allowedLanguages := make([]string, len(languages))

	i := 0
	for k := range languages {
		allowedLanguages[i] = k
		i++
	}

	log.Debug().
		Strs("allowed_languages", allowedLanguages).
		Int("msg_char_lim", MSG_CHAR_LIM).
		Dur("max_exectime", MAX_EXECTIME).
		Msg("Configured Settings")
}

func main() {
	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + Token)
	// TODO: set activity/status
	if err != nil {
		log.Error().Err(err).Msg("Error creating Discord session")
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
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, syscall.SIGTERM)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}

var languages = map[string][]string{
	"awk":          {"awk"},
	"bash":         {"bash", "sh", "zsh", "ksh", "shell"},
	"befunge93":    {"befunge"},
	"brainfuck":    {"brainfuck", "bf"},
	"c":            {"c"},
	"c++":          {"cpp", "c++", "h"},
	"cjam":         {},
	"clojure":      {"clojure", "clj", "clojurescript", "cljs"},
	"cobol":        {"cobol"},
	"coffeescript": {"coffeescript", "coffee-script", "coffee"},
	"cow":          {},
	"crystal":      {"cr", "crystal"},
	"csharp":       {"csharp", "c#", "cs", "aspx-cs"},
	"csharp.net":   {},
	"d":            {"d"},
	"dart":         {"dart"},
	"dash":         {},
	"dragon":       {},
	"elixir":       {"elixir", "ex", "exs"},
	"emacs":        {"emacs-lisp", "elisp", "emacs"},
	"erlang":       {"erlang"},
	"file":         {},
	"forte":        {},
	"fortran":      {"fortran"},
	"freebasic":    {"basic"},
	"fsharp.net":   {},
	"fsi":          {"fsharp", "f#"},
	"go":           {"go", "golang"},
	"golfscript":   {},
	"groovy":       {"groovy"},
	"haskell":      {"haskell", "hs"},
	"husk":         {},
	"iverilog":     {"verilog", "v"},
	"japt":         {},
	"java":         {"java"},
	"javascript":   {"javascript", "js"},
	"jelly":        {},
	"julia":        {"julia", "jl"},
	"kotlin":       {"kotlin"},
	"lisp":         {"common-lisp", "cl", "lisp"},
	"llvm_ir":      {"llvm"},
	"lolcode":      {},
	"lua":          {"lua"},
	"nasm":         {"nasm"},
	"nasm64":       {},
	"nim":          {"nimrod", "nim"},
	"ocaml":        {"ocaml"},
	"octave":       {"octave"},
	"osabie":       {},
	"paradoc":      {},
	"pascal":       {"delphi", "pas", "pascal", "objectpascal"},
	"perl":         {"perl", "pl"},
	"php":          {"php", "php3", "php4", "php5"},
	"ponylang":     {"pony"},
	"powershell":   {"powershell", "pwsh", "posh", "ps1", "psm1"},
	"prolog":       {"prolog"},
	"pure":         {},
	"pyth":         {},
	"python":       {"python", "py", "sage", "python3", "py3"},
	"python2":      {"python2", "py2"},
	"racket":       {"racket", "rkt"},
	"raku":         {"perl6", "pl6", "raku"},
	"retina":       {},
	"rockstar":     {},
	"rscript":      {"rd"},
	"ruby":         {"ruby", "rb", "duby"},
	"rust":         {"rust", "rs"},
	"scala":        {"scala"},
	"sqlite3":      {"sql"},
	"swift":        {"swift"},
	"typescript":   {"typescript", "ts"},
	"basic":        {"basic"},
	"basic.net":    {},
	"vlang":        {},
	"vyxal":        {},
	"yeethon":      {},
	"zig":          {"zig"},
}

// TODO: check for exec time
const MAX_EXECTIME time.Duration = 60 * time.Second

// TODO: yeet this for SLASH commands
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID || !isCodeMessage(m.Message) {
		return
	}

	log.Debug().
		Str("author", m.Author.Username).
		Str("channel", m.ChannelID).
		Msg("Code Message")

	lang, code := getLanguageAndCodeFromMessage(m.Message)

	if lang == "" {
		return
	}

	log.Debug().
		Str("lang", lang).
		Msg("Language detected")

	output, err := Exec(lang, code)
	if err != nil {
		sendMessage(s, m.ChannelID, fmt.Sprintf("Encountered Error: %v", err))
		return
	}

	for _, message := range splitOutput(output, 500) {
		sendMessage(s, m.ChannelID, message)
	}
}

func isCodeMessage(m *discordgo.Message) bool {
	c := strings.Split(strings.ReplaceAll(m.Content, "\r\n", "\n"), "\n")
	if c[0][:3] != "```" || c[len(c)-1] != "```" || len(c) < 3 {
		return false
	}
	return true
}

func getLanguageAndCodeFromMessage(m *discordgo.Message) (string, string) {
	c := strings.Split(strings.ReplaceAll(m.Content, "\r\n", "\n"), "\n")
	if c[0][:3] != "```" || c[len(c)-1] != "```" || len(c) < 3 {
		return "", ""
	}

	for i, j := range languages {
		for _, k := range j {
			if strings.EqualFold(k, c[0][3:]) {
				return i, strings.Join(c[1:len(c)-1], "\n")
			}
		}
	}

	return "", ""
}

//const MSG_CHAR_LIM = 2000
const MSG_CHAR_LIM = 500 // to avoid getting rate limited

func splitOutput(output string, limit int) []string {
	var messages []string
	codeLimit := limit - 8

	for len(output) > limit {
		messages = append(messages, "```\n"+output[:codeLimit]+"\n```")
		output = output[limit:]
	}

	messages = append(messages, "```\n"+output+"\n```")

	return messages
}

func sendMessage(session *discordgo.Session, channelID string, message string) error {
	_, err := session.ChannelMessageSend(channelID, message)
	if err != nil {
		log.Error().Err(err).Msg("Received error while sending message")
		session.ChannelMessageSend(channelID, "Experienced error sending message...")
		return err
	}
	return nil
}
