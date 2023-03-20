package rollsassign

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/KKopilka/discord-bot/internal/service"
	ramsg "github.com/KKopilka/discord-bot/internal/service/rolls-assign/message"
	"github.com/bwmarrin/discordgo"
	emj "github.com/enescakir/emoji"
)

var roleAssignMessages = make(map[string][]*ramsg.RoleAssignMessage)

func NewAction() service.Action {
	return processGuilds
}

func processGuilds(s *service.Service) error {
	for _, guild := range s.BotGuilds() {
		// fmt.Println(guild.Name, guild.ID)
		if err := checkChannels(s, guild.ID); err != nil {
			fmt.Println(err.Error())
		}
	}
	return nil
}

func IgnoreChannel(channelID string) bool {
	if channelID == "1049105231402766356" ||
		channelID == "1001918395807186964" {
		return false
	}

	if channelID == "" {
		return true
	}

	return false
}

func IsMessageContains(substr string, message *discordgo.Message) bool {
	return strings.Contains(message.Content, substr)
}

func SearchMessage(goBot *discordgo.Session, channel *discordgo.Channel, substr string, fromMessageID string) *discordgo.Message {
	return searchMessage(goBot, channel, substr, fromMessageID, 0)
}

// searchMessage ищет сообщение с содержанием substr в канале channel
// этот метод очень долго перебирает все сообщения в канале, до момента пока не найдет нужное сообщение.
// вообще нужно сделать ограничитель, чтобы бот:
// a) читал определенное кол-во сообщений максимум (ограничение по уровню)
// б) запоминал до какого-то определенного id, что сообещния такого не было. (searchHistory)
// в) посмотреть может есть у дискорд API функция поиска сообщения.
func searchMessage(goBot *discordgo.Session, channel *discordgo.Channel, substr string, fromMessageID string, searchLevel int) *discordgo.Message {
	var messages []*discordgo.Message
	var err error
	outPrefTab := strings.Repeat("\t", searchLevel)

	if fromMessageID == "" {
		fmt.Println(fmt.Sprintf("%sSearch#%d in messages for:", outPrefTab, searchLevel), substr, "around last messageID:", channel.LastMessageID)
		messages, err = goBot.ChannelMessages(channel.ID, 100, "", "", channel.LastMessageID)
	} else {
		fmt.Println(fmt.Sprintf("%sSearch#%d in messages for:", outPrefTab, searchLevel), substr, "from last messageID:", fromMessageID)
		messages, err = goBot.ChannelMessages(channel.ID, 100, fromMessageID, "", "")
	}
	if err != nil {
		fmt.Println("err:", err)
		return nil
	}

	lastMessageId := channel.LastMessageID
	isLastPage := len(messages) < 100
	for _, message := range messages {
		lastMessageId = message.ID
		if IsMessageContains(substr, message) {
			return message
		}
	}

	fmt.Printf("%s#%d Content not found in messages\n", outPrefTab, searchLevel)
	if !isLastPage {
		searchLevel++

		return searchMessage(goBot, channel, substr, lastMessageId, searchLevel)
	}

	return nil

}

func AddRoleAssignMessage(guildID string, rm *ramsg.RoleAssignMessage) {
	roleAssignMessages[guildID] = append(roleAssignMessages[guildID], rm)
}

// func ReactOn(i interface{}) {
// 	if m, ok := i.(*discordgo.MessageReactionAdd); ok {

// 	}
// }

func checkChannels(s *service.Service, guildID string) error {
	goBot := s.BotSession()

	channels, err := goBot.GuildChannels(guildID)

	if err != nil {
		return err
	}
	// Проверяем все каналы в гильдиях
	for _, channel := range channels {
		fmt.Println(channel.Name, len(channel.Messages), channel.ID)
		if !IgnoreChannel(channel.ID) {
			// Ищем сообщения с тегом #role-a$$ign
			if message := SearchMessage(goBot, channel, "#role-a$$ign", ""); message != nil {
				// Нашли сообщение
				fmt.Println("Found message:", message.ID, ">>", message.Content)
				rm := ramsg.ParseRoleAssignMessage(message.Content)
				if rm == nil {
					fmt.Println("RoleAssign config not found in message:", message.ID)
					continue
				}

				// AddRoleAssignMessage(guildID, rm)

				for _, roleConf := range rm.Roles {
					em := emj.Parse(roleConf.EmojiChar)
					usrRoleReact, err := goBot.MessageReactions(message.ChannelID, message.ID, em, 100, "", "")
					if err != nil {
						fmt.Println("MessageReactions for emoji:", em, "fetch err:", err)
						continue
					}

					// check if guild has this role (findRoleInGuild)
					guildRoles, err := goBot.GuildRoles(guildID)
					if err != nil {
						fmt.Println("GuildRoles fetch err:", err)
						continue
					}

					var reactRole *discordgo.Role
					for _, gRole := range guildRoles {
						if roleConf.Name == gRole.Name {
							reactRole = gRole
							break
						}
					}

					if reactRole == nil {
						// role not found
						// try to create role
						roleName := roleConf.Name
						roleColor := 0
						roleHoist := true

						if r, err := strconv.ParseInt(roleConf.Color, 16, 64); err != nil {
							fmt.Println("CreateRoleParams color convert err:", err)
							continue
						} else {
							roleColor = int(r)
						}

						reactRole, err = goBot.GuildRoleCreate(guildID, &discordgo.RoleParams{
							Name:  roleName,
							Color: &roleColor,
							Hoist: &roleHoist,
						})
						if err != nil {
							fmt.Println("CreateRoleParams color convert err:", err)
							continue
						}
					}
					// check role again
					if reactRole == nil {
						fmt.Println("Can not find role:", roleConf.Name)
						continue
					}

					if len(usrRoleReact) == 0 {
						if err := goBot.MessageReactionAdd(message.ChannelID, message.ID, em); err != nil {
							fmt.Println("MessageReactionAdd err:", err)
						}
					}
				}
			} else {
				fmt.Println("Message not found in channel:", channel.ID)
			}
		}
	}

	return nil
}
