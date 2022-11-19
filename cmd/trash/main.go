package main

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

var discordBotToken = "MTAwMjI4MTI1MzM5Mjg4Nzg3OQ.GnmLD-.irU9_tc-iOxdySzpnrk6QGZC3CtJTRRZWNNIpE"

func main() {
	goBot, err := discordgo.New("Bot " + discordBotToken)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	err = goBot.Open()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println("Привет!")
	guilds, err := goBot.UserGuilds(100, "", "")

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	for _, guild := range guilds {
		fmt.Println(guild.Name, guild.ID)
		if err := removeTrash(goBot, guild.ID); err != nil {
			fmt.Println(err.Error())
		}
	}

}

func removeTrash(goBot *discordgo.Session, guildID string) error {
	channels, err := goBot.GuildChannels(guildID)

	if err != nil {
		return err
	}

	for _, channel := range channels {
		fmt.Println(channel.Name, len(channel.Messages), channel.ID)
		if IsChannelNeedToClean(channel.Name) {
			messages, err := goBot.ChannelMessages(channel.ID, 100, "", "", "")

			if err != nil {
				fmt.Println(err.Error())
				continue
			}

			if len(messages) > 3 {
				for _, message := range messages {
					if err := goBot.ChannelMessageDelete(channel.ID, message.ID); err != nil {
						fmt.Println(err.Error())
						continue
					}
				}
			}
		}
	}

	return nil
}

func IsChannelNeedToClean(channelName string) bool {
	return channelName == "trash"
}
