package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/ColeJSmith19/destinyBot/models"

	"github.com/bwmarrin/discordgo"
)

//channelsToIgnore contains a list of channel IDs that should NOT be evaluated against
//512548489155182592 = AFK
//509279158732455936 = Destiny Chit Chat
var channelsToIgnore = []string{"512548489155182592", "509279158732455936"}

//destiny2AppID The application ID of the game Destiny 2
var destiny2AppID = "372438022647578634"

// Variables used for command line parameters
var (
	Token string
)

func init() {

	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.Parse()
}

func main() {

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(messageCreate)

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	text := strings.Split(m.Content, " ")

	if text[0] != "!TR3VR" {
		return
	}
	// s.ChannelMessageSend(m.ChannelID, "Hello Guardian :voidtitan1:")
	g, _ := s.Guild(m.GuildID)

	var gameUsers []models.GameUser

	for _, voiceStates := range g.VoiceStates {
		var vs discordgo.VoiceState = *voiceStates

		//prints out each instance of a player in a channel.
		fmt.Println(vs)

	}

	for _, p := range g.Presences {
		var pre discordgo.Presence = *p

		if pre.Game != nil && pre.Game.Name != "" {
			user, _ := s.User(pre.User.ID)
			var gameUser models.GameUser
			gameUser.Game = pre.Game.Name
			gameUser.UserName = user.String()
			gameUser.UserID = user.ID
			gameUser.IsPlayingDestiny2 = pre.Game.ApplicationID == destiny2AppID

			gameUsers = append(gameUsers, gameUser)
		}
	}
	b, e := json.Marshal(gameUsers)
	if e != nil {
		fmt.Println(e.Error())
	}
	f, e := os.Open("test.json")
	if e != nil {
		fmt.Println(e)
	}
	defer f.Close()

	e = ioutil.WriteFile("test.json", b, 0644)
	if e != nil {
		fmt.Println(e)
	}
}
