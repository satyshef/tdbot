package tdbot

import (
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
		chatType = chat.TypePrivate
	default:
		chatType = "unknown"
	}
	return chatType
}
