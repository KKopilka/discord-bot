package service

import (
	"fmt"
	"time"

	"github.com/KKopilka/discord-bot/internal/commands"
	ramsg "github.com/KKopilka/discord-bot/internal/service/rolls-assign/message"
	"github.com/bwmarrin/discordgo"
	emj "github.com/enescakir/emoji"
)

type Service struct {
	botSession                  *discordgo.Session
	botToken                    string
	botUser                     *discordgo.User
	botGuilds                   []*discordgo.UserGuild
	registerBotCommandsInGuilds bool
	actionTasks                 []*ActionTask
}

// New create and start new service
func New(botToken string, debug bool) (*Service, error) {
	fmt.Println("Initializing discordgo bot session")
	goBot, err := discordgo.New("Bot " + botToken)
	if err != nil {
		return nil, err
	}

	fmt.Println("Loading bot user identity info")
	// TODO: logger > fmt.Println("Привет!")
	user, err := goBot.User("@me")
	if err != nil {
		return nil, err
	}

	service := &Service{
		botToken:                    botToken,
		botSession:                  goBot,
		botUser:                     user,
		registerBotCommandsInGuilds: !debug, // во время отладки не регистрируем команды
	}

	// Maybe Start() method ? ---->
	fmt.Println("Binding bot command handlers")
	commands.BindCommandHandlers(service.botSession)
	service.botSession.AddHandler(func(s *discordgo.Session, m *discordgo.MessageReactionAdd) {
		fmt.Println(
			"guildID:", m.GuildID,
			"channelID:", m.ChannelID,
			"messageID:", m.MessageID,
			"emoji:", m.Emoji.ID, m.Emoji.Name,
		)
		message, err := service.botSession.ChannelMessage(m.ChannelID, m.MessageID)
		if err != nil {
			fmt.Println("ChannelMessage err:", err)
			return
		}

		rm := ramsg.ParseRoleAssignMessage(message.Content)
		if rm == nil {
			fmt.Println("RoleAssign config not found in message:", message.ID)
			return
		}

		var rCfg *ramsg.RoleConf
		// find role by emoji
		for _, roleConf := range rm.Roles {
			if em := emj.Parse(roleConf.EmojiChar); em == m.Emoji.APIName() {
				// role found
				rCfg = &roleConf
				break
			} else {
				fmt.Println("emoji:", m.Emoji, "em:", em)
			}
		}

		if rCfg == nil {
			// role not found
			fmt.Println("role not found")
			return
		}

		guildRoles, err := goBot.GuildRoles(m.GuildID)
		if err != nil {
			fmt.Println("GuildRoles err:", err)
			return
		}

		var assignRole *discordgo.Role
		for _, gRole := range guildRoles {
			if gRole.Name == rCfg.Name {
				assignRole = gRole
			}
		}

		if assignRole == nil {
			// assignRole not found
			fmt.Println("assignRole not found")
			return
		}

		if err := goBot.GuildMemberRoleAdd(m.GuildID, m.UserID, assignRole.ID); err != nil {
			fmt.Println("GuildMemberRoleAdd err:", err)
		} else {
			fmt.Println("Successfully added role", assignRole.ID, assignRole.Name, "to user", m.UserID, m.Member.User.Username)
		}

	})

	fmt.Println("Running bot session")
	err = service.botSession.Open()
	if err != nil {
		return nil, err
	}

	fmt.Println("Loading bot available guilds")
	// TODO: periodic update
	botGuilds, err := FetchSessionGuilds(service.botSession)
	if err != nil {
		service.CloseSession()
		return nil, err
	}
	service.botGuilds = botGuilds
	// во время дебага не регистрируем команды, т.к. это занимает много времени
	if service.registerBotCommandsInGuilds {
		fmt.Println("Registering available bot commands in loaded guilds")
		service.RegisterBotCommands()
	}
	// TODO: unregister commands only for selected guild
	// <--- Start() method?

	return service, nil
}

func (s *Service) BotId() string {
	// return empty string if user not loaded
	if s.botUser == nil {
		return ""
	}

	return s.botUser.ID
}

// BotSession is temporary. TODO: delete after actions implemented
func (s *Service) BotSession() *discordgo.Session {
	return s.botSession
}

// BotGuilds returns slice of current bot joined guilds.
func (s *Service) BotGuilds() []*discordgo.UserGuild {
	// cached!
	return s.botGuilds
}

func (s *Service) RegisterBotCommands() {
	for _, guild := range s.botGuilds {
		commands.UnregisterAllGuildCommands(s.botSession, guild.ID)
		commands.RegisterBotCommands(s.botSession, guild.ID)
	}
}

func FetchSessionGuilds(goBot *discordgo.Session) ([]*discordgo.UserGuild, error) {
	return goBot.UserGuilds(100, "", "")
}

func (s *Service) CloseSession() error {
	return s.botSession.Close()
}

func (s *Service) Stop() error {
	// stop actions tasks
	runnedActions := s.actionTasks
	for _, at := range runnedActions {
		at.Stop()
		fmt.Println("Stopped task for action:", at.ActFuncName())

	}

	if s.registerBotCommandsInGuilds {
		commands.UnregisterBotCommands(s.botSession)
	}

	return s.CloseSession()
}

func (s *Service) RunAction(act Action, d time.Duration) error {
	at := NewActionTask(act, d)
	at.Run(s)
	fmt.Println("Added task for action:", at.ActFuncName(), "ticker:", d)

	s.actionTasks = append(s.actionTasks, at)

	return nil
}
