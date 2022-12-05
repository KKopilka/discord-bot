package service

import (
	"fmt"

	"github.com/KKopilka/discord-bot/internal/commands"
	"github.com/bwmarrin/discordgo"
)

type Service struct {
	botSession                  *discordgo.Session
	botToken                    string
	botUser                     *discordgo.User
	botGuilds                   []*discordgo.UserGuild
	registerBotCommandsInGuilds bool
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
	if s.registerBotCommandsInGuilds {
		commands.UnregisterBotCommands(s.botSession)
	}

	return s.CloseSession()
}
