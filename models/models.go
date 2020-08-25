package models

//GameUser holds a game name, a username and a user id
type GameUser struct {
	Game              string `json:"game"`
	UserName          string `json:"username"`
	UserID            string `json:"userid"`
	IsPlayingDestiny2 bool   `json:"isplayingdestiny2"`
	IsInClan          bool   `json:"isinclan"`
	MonthlySeen       bool   `json:"monthlyseen"`
	ChannelID         string `json:"channel_id"`
	Deaf              bool   `json:"deaf"`
}

//IsEmpty returns true if the GameUser struct is empty
func (gu GameUser) IsEmpty() bool {
	return gu.UserID == ""
}
