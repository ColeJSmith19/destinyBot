package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

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

//monthlySeenRoleID ID of the MonthlySeen Role
// const monthlySeenRoleID = "747599821581320262"

//This is the October Seen ID
const monthlySeenRoleID = "761288604357230613"

//monthlyUneenRoleID ID of the MonthlyUnseen Role
const monthlyUnseenRoleID = "747599618954756217"

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
	dg.AddHandler(voiceStateUpdate)

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
// func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
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
	if len(text) == 1 {
		s.ChannelMessageSend(m.ChannelID, "Try '!TR3VR help'")
	} else if len(text) == 2 && text[1] == "help" {
		s.ChannelMessageSend(m.ChannelID, "Here is a list of commands\nhelp\nunseen")
	} else if len(text) == 2 && text[1] == "unseen" {
		unseenUsers := getMonthlyUnseenUsers(s, m.GuildID)
		if unseenUsers == "" {
			s.ChannelMessageSend(m.ChannelID, "All members have been seen!")
		} else {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("The following members have not been seen yet this month.\n%+v", unseenUsers))
		}
	}
}

func voiceStateUpdate(s *discordgo.Session, v *discordgo.VoiceStateUpdate) {

	time.Sleep(time.Minute * 5)

	g, _ := s.Guild(v.GuildID)

	// fmt.Println(getRolesForGuild(*g, *s, v.GuildID))

	var gameUsers []models.GameUser

	for _, voiceStates := range g.VoiceStates {
		var vs discordgo.VoiceState = *voiceStates

		// fmt.Printf("voiceStates\n%+v\n\n", vs)

		//skip any users in channelsToIgnore const
		if stringInList(channelsToIgnore, vs.ChannelID) {
			continue
		}

		//skip any users NOT in the guild King's Gambit
		if vs.GuildID != kingsGambitGuildID {
			continue
		}

		//skip any users WITHOUT the role King's Gambit Clan Member
		if !userIsInClan(*s, v.GuildID, vs.UserID) {
			continue
		}

		gameUser := getPresenceForUser(*g, *s, vs.UserID)

		if gameUser.IsEmpty() {
			continue
		}

		gameUser.ChannelID = vs.ChannelID
		gameUsers = append(gameUsers, gameUser)
	}

	// b, e := json.Marshal(gameUsers)
	// if e != nil {
	// 	fmt.Println(e.Error())
	// }
	// f, e := os.Open("test.json")
	// if e != nil {
	// 	fmt.Println(e)
	// }
	// defer f.Close()

	// e = ioutil.WriteFile("test.json", b, 0644)
	// if e != nil {
	// 	fmt.Println(e)
	// }

	// fmt.Printf("gameUsers\n%+v\n\n", gameUsers)

	if len(gameUsers) < 2 {
		return
	}

	var channelUsers = make(map[string][]models.GameUser)

	for _, gu := range gameUsers {
		channelUsers[gu.ChannelID] = append(channelUsers[gu.ChannelID], gu)
	}

	// fmt.Printf("channelUsers\n%+v\n\n", channelUsers)

	for _, users := range channelUsers {
		if len(users) < 2 {
			continue
		}

		var usersPlayingDestiny bool
		var clanCount int
		for _, user := range users {
			if user.IsPlayingDestiny2 {
				usersPlayingDestiny = true
			}
			if user.IsInClan {
				clanCount++
			}

		}
		var usersToNoteAsSeen string
		var usersToMarkAsSeenCount int
		//if the users in the channel are playing destiny 2 and the number of user in the channel is
		//greater than or equal to the number of users that are in the clan (at least two at this step)
		if usersPlayingDestiny && clanCount <= len(users) {
			for _, user := range users {
				if user.IsInClan && !user.MonthlySeen {
					s.GuildMemberRoleRemove(v.GuildID, user.UserID, monthlyUnseenRoleID)
					s.GuildMemberRoleAdd(v.GuildID, user.UserID, monthlySeenRoleID)
					usersToNoteAsSeen += user.UserName + ", "
					usersToMarkAsSeenCount++
				}
			}
		}

		if usersToNoteAsSeen != "" {
			if usersToMarkAsSeenCount == 1 {
				s.ChannelMessageSend(botTestingChannelID, fmt.Sprintf("%s has been marked as seen for the month!", strings.Trim(usersToNoteAsSeen, ", ")))
			} else {
				s.ChannelMessageSend(botTestingChannelID, fmt.Sprintf("%s have been marked as seen for the month!", strings.Trim(usersToNoteAsSeen, ", ")))
			}
		}
	}

}

func getPresenceForUser(g discordgo.Guild, s discordgo.Session, uid string) models.GameUser {
	var gameUser models.GameUser

	for _, p := range g.Presences {
		var pre discordgo.Presence = *p

		if pre.User.ID == uid {
			user, _ := s.User(pre.User.ID)
			if pre.Game != nil {
				gameUser.Game = pre.Game.Name
				gameUser.IsPlayingDestiny2 = pre.Game.ApplicationID == destiny2AppID
			}
			gameUser.UserName = user.String()
			gameUser.UserID = user.ID
			gameUser.IsInClan = userIsInClan(s, g.ID, user.ID)
			gameUser.MonthlySeen = userHasBeenSeen(s, g.ID, user.ID)
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

func userHasBeenSeen(s discordgo.Session, gid, uid string) bool {
	mem, _ := s.GuildMember(gid, uid)
	var member discordgo.Member = *mem

	return stringInList(member.Roles, monthlySeenRoleID)
}

func getMonthlyUnseenUsers(s *discordgo.Session, gid string) string {
	unseenMembers := ""
	members, e := s.GuildMembers(gid, "", 1000)
	if e != nil {
		fmt.Println(e)
		return ""
	}
	for _, member := range members {
		if stringInList(member.Roles, kingsGambitClanMemberRoleID) && !stringInList(member.Roles, monthlySeenRoleID) {
			unseenMembers += member.User.Username + ", "
		}
	}
	return strings.Trim(unseenMembers, ", ")
}
