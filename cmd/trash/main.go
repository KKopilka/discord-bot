package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/KKopilka/discord-bot/internal/config"
	"github.com/KKopilka/discord-bot/internal/service"
	rollsassign "github.com/KKopilka/discord-bot/internal/service/rolls-assign"
)

func main() {
	// 1. Загрузка конфигурации пакет config
	if err := config.ReadConfig(); err != nil {
		fmt.Println(err.Error())
		return
	}

	discordBotToken := config.BotToken()

	fmt.Println("Configuration readed successfully. Create and start bot service.", discordBotToken, "lol")
	// 2. Структура сервиса бота пакет service
	botService, err := service.New(discordBotToken, false)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer botService.Stop()

	fmt.Println("Bot service started", "bot.id:", botService.BotId())
	// 3. Рутинные задачи бота
	// botService.RunAction(trash.NewAction(), 5*time.Second)
	// botService.RunAction(polls.NewAction(), 10*time.Second)
	botService.RunAction(rollsassign.NewAction(), 10*time.Second)

	// ожидание закрытия программы
	stopch := make(chan os.Signal, 1)
	fmt.Println("Waiting for SIGTERM")
	signal.Notify(stopch, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-stopch

	fmt.Println("SIGTERM received")
	fmt.Println("Good bye honney")
}
