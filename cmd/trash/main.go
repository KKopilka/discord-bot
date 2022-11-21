package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

var discordBotToken string

var botID string

const WasteBasketEmoji = "🗑"

func main() {
	err := readBotToken()

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	goBot, err := discordgo.New("Bot " + discordBotToken)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	user, err := goBot.User("@me")
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	botID = user.ID

	err = goBot.Open()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println("Привет!")
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
				err = processGuilds(goBot)
				if err != nil {
					fmt.Println(err.Error())
					return
				}
			}
		}
	}()

	go func() {
		for {
			select {
			case <-done:
				return
			case t := <-ticker2.C:
				fmt.Println("Tick2 at", t)
				err = checkPolls(goBot)
				if err != nil {
					fmt.Println(err.Error())
					return
				}
			}
		}
	}()

	stopch := make(chan os.Signal, 1)
	signal.Notify(stopch, os.Interrupt, syscall.SIGTERM)
	<-stopch

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

func processGuilds(goBot *discordgo.Session) error {
	guilds, err := goBot.UserGuilds(100, "", "")

	if err != nil {
		return err
	}

	for _, guild := range guilds {
		// fmt.Println(guild.Name, guild.ID)
		if err := removeTrash(goBot, guild.ID); err != nil {
			fmt.Println(err.Error())
		}
	}
	return nil
}

func checkPolls(goBot *discordgo.Session) error {
	guilds, err := goBot.UserGuilds(100, "", "")

	if err != nil {
		return err
	}

	for _, guild := range guilds {
		// fmt.Println(guild.Name, guild.ID)
		if err := checkThreads(goBot, guild.ID); err != nil {
			fmt.Println(err.Error())
		}
	}
	return nil
}

func checkThreads(goBot *discordgo.Session, guildID string) error {
	threads, err := goBot.GuildThreadsActive(guildID)

	if err != nil {
		return err
	}

	for _, thread := range threads.Threads {
		fmt.Println("Thread", thread.Name, thread.ID, thread.Type)
		transformPolls(goBot, thread)
	}
	return nil
}

func transformPolls(goBot *discordgo.Session, channel *discordgo.Channel) error {
	messages, err := goBot.ChannelMessages(channel.ID, 10, "", "", channel.LastMessageID)

	if err != nil {
		return err
	}

	for _, message := range messages {
		if botID != "" && message.Author.ID != botID {
			fmt.Println("tm", message.ID, message.Author, message.Timestamp, message.Content)

			if strings.Index(message.Content, "https://steamcommunity.com/") >= 0 {

				if err := createPoll(goBot, message.ChannelID, message.Content); err != nil {
					fmt.Println(err.Error())
					continue
				}

			} else {
				if err := notifyAuthor(goBot, message); err != nil {
					fmt.Println(err.Error())
					continue
				}
			}

			if err := goBot.ChannelMessageDelete(message.ChannelID, message.ID); err != nil {
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
