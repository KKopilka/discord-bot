package message

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/bwmarrin/discordgo"
)

type RoleConf struct {
	EmojiChar string
	Emoji     string
	Name      string
	Color     string
}
type RoleAssignMessage struct {
	Message *discordgo.Message
	Roles   []RoleConf
}

// replaceEmojiStr replace emoji char with \uXXXXXXXX formatted string.
func replaceEmojiStr(emoji string) string {
	//sanitization
	// re := regexp.MustCompile(`[#*0-9]\x{FE0F}?\x{20E3}|©\x{FE0F}?|...`)
	// omji := re.Find(emoji)
	r, _ := utf8.DecodeRuneInString(emoji)

	return fmt.Sprintf("\\U%08X", r)

	// out := ""
	// for _, r := range omji {

	// 	out += fmt.Sprintf("\\u%08X", r2)
	// }
	// // conveting
	// return out
}

func ParseRoleAssignMessage(content string) *RoleAssignMessage {
	lines := strings.Split(content, "\n")
	r := &RoleAssignMessage{
		Roles: make([]RoleConf, 0),
	}

	for _, line := range lines {
		if strings.Contains(line, "-") {
			roleConfStr := strings.Split(line, "-")
			for k, v := range roleConfStr {
				roleConfStr[k] = strings.TrimSpace(v)
			}
			// detect emoji
			emoji := replaceEmojiStr(roleConfStr[0])
			fmt.Println("emoji(", roleConfStr[0], ")", emoji)
			// TODO: check emoji

			// detect color (hash)
			if !strings.Contains(roleConfStr[1], "#") {
				// no role color
				continue
			}

			ncCfg := strings.Split(roleConfStr[1], "#")
			if len(ncCfg[0]) == 0 {
				// no role name
				continue
			}
			if len(ncCfg[1]) == 0 {
				// no role color
				continue
			}
			// TODO: check color

			roleCfg := RoleConf{
				EmojiChar: roleConfStr[0],
				Emoji:     string(emoji),
				Name:      ncCfg[0],
				Color:     ncCfg[1],
			}

			r.Roles = append(r.Roles, roleCfg)
		}
	}

	return r
}
