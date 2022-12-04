package service

import (
	"github.com/KKopilka/discord-bot/internal/commands"
	"github.com/bwmarrin/discordgo"
)

type Service struct {
	botSession *discordgo.Session
	botToken   string
	botUser    *discordgo.User
	botGuilds  []*discordgo.UserGuild
}

// New create and start new service
func New(botToken string) (*Service, error) {
	goBot, err := discordgo.New("Bot " + botToken)
	if err != nil {
		return nil, err
	}

	// TODO: logger > fmt.Println("Привет!")
	user, err := goBot.User("@me")
	if err != nil {
		return nil, err
	}

	service := &Service{
		botToken:   botToken,
		botSession: goBot,
		botUser:    user,
	}

	// Maybe Start() method ? ---->
	commands.BindCommandHandlers(service.botSession)
	err = service.botSession.Open()
	if err != nil {
		return nil, err
	}

	// TODO: periodic update
	botGuilds, err := FetchSessionGuilds(goBot)
	if err != nil {
		service.CloseSession()
		return nil, err
	}
	service.botGuilds = botGuilds
	service.RegisterBotCommands()
	// TODO: unregister commands only for selected guild
	// <--- Start() method?

	return service, nil
}

// BotSession is temporary. TODO: delete after actions implemented
func (s *Service) BotSession() *discordgo.Session {
	return s.botSession
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
	commands.UnregisterBotCommands(s.botSession)
	return s.CloseSession()
}
