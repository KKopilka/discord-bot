package commands

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/bwmarrin/discordgo"
)

type commandHandlerFunc func(s *discordgo.Session, i *discordgo.InteractionCreate)

var rollMinValue = float64(2)
var randSource = rand.NewSource(time.Now().UnixNano())

func NewDiceAppCommand(name string, maxValue int) *discordgo.ApplicationCommand {
	commandHandlers[name] = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		// maybe /dice command
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Roll number from 1 to %d: %d", maxValue, RollValue(1, maxValue)),
			},
		})
	}

	return &discordgo.ApplicationCommand{
		Name:        name,
		Description: fmt.Sprintf("Roll number from 1 to %d.", maxValue),
		// NameLocalizations: &map[discordgo.Locale]string{
		// 	discordgo.Russian: fmt.Sprintf("roll-dice-%d", maxValue),
		// },
		// DescriptionLocalizations: &map[discordgo.Locale]string{
		// 	discordgo.Russian: "Заебали эти челы в дискорде", //是一个本地化的命令
		// },
	}
}

var commands = []*discordgo.ApplicationCommand{
	{
		Name:        "roll",
		Description: "Roll number from 1 to N.",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "max-value",
				Description: "max or random value",
				MinValue:    &rollMinValue,
				Required:    true,
			},
		},
	},
	NewDiceAppCommand("dice", 6),
	NewDiceAppCommand("d4", 4),
	NewDiceAppCommand("d6", 6),
	NewDiceAppCommand("d8", 8),
	NewDiceAppCommand("d10", 10),
	NewDiceAppCommand("d18", 18),
	NewDiceAppCommand("d20", 20),
}

var commandHandlers = map[string]commandHandlerFunc{
	"roll": RollCommand,
	"dice": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		// maybe /dice command
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Roll number from 1 to 6: %d", RollValue(1, 6)),
			},
		})
	},
}

var registeredCommands = make([]*discordgo.ApplicationCommand, len(commands))

func RollValue(minVal int, maxVal int) int {
	return rand.New(randSource).Intn(maxVal-minVal) + minVal
}

func ParseCommandOptions(i *discordgo.InteractionCreate) map[string]*discordgo.ApplicationCommandInteractionDataOption {
	// Access options in the order provided by the user.
	options := i.ApplicationCommandData().Options

	// Or convert the slice into a map
	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
	for _, opt := range options {
		optionMap[opt.Name] = opt
	}

	return optionMap
}

func sendError(s *discordgo.Session, i *discordgo.InteractionCreate, err error) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf(":x: %s", err.Error()),
		},
	})
}

func RollCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	optionMap := ParseCommandOptions(i)
	maxValue := 0
	if v, ok := optionMap["max-value"]; ok {
		maxValue = int(v.IntValue())
	}

	if maxValue < 2 {
		// TODO: throw error
		sendError(s, i, errors.New("max-value must be greater than 2"))
		return
	}

	// maybe /dice command
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("Roll number from 1 to %d: %d", maxValue, RollValue(1, maxValue)),
		},
	})
}

func BindCommandHandlers(s *discordgo.Session) {
	// здесь назначаются все обработчики комманд
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})

}

func RegisterBotCommands(s *discordgo.Session, guildID string) {
	for i, v := range commands {
		cmd, err := s.ApplicationCommandCreate(s.State.User.ID, guildID, v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
		registeredCommands[i] = cmd
	}
}

func UnregisterAllGuildCommands(s *discordgo.Session, guildID string) error {
	cmds, err := s.ApplicationCommands(s.State.User.ID, guildID)
	if err != nil {
		return err
	}
	for _, cmd := range cmds {
		err := s.ApplicationCommandDelete(cmd.ApplicationID, cmd.GuildID, cmd.ID)
		if err != nil {
			log.Panicf("Cannot delete '%v' command: %v", cmd.Name, err)
		}
	}
	return nil
}

func UnregisterBotCommands(s *discordgo.Session) {
	for i, v := range registeredCommands {
		err := s.ApplicationCommandDelete(v.ApplicationID, v.GuildID, v.ID)
		if err != nil {
			log.Panicf("Cannot delete '%v' command: %v", v.Name, err)
		}
		registeredCommands[i] = nil
	}
}
