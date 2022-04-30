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
	AppID          = flag.String("app", "", "Application ID")
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
			Name:        "ip",
			Description: "The command that sends the server IP",
			Type:        discordgo.MessageApplicationCommand,
		},
		{
			Name:        "win",
			Description: "The command that sends the winner day",
			Type:        discordgo.MessageApplicationCommand,
		},
		{
			Name:        "rickroll",
			Description: "The command that sends the winner day",
			Type:        discordgo.MessageApplicationCommand,
		},
	}

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"ip":       messageIPCreate,
		"win":      messageWinCreate,
		"rickroll": rickroll,
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

func messageIPCreate(s *discordgo.Session, m *discordgo.InteractionCreate) {

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
		if err != nil {

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
func messageWinCreate(s *discordgo.Session, m *discordgo.InteractionCreate) {

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
		if err != nil {

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
func rickroll(s *discordgo.Session, i *discordgo.InteractionCreate) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Operation rickroll has begun",
			Flags:   1 << 6,
		},
	})
	if err != nil {
		panic(err)
	}

	ch, err := s.UserChannelCreate(
		i.ApplicationCommandData().TargetID,
	)
	if err != nil {
		_, err = s.FollowupMessageCreate(*AppID, i.Interaction, true, &discordgo.WebhookParams{
			Content: fmt.Sprintf("Mission failed. Cannot send a message to this user: %q", err.Error()),
			Flags:   1 << 6,
		})
		if err != nil {
			panic(err)
		}
	}
	_, err = s.ChannelMessageSend(
		ch.ID,
		fmt.Sprintf("%s sent you this: https://youtu.be/dQw4w9WgXcQ", i.Member.Mention()),
	)
	if err != nil {
		panic(err)
	}
}
