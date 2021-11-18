package tdbot

import (
	"github.com/polihoster/tdbot/user"
	"github.com/polihoster/tdlib"
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
	//коды ошибок данного модуля начинаются с 3
	ErrorCodeSystem           = 301
	ErrorCodeWrongData        = 302
	ErrorCodeContactDuplicate = 303
	ErrorCodeUserNotExists    = 304
	ErrorCodeNotInit          = 305
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
