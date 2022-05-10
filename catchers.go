package tdbot

import (
	"strings"

	"github.com/polihoster/tdbot/events/event"
	"github.com/polihoster/tdbot/profile"
	"github.com/polihoster/tdlib"
)

// обработчик событий Телеграм клиента
func (bot *Bot) eventCatcher(tdEvent *tdlib.SystemEvent) *tdlib.Error {
	//bot.Logger.Infof("NEW EVENT %s: %#v\n", eventType, update)
	if tdEvent == nil {
		return tdlib.NewError(tdlib.ErrorCodeSystem, "CLIENT_EMPTY_UPDATE", "Received an empty update to client")
	}
	if bot.Profile == nil {
		go bot.Restart()
		return tdlib.NewError(profile.ErrorCodeDirNotExists, "PROFILE_NOT_INIT", "Profile not init")
	}
	if bot.Client == nil {
		return tdlib.NewError(tdlib.ErrorCodeNotInit, "CLIENT_NOT_INIT", "Client not init")
	}
	// Проверяем существование директории профиля. Решить что делать при отсутcтвии директории
	if !bot.Profile.CheckDir() {
		//bot.Profile.User.Status = user.StatusNotExist
		//go bot.Stop()
		return tdlib.NewError(profile.ErrorCodeDirNotExists, "PROFILE_DIR_NOT_EXISTS", "Profile Dir Not Exists")
	}
	ev := event.New(string(tdEvent.Type), tdEvent.Name, tdEvent.Time, tdEvent.DataJSON())
	if err := bot.CheckEventLimits(ev); err != nil {
		return err
	}
	if err := bot.Profile.Event.Write(ev); err != nil && !strings.Contains(err.Error(), "Event not observed") {
		bot.Logger.Errorln(err)
		bot.Stop()
		return tdlib.NewError(profile.ErrorCodeLimitExceeded, "PROFILE_EVENT_DONT_WRITE", err.Error())
	}
	return nil
}
