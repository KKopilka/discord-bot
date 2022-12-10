package polls

import (
	"fmt"
	"strings"

	"github.com/KKopilka/discord-bot/internal/service"
	"github.com/bwmarrin/discordgo"
)

func NewAction() service.Action {
	return checkPolls
}

func checkPolls(s *service.Service) error {
	for _, guild := range s.BotGuilds() {
		// ignore this guild for some good times
		if guild.ID == "695782620793012225" {
			continue
		}
		// fmt.Println(guild.Name, guild.ID)
		if err := checkThreads(s, guild.ID); err != nil {
			fmt.Println(err.Error())
		}
	}
	return nil
}

func checkThreads(s *service.Service, guildID string) error {
	threads, err := s.BotSession().GuildThreadsActive(guildID)

	if err != nil {
		return err
	}

	for _, thread := range threads.Threads {
		fmt.Println("Thread", thread.Name, thread.ID, thread.Type, thread.LastMessageID)
		if err := transformPolls(s, thread); err != nil {
			fmt.Println(err.Error())
		}
	}
	return nil
}

func transformPolls(s *service.Service, channel *discordgo.Channel) error {
	messages, err := s.BotSession().ChannelMessages(channel.ID, 10, "", "", channel.LastMessageID)
	if err != nil {
		return err
	}
	fmt.Println("ChannelMessages", "id", channel.ID, "last.id", channel.LastMessageID, "len", len(messages))

	for _, message := range messages {
		if s.BotId() != "" && message.Author.ID != s.BotId() {
			fmt.Println("tm", message.ID, message.Author, message.Timestamp, message.Content)

			if strings.Index(message.Content, "https://steamcommunity.com/") >= 0 {

				if err := createPoll(s.BotSession(), message.ChannelID, message.Content); err != nil {
					fmt.Println(err.Error())
					continue
				}

			} else {
				if err := notifyAuthor(s.BotSession(), message); err != nil {
					fmt.Println(err.Error())
					continue
				}
			}

			if err := s.BotSession().ChannelMessageDelete(message.ChannelID, message.ID); err != nil {
				fmt.Println(err.Error())
				continue
			}
		}
	}
	return nil
}

func createPoll(goBot *discordgo.Session, messageChannelID string, content string) error {
	botMessage, err := goBot.ChannelMessageSend(messageChannelID, "Какая мразь прислала этот контент "+content)

	if err != nil {
		return err
	}

	if err := goBot.MessageReactionAdd(botMessage.ChannelID, botMessage.ID, "\U0001F44D"); err != nil {
		return err
	}
	if err := goBot.MessageReactionAdd(botMessage.ChannelID, botMessage.ID, "\U0001F44E"); err != nil {
		return err
	}
	return nil
}

func notifyAuthor(goBot *discordgo.Session, message *discordgo.Message) error {
	channel, err := goBot.UserChannelCreate(message.Author.ID)

	if err != nil {
		return err
	}

	_, err = goBot.ChannelMessageSend(channel.ID, "Ну ты и уёбок конечно. Перепиши сообщение блять нормально "+message.Content)

	return err
}
