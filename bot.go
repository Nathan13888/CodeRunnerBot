package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
)

var (
	Token        string
	BuildVersion string = "unknown"
	BuildTime    string = "unknown"
	GOOS         string = runtime.GOOS
	ARCH         string = runtime.GOARCH
	PISTON_URL   string
)

func init() {

	// Initialize zerolog
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	zerolog.TimeFieldFormat = time.RFC3339
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout}
	multi := zerolog.MultiLevelWriter(consoleWriter)
	log.Logger = zerolog.New(multi).With().Timestamp().Logger()

	// Load environment from .env.
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("Error loading .env file.")
	}

	Token = os.Getenv("TOKEN")
	if Token == "" {
		log.Fatal().
			Msg("TOKEN not found in .env file.")
	}

	PISTON_URL = os.Getenv("PISTON_URL")
	if PISTON_URL == "" {
		log.Info().
			Msg("PISTON_URL not found in .env file, using default API endpoint.")
		PISTON_URL = "https://emkc.org/api/v2/piston/"
	}
}

func main() {
	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("Error creating Discord session.")
	}

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("Error opening Disord connection.")
	}

	// Add handler to run the corresponding function when a command is run.
	dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandsHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)

			log.Debug().
				Str("message_id",
					i.ApplicationCommandData().
						Resolved.
						Messages[i.ApplicationCommandData().TargetID].
						ID).
				Str("user_id",
					i.ApplicationCommandData().
						Resolved.
						Messages[i.ApplicationCommandData().TargetID].
						Author.ID).
				Str("channel_id",
					i.ApplicationCommandData().
						Resolved.
						Messages[i.ApplicationCommandData().TargetID].
						ChannelID).
				Msg(fmt.Sprintf("Command recieved: \"%s\"", i.ApplicationCommandData().Name))
		}
	})

	// Add all the application commands in the commands slice.
	for _, cmd := range commands {
		_, err := dg.ApplicationCommandCreate(dg.State.User.ID, "914901044595658782", &cmd)
		if err != nil {
			log.Fatal().
				Err(err).
				Msg(fmt.Sprintf("Error creating command: \"%s\"", cmd.Name))
		}
	}

	// Wait here until CTRL-C or other term signal is received.
	log.Info().Msg("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Cleanly close the Discord session.
	dg.Close()
}

// Array of all available languages as well as their markdown codes.
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
	// Commands slice of all available commands.
	commands = []discordgo.ApplicationCommand{
		{
			Name: "Run Code",
			Type: discordgo.MessageApplicationCommand,
		},
	}

	// CommandsHandlers map of all available commands and their corresponding handlers.
	commandsHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"Run Code": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			// Send deferred message, telling the user that a response is coming shortly.
			err := s.InteractionRespond(
				i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
				},
			)

			if err != nil {
				log.Error().Err(err).Msg("Error responding to interaction.")
				return
			}

			// Get message from ApplicationCommandData.
			message := i.ApplicationCommandData().
				Resolved.
				Messages[i.ApplicationCommandData().TargetID]

			// Check if the message is a code message.
			if !isCodeMessage(message) {
				_, err := s.FollowupMessageCreate(s.State.User.ID, i.Interaction, false, &discordgo.WebhookParams{
					Content: "Message is not a code message. Did you remember to wrap your code in backticks (```)?",
				})

				if err != nil {
					log.Error().
						Err(err).
						Msg("Error sending followup message.")
				}

				return
			}

			// Get the language and code from the message.
			lang, code := getLanguageAndCodeFromMessage(message)

			if lang != "" {
				log.Debug().
					Str("language", lang).
					Msg("Language found from message.")
			} else {
				log.Debug().
					Msg("No language found from message.")

				_, err := s.FollowupMessageCreate(s.State.User.ID, i.Interaction, false, &discordgo.WebhookParams{
					Content: "No language provided. Did you remember to put a valid language after the opening backticks? (```py)",
				})

				if err != nil {
					log.Error().
						Err(err).
						Msg("Error sending followup message.")
				}

				return
			}

			// Get output of executed code.
			output, err := Exec(lang, "latest", code)

			if err != nil {
				log.Error().
					Err(err).
					Msg("Error executing code.")

				_, err := s.FollowupMessageCreate(s.State.User.ID, i.Interaction, false, &discordgo.WebhookParams{
					Content: fmt.Sprintf("Error executing code: `%v`", err),
				})

				if err != nil {
					log.Error().
						Err(err).
						Msg("Error sending followup message.")
				}

				return
			}

			// Split code output into chunks of 500 characters and send them as followup messages.
			for _, message := range splitOutput(output, 500) {
				_, err := s.FollowupMessageCreate(s.State.User.ID, i.Interaction, false, &discordgo.WebhookParams{
					Content: message,
				})

				if err != nil {
					log.Error().
						Err(err).
						Msg("Error sending followup message.")
				}
			}
		},
	}
)

func isCodeMessage(m *discordgo.Message) bool {
	// Split on newlines.
	c := strings.Split(strings.ReplaceAll(m.Content, "\r\n", "\n"), "\n")

	// Check if the number of lines is greater than 1.
	if len(c) < 2 {
		return false
	}

	// Check if the first line starts with 3 backticks, and the last line is 3 backticks.
	return c[0][:3] == "```" && c[len(c)-1] == "```"
}

func getLanguageAndCodeFromMessage(m *discordgo.Message) (string, string) {
	// Split on newlines.
	c := strings.Split(strings.ReplaceAll(m.Content, "\r\n", "\n"), "\n")

	// Get language from first line.
	for i, j := range languages {
		for _, k := range j {
			// Check if the language in the first line is a valid language.
			if strings.EqualFold(k, c[0][3:]) {
				return i, strings.Join(c[1:len(c)-1], "\n")
			}
		}
	}

	return "", ""
}

func splitOutput(output string, limit int) []string {
	// Initialize slice of messages.
	var messages []string

	// Remove the 6 backticks and 2 newlines from the limit.
	codeLimit := limit - 8

	// While the output is larger than the limit, add limit-sized chunks to the slice.
	for len(output) > limit {
		messages = append(messages, "```\n"+output[:codeLimit]+"\n```")
		output = output[limit:]
	}

	// Add the remaining output to the slice.
	messages = append(messages, "```\n"+output+"\n```")

	return messages
}
