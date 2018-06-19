package model

type MMSlashCommand struct {
	ChannelId   string `schema:"channel_id"`
	ChannelName string `schema:"channel_name"`
	Command     string `schema:"command"`
	TeamName    string `schema:"team_domain"`
	TeamId      string `schema:"team_id"`
	Text        string `schema:"text"`
	Token       string `schema:"token"`
	UserId      string `schema:"user_id"`
	Username    string `schema:"user_name"`
}
