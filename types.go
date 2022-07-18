package tdbot

import (
	"github.com/satyshef/go-tdlib/tdlib"
	"github.com/satyshef/tdbot/user"
)

const (
	EventNameSendMessageByPhone = "sendMessageByPhone"
	EventNameBotReady           = "botReady"
	EventNameChatMemberLeft     = "chatMemberLeft"
)

type botStatus string

const (

	//-----------------------------------------------

	// StatusReady ...
	StatusReady botStatus = "ready"
	// StatusStopped ...
	StatusStopping botStatus = "stopping"
	StatusStopped  botStatus = "stopped"
	// StatusInit ...
	StatusInit botStatus = "init"
	// StatusTimeout
	//StatusTimeout botStatus = "timeout"
	// StatusRestart
	StatusRestart botStatus = "restart"
)

const (
	ErrorCodeSystem           = 551
	ErrorCodeWrongData        = 552
	ErrorCodeContactDuplicate = 553
	ErrorCodeUserNotExists    = 554
	ErrorCodeUserExists       = 555
	ErrorCodeNotInit          = 556
	ErrorCodeChatMemberLeft   = 557
)

//ConvertUser конвертируем пользователя tdlib в собственную структуру
//@orig структура пользователя tdlib
func ConvertUser(orig *tdlib.User) *user.User {

	usr := user.New(orig.FirstName, orig.PhoneNumber, user.TypeTelegram)

	switch orig.Status.GetUserStatusEnum() {
	case "userStatusOffline":
		usr.Status = user.StatusOffline
		usr.WasOnline = orig.Status.(*tdlib.UserStatusOffline).WasOnline
	case "userStatusOnline":
		if usr.Status != user.StatusRestricted {
			usr.Status = user.StatusReady
		}
	default:
		usr.Status = user.StatusUnknown
	}

	usr.ID = orig.ID
	usr.FirstName = orig.FirstName
	usr.LastName = orig.LastName
	usr.Addr = orig.Username

	//WasOnline:   ???? //доделать получение времени активности

	return usr
}
