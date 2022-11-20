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

	err = goBot.Open()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println("Привет!")
	ticker := time.NewTicker(30 * time.Second)
	done := make(chan bool)
	go func() {
		for {
			select {
			case <-done:
				return
			case t := <-ticker.C:
				fmt.Println("Tick at", t)
				err = processGuilds(goBot)
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
		fmt.Println(channel.Name, len(channel.Messages), channel.ID)
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
		fmt.Println(guild.Name, guild.ID)
		if err := removeTrash(goBot, guild.ID); err != nil {
			fmt.Println(err.Error())
		}
	}
	return nil
}
