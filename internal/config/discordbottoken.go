package config

import (
	"errors"
	"fmt"
	"strings"
)

type DiscordBotToken string

func (bt *DiscordBotToken) Sanitize() {
	content := string(*bt)
	content = strings.Trim(content, " \n\r\t")
	// TODO: check bot-token specification for whitespaces inside it.
	content = strings.ReplaceAll(content, " ", "")

	*bt = DiscordBotToken(content)
}

func (bt DiscordBotToken) Validate() error {
	content := string(bt)

	if strings.Contains(content, " ") {
		return errors.New("whitespace in bot-token")
	}
	if strings.Contains(content, "\n") {
		return errors.New("end of line in bot-token")
	}

	if len(content) != 72 {
		return errors.New(fmt.Sprintln("invalid bot-token. Token length:", len(content)))
	}

	return nil
}
