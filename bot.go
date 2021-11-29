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
	TOKEN        string
	PISTON_URL   string
	DOTENV       string
	GUILD_ID     string
	BuildVersion string = "unknown"
	BuildTime    string = "unknown"
	GOOS         string = runtime.GOOS
	ARCH         string = runtime.GOARCH
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
	DOTENV = os.Getenv("DOTENV")
	if len(DOTENV) == 0 {
		log.Info().
			Msg("Environment variable DOTENV not found, using default .env file.")
		DOTENV = ".env"
	}

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal().
			Err(err).
			Str("env_file", DOTENV).
			Msg("Error loading environment file.")
	}

	TOKEN = os.Getenv("TOKEN")
	if TOKEN == "" {
		log.Fatal().
			Msg("TOKEN not found in .env file.")
	}

	PISTON_URL = os.Getenv("PISTON_URL")
	if PISTON_URL == "" {
		log.Info().
			Msg("PISTON_URL not found in .env file, using default API endpoint.")
		PISTON_URL = "https://emkc.org/api/v2/piston/"
	}

	GUILD_ID = os.Getenv("GUILD_ID")
	if GUILD_ID == "" {
		log.Info().
			Msg("GUILD_ID not found in .env file, registering commands globally.")
	}

	// Log the languages and environment.
	i := 0
	for l := range languageMappings {
		languages[i] = l
		i++
	}

	log.Debug().
		Strs("languages", languages).
		Str("env_file", DOTENV).
		Str("token", TOKEN[:10]+strings.Repeat("*", len(TOKEN)-10)).
		Str("piston_url", PISTON_URL).
		Str("guild_id", GUILD_ID).
		Msg("Configured settings.")
}

func main() {
	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + TOKEN)
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("Error creating Discord session.")
	}

	// Add a handler for the bot's status.
	dg.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		s.UpdateListeningStatus("/run")
	})

	// Add guild messages intent.
	dg.Identify.Intents = discordgo.IntentsGuildMessages

	// Add handler to run the corresponding function when a command is run.
	dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandsHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)

			log.Debug().
				Str("command",
					i.ApplicationCommandData().Name).
				Str("user_id",
					i.Member.User.ID).
				Str("channel_id",
					i.ChannelID).
				Str("guild_id",
					i.GuildID).
				Msg("Command recieved.")
		}
	})

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("Error opening Disord connection.")
	}

	// Add all the application commands in the commands slice.
	for _, cmd := range commands {
		_, err := dg.ApplicationCommandCreate(dg.State.User.ID, GUILD_ID, &cmd)
		if err != nil {
			log.Fatal().
				Err(err).
				Str("command", cmd.Name).
				Msg("Error creating command.")
		}
	}

	// Wait here until CTRL-C or other term signal is received.
	log.Info().Msg("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	for _, cmd := range commands {
		err := dg.ApplicationCommandDelete(dg.State.User.ID, GUILD_ID, cmd.ID)
		if err != nil {
			log.Fatal().
				Err(err).
				Str("command", cmd.Name).
				Msg("Error deleting command.")
		}
	}

	// Cleanly close the Discord session.
	dg.Close()
}

var languages = make([]string, len(languageMappings))

