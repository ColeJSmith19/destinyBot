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

//destiny2AppID application ID of the game Destiny 2
var destiny2AppID = "372438022647578634"

//botTestingChannelID ID of the bot-testing channel in discord
const botTestingChannelID = "739454920473968660"

//kingsGambitGuildID ID of our Clan, King's Gambit
const kingsGambitGuildID = "377181933614006282"

//kingsGambitClanMemberID ID of the Role given to players in the clan
const kingsGambitClanMemberRoleID = "409209843501629450"

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
	if m.ChannelID != "739454920473968660" {
		return
	}

	g, _ := s.Guild(m.GuildID)

	fmt.Println(getRolesForGuild(*g, *s, m.GuildID))

	// fmt.Println(s.GuildMemberRoleAdd(m.GuildID, "178671436435554305", "739222006880534639"))

	var gameUsers []models.GameUser

	for _, voiceStates := range g.VoiceStates {
		var vs discordgo.VoiceState = *voiceStates

		//prints out each instance of a player in a channel.
		fmt.Println(vs) //REMOVE

		//skip any users in channelsToIgnore const
		if stringInList(channelsToIgnore, vs.ChannelID) {
			continue
		}

		//skip any users NOT in the guild King's Gambit
		if vs.GuildID != kingsGambitGuildID {
			continue
		}

		//skip any users WITHOUT the role King's Gambit Clan Member
		if !userIsInClan(*s, m.GuildID, vs.UserID) {
			continue
		}

		gameUser := getPresenceForUser(*g, *s, vs.UserID)

		if gameUser.IsEmpty() {
			continue
		}

		gameUser.ChannelID = vs.ChannelID
		gameUsers = append(gameUsers, gameUser)
	}

	b, e := json.Marshal(gameUsers)
	// mes, e := s.ChannelMessageSend(botTestingChannelID, string(b))

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

	if len(gameUsers) < 2 {
		return
	}

	//loop over all gameUser instances and check to see if users are in the same Channel based on channelID
	//if so, check to see if they are playing Destiny 2
	//if so, check to see if other players in the same channel are playing destiny 2
	//if so, check to see which users are "unseen" for the month and upgrade their role to "seen"

}

func getPresenceForUser(g discordgo.Guild, s discordgo.Session, uid string) models.GameUser {
	var gameUser models.GameUser

	for _, p := range g.Presences {
		var pre discordgo.Presence = *p

		if pre.Game != nil && pre.Game.Name != "" && pre.User.ID == uid {
			user, _ := s.User(pre.User.ID)
			gameUser.Game = pre.Game.Name
			gameUser.UserName = user.String()
			gameUser.UserID = user.ID
			gameUser.IsPlayingDestiny2 = pre.Game.ApplicationID == destiny2AppID
			break
		}
	}
	return gameUser

}

func stringInList(list []string, candidate string) bool {
	for _, s := range list {
		if s == candidate {
			return true
		}
	}
	return false
}

func getRolesForGuild(g discordgo.Guild, s discordgo.Session, gid string) []discordgo.Role {
	var rolesToReturn []discordgo.Role

	roles, _ := s.GuildRoles(gid)

	for _, role := range roles {
		var r discordgo.Role = *role
		rolesToReturn = append(rolesToReturn, r)
	}

	return rolesToReturn
}

func userIsInClan(s discordgo.Session, gid, uid string) bool {
	mem, _ := s.GuildMember(gid, uid)
	var member discordgo.Member = *mem

	return stringInList(member.Roles, kingsGambitClanMemberRoleID)
}
