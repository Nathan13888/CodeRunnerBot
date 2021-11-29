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

	dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandsHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})

	for _, cmd := range commands {
		_, err := dg.ApplicationCommandCreate(dg.State.User.ID, "", &cmd)
		if err != nil {
			log.Error().Err(err).Msg("error creating command")
			return
		}
	}

	// Wait here until CTRL-C or other term signal is received.
	log.Info().Msg("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, syscall.SIGTERM)
	<-sc

	for _, cmd := range commands {
		dg.ApplicationCommandDelete(dg.State.User.ID, "", cmd.ID)
	}

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

var (
	commands = []discordgo.ApplicationCommand{
		{
			Name: "Run Code",
			Type: discordgo.MessageApplicationCommand,
		},
	}
	commandsHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"Run Code": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			err := s.InteractionRespond(
				i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
				},
			)

			message := i.ApplicationCommandData().
				Resolved.Messages[i.ApplicationCommandData().TargetID]

			if err != nil {
				log.Error().Err(err).Msg("error responding to interaction")
				return
			}

			if !isCodeMessage(message) {
				s.FollowupMessageCreate(s.State.User.ID, i.Interaction, false, &discordgo.WebhookParams{
					Content: "Not a code message. Did you remember to wrap your code in backticks (```)?",
				})
				return
			}

			lang, code := getLanguageAndCodeFromMessage(message)

			log.Debug().
				Str("lang", lang).
				Msg("Language detected")

			if lang == "" {
				s.FollowupMessageCreate(s.State.User.ID, i.Interaction, false, &discordgo.WebhookParams{
					Content: "No language provided. Did you remember to put a language after the opening backticks? (```py)",
				})
				return
			}

			output, err := Exec(lang, code)
			if err != nil {
				s.FollowupMessageCreate(s.State.User.ID, i.Interaction, false, &discordgo.WebhookParams{
					Content: fmt.Sprintf("Encountered error while running code: %v", err),
				})
				return
			}

			for _, message := range splitOutput(output, 500) {
				s.FollowupMessageCreate(s.State.User.ID, i.Interaction, false, &discordgo.WebhookParams{
					Content: message,
				})
			}

			if err != nil {
				panic(err)
			}
		},
	}
)

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID || !isCodeMessage(m.Message) {
		return
	}
}

func isCodeMessage(m *discordgo.Message) bool {
	c := strings.Split(strings.ReplaceAll(m.Content, "\r\n", "\n"), "\n")
	if len(c) < 2 {
		return false
	}
	return c[0][:3] == "```" && c[len(c)-1] == "```"
}

func getLanguageAndCodeFromMessage(m *discordgo.Message) (string, string) {
	c := strings.Split(strings.ReplaceAll(m.Content, "\r\n", "\n"), "\n")

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