// Array of all available languages as well as their markdown codes.
var languageMappings = map[string][]string{
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
		{
			Name:        "run",
			Description: "Runs code in a language. Run this command in a reply to a code message.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "language",
					Description: "The language to run the code in.",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    false,
				},
			},
		},
		{
			Name:        "help",
			Description: "Shows the help message.",
		},
		{
			Name:        "build_info",
			Description: "Shows the build info for the bot.",
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
				log.Error().
					Err(err).
					Msg("Error responding to interaction.")
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
					Content: "No language provided. Did you remember to put a valid language after the opening backticks? (e.g. ```py)",
				})

				if err != nil {
					log.Error().
						Err(err).
						Msg("Error sending followup message.")
				}

				return
			}

			// Get output of executed code.
			output, err := Exec(lang, "", code)

			if err != nil {
				log.Error().
					Err(err).
					Msg("Error executing code.")

				_, err := s.FollowupMessageCreate(s.State.User.ID, i.Interaction, false, &discordgo.WebhookParams{
					Content: fmt.Sprintf("Error executing code.```\n%v\n```", err),
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
		"run": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			// Send deferred message, telling the user that a response is coming shortly.
			err := s.InteractionRespond(
				i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
				},
			)

			if err != nil {
				log.Error().
					Err(err).
					Msg("Error responding to interaction.")
				return
			}

			// Get last 10 messages in channel.
			messages, err := s.ChannelMessages(i.ChannelID, 10, "", "", "")

			if err != nil {
				log.Error().
					Err(err).
					Msg("Error getting messages in channel.")

				_, err := s.FollowupMessageCreate(s.State.User.ID, i.Interaction, false, &discordgo.WebhookParams{
					Content: "Error getting messages in channel.",
				})

				if err != nil {
					log.Error().
						Err(err).
						Msg("Error sending followup message.")
				}

				return
			}

			// Check if any of those messages is a code message.
			var message *discordgo.Message

			for _, m := range messages {
				if isCodeMessage(m) {
					message = m
					break
				}
			}

			if (message == &discordgo.Message{}) {
				_, err := s.FollowupMessageCreate(s.State.User.ID, i.Interaction, false, &discordgo.WebhookParams{
					Content: "No code messages found in the last 10 messages. Did you remember to wrap your code in backticks (```)?",
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

			if len(i.ApplicationCommandData().Options) > 0 {
				lang = i.ApplicationCommandData().Options[0].StringValue()

				log.Debug().
					Str("language", lang).
					Msg("Language found from options.")

				if !stringInSlice(lang, languages) {
					_, err := s.FollowupMessageCreate(s.State.User.ID, i.Interaction, false, &discordgo.WebhookParams{
						Content: fmt.Sprintf("Language %v is not supported. Supported languages are: %v", lang, languages),
					})

					if err != nil {
						log.Error().
							Err(err).
							Msg("Error sending followup message.")
					}

					return
				}
			}

			if lang != "" {
				log.Debug().
					Str("language", lang).
					Msg("Language found from message.")
			} else {
				log.Debug().
					Msg("No language found from message.")

				_, err := s.FollowupMessageCreate(s.State.User.ID, i.Interaction, false, &discordgo.WebhookParams{
					Content: "No language provided. Did you remember to put a valid language after the opening backticks? (e.g. ```py)",
				})

				if err != nil {
					log.Error().
						Err(err).
						Msg("Error sending followup message.")
				}

				return
			}

			// Get output of executed code.
			output, err := Exec(lang, "", code)

			if err != nil {
				log.Error().
					Err(err).
					Msg("Error executing code.")

				_, err := s.FollowupMessageCreate(s.State.User.ID, i.Interaction, false, &discordgo.WebhookParams{
					Content: fmt.Sprintf("Error executing code.```\n%v\n```", err),
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
		"help": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			err := s.InteractionRespond(
				i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Embeds: []*discordgo.MessageEmbed{
							{
								Title: "Help",
								Fields: []*discordgo.MessageEmbedField{
									{
										Name:  "Run Code",
										Value: "Right click on any message to run it, if that message is a code message.",
									},
									{
										Name: "`/run [language]`",
										Value: strings.Join([]string{
											"Looks for a code message in the last 10 messages in the channel and executes it.",
											"If the language is not specified, it will try to detect the language from the language specified after the backticks (e.g. \\`\\`\\`py).",
										}, "\n"),
									},
									{
										Name:  "Supported Languages",
										Value: strings.Join(languages, ", "),
									},
								},
							},
						},
					},
				},
			)

			if err != nil {
				log.Error().
					Err(err).
					Msg("Error responding to interaction.")
				return
			}
		},
		"build_info": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			err := s.InteractionRespond(
				i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Embeds: []*discordgo.MessageEmbed{
							{
								Title: "Build Info",
								Fields: []*discordgo.MessageEmbedField{
									{
										Name:  "Version",
										Value: BuildVersion,
									},
									{
										Name:  "Time",
										Value: BuildTime,
									},
									{
										Name:  "Operating System",
										Value: GOOS,
									},
									{
										Name:  "Architecture",
										Value: ARCH,
									},
								},
							},
						},
					},
				},
			)

			if err != nil {
				log.Error().
					Err(err).
					Msg("Error responding to interaction.")
				return
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
	for i, j := range languageMappings {
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

func stringInSlice(s string, a []string) bool {
	for _, i := range a {
		if i == s {
			return true
		}
	}
	return false
}
