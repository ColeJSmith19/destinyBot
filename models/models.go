package models

//GameUser holds a game name, a username and a user id
type GameUser struct {
	Game              string `json:"game"`
	UserName          string `json:"username"`
	UserID            string `json:"userid"`
	IsPlayingDestiny2 bool   `json:"isplayingdestiny2"`
}
