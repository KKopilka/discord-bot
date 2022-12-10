package config

import "io/ioutil"

type AppConfig struct {
}

var discordBotToken string

func ReadConfig() error {
	file, err := ioutil.ReadFile("bot-token")

	if err != nil {
		return err
	}

	botToken := DiscordBotToken(string(file))
	botToken.Sanitize()

	if err := botToken.Validate(); err != nil {
		return err
	}

	discordBotToken = string(botToken)

	return nil
}

func BotToken() string {
	return discordBotToken
}
