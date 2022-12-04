package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/KKopilka/discord-bot/internal/service"
	"github.com/bwmarrin/discordgo"
)

var discordBotToken string

const WasteBasketEmoji = "🗑"

func main() {
	// 1. Загрузка конфигурации пакет config
	err := readBotToken()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println("Configuration readed successfully")
	// 2. Структура сервиса бота пакет service
	botService, err := service.New(discordBotToken)
	defer botService.Stop()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println("Bot service started", "bot.id:", botService.BotId())

	// 3. Рутинные задачи бота
	ticker := time.NewTicker(5 * time.Second)
	ticker2 := time.NewTicker(10 * time.Second)
	done := make(chan bool)
	go func() {
		for {
			select {
			case <-done:
				return
			case t := <-ticker.C:
				fmt.Println("Tick1 at", t)
				err = processGuilds(botService)
				if err != nil {
					fmt.Println(err.Error())
					return
				}
			}
		}
	}()
	fmt.Println("action:trash started")

	go func() {
		for {
			select {
			case <-done:
				return
			case t := <-ticker2.C:
				fmt.Println("Tick2 at", t)
				err = checkPolls(botService)
				if err != nil {
					fmt.Println(err.Error())
					return
				}
			}
		}
	}()
	fmt.Println("action:polls started")

	// ожидание закрытия программы
	stopch := make(chan os.Signal, 1)
	fmt.Println("Waiting for SIGTERM")
	signal.Notify(stopch, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-stopch

	// остановка программы, тикеров и циклов го-рутин
	ticker.Stop()
	done <- true

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
				for _, message := range messages {
					messageToDelete = append(messageToDelete, message.ID)
				}

				if err := goBot.ChannelMessagesBulkDelete(channel.ID, messageToDelete); err != nil {
					fmt.Println(err.Error())
					continue
				}
			}
		}
	}
	return nil
}

func IsChannelNeedToClean(channelName string) bool {
	if strings.Index(channelName, WasteBasketEmoji) >= 0 {
		return true
	}

	if strings.Index(channelName, "trash") >= 0 {
		return true
	}
	return false
}

func readBotToken() error {
	file, err := ioutil.ReadFile("bot-token")

	if err != nil {
		return err
	}

	discordBotToken = string(file)
	return nil
}

func processGuilds(s *service.Service) error {
	// guilds, err := goBot.UserGuilds(100, "", "")

	// if err != nil {
	// 	return err
	// }

	for _, guild := range s.BotGuilds() {
		// fmt.Println(guild.Name, guild.ID)
		if err := removeTrash(s.BotSession(), guild.ID); err != nil {
			fmt.Println(err.Error())
		}
	}
	return nil
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
