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

	"github.com/bwmarrin/discordgo"
)

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

	var gameUsers []GameUser

	for _, p := range g.Presences {
		var pre discordgo.Presence = *p

		if pre.Game != nil && pre.Game.Name != "" {
			user, _ := s.User(pre.User.ID)

			var gameUser GameUser
			gameUser.Game = pre.Game.Name
			gameUser.UserName = user.String()
			gameUser.UserID = user.ID

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

//GameUser holds a game name, a username and a user id
type GameUser struct {
	Game     string `json:"game"`
	UserName string `json:"username"`
	UserID   string `json:"userid"`
}

// func getUserByID(s *discordgo.Session, m *discordgo.MessageCreate, uid string) discordgo.User {

// }
