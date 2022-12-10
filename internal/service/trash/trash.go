package trash

import (
	"fmt"
	"strings"
	"time"

	"github.com/KKopilka/discord-bot/internal/service"
	"github.com/bwmarrin/discordgo"
)

const WasteBasketEmoji = "🗑"

var trashChannelMarkers = []string{
	WasteBasketEmoji,
	"trash",
}

func NewAction() service.Action {
	return processGuilds
}

func processGuilds(s *service.Service) error {
	for _, guild := range s.BotGuilds() {
		// fmt.Println(guild.Name, guild.ID)
		if err := removeTrash(s.BotSession(), guild.ID); err != nil {
			fmt.Println(err.Error())
		}
	}
	return nil
}

func IsChannelNeedToClean(channelName string) bool {
	for _, marker := range trashChannelMarkers {
		if strings.Contains(channelName, marker) {
			return true
		}
	}

	return false
}

func removeTrash(goBot *discordgo.Session, guildID string) error {
	channels, err := goBot.GuildChannels(guildID)

	if err != nil {
		return err
	}

	for _, channel := range channels {
		// fmt.Println(channel.Name, len(channel.Messages), channel.ID)
		if IsChannelNeedToClean(channel.Name) {
			messages, err := goBot.ChannelMessages(channel.ID, 100, "", "", "")

			if err != nil {
				fmt.Println(err.Error())
				continue
			}

			if len(messages) > 3 {
				messageToDelete := []string{}
				messageToDeleteBulk := []string{}
				for _, message := range messages {
					if message.Timestamp.Sub(time.Now().Add(time.Hour*24*-14)) >= 0 {
						messageToDeleteBulk = append(messageToDeleteBulk, message.ID)
					} else {
						messageToDelete = append(messageToDelete, message.ID)
					}
				}

				if err := goBot.ChannelMessagesBulkDelete(channel.ID, messageToDeleteBulk); err != nil {
					fmt.Println(err.Error())
				}

				for _, messageID := range messageToDelete {
					if err := goBot.ChannelMessageDelete(channel.ID, messageID); err != nil {
						fmt.Println(err.Error())
					}
				}
			}
		}
	}
	return nil
}
