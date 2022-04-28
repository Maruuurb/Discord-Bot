package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

var (
	GuildID        = flag.String("guild", "", "Test guild ID. If not passed - bot registers commands globally")
	Token          string
	RemoveCommands = flag.Bool("rmcmd", false, "Remove all commands after shutdowning or not")
)
var s *discordgo.Session

func init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.Parse()
}

var (
	integerOptionMinValue = 1.0

	commands = []*discordgo.ApplicationCommand{
		{
			Name: "ip",
			// All commands and options must have a description
			// Commands/options without description will fail the registration
			// of the command.
			Description: "The command that sends the server IP",
		},
	}

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"ip": messageCreate,
	}
)

func main() {
	s, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})

	s.Identify.Intents = discordgo.IntentsGuildMessages
	err = s.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}
	log.Println("Adding commands...")
	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for i, v := range commands {
		cmd, err := s.ApplicationCommandCreate(s.State.User.ID, *GuildID, v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
		registeredCommands[i] = cmd
	}

	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	s.Close()
}

func messageCreate(s *discordgo.Session, m *discordgo.InteractionCreate) {

	resp, err := http.Get("http://myexternalip.com/raw")
	if err != nil {
		os.Stderr.WriteString(err.Error())
		os.Stderr.WriteString("\n")
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		bodyString := string(bodyBytes)
		s.InteractionRespond(m.Interaction, &discordgo.InteractionResponse{

			Type: discordgo.InteractionResponseChannelMessageWithSource,

			Data: &discordgo.InteractionResponseData{

				Flags: 1 << 6,

				Content: "IP майнкрафт сервера : " + bodyString + ":25565",
			},
		})
		//_, err = m.InteractionRespond(m.Interaction, "IP майнкрафт сервера : "+bodyString+":25565")
		if err != nil {
			// If an error occurred, we failed to send the message.
			//
			// It may occur either when we do not share a server with the
			// user (highly unlikely as we just received a message) or
			// the user disabled DM in their settings (more likely).
			fmt.Println("error sending DM message:", err)
			s.ChannelMessageSend(
				m.ChannelID,
				"Failed to send you a DM. "+
					"Did you disable DM in your privacy settings?",
			)
		}
		fmt.Println(bodyString)
	}

}
