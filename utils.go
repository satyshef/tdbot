package tdbot

import (
	"strings"

	"github.com/satyshef/go-tdlib/tdlib"
	"github.com/satyshef/tdbot/chat"
)

func GetChatType(tdChat *tdlib.Chat) chat.Type {
	var chatType chat.Type
	switch tdChat.Type.GetChatTypeEnum() {
	case tdlib.ChatTypeBasicGroupType:
		chatType = chat.TypeBasicGroup
	case tdlib.ChatTypeSupergroupType:
		superGroup := tdChat.Type.(*tdlib.ChatTypeSupergroup)
		if superGroup.IsChannel {
			chatType = chat.TypeChannel
		} else {
			chatType = chat.TypeGroup
		}
	case tdlib.ChatTypePrivateType:
		// ?????
		// Боты и пользователи имеют тип private
		chatType = chat.TypePrivate
	default:
		chatType = "unknown"
	}
	return chatType
}

func IsPublicLink(link string) bool {
	if link == "" || strings.Contains(link, "t.me/joinchat/") || strings.Contains(link, "t.me/+") {
		return false
	}
	return true
}

func ExtrctChatName(link string) string {
	chatname := strings.ReplaceAll(link, "https://t.me/", "")
	chatname = strings.ReplaceAll(chatname, "t.me/", "")
	chatname = strings.ReplaceAll(chatname, "@", "")
	return chatname
}
