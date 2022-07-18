package tdbot

import (
	"strings"

	tdc "github.com/satyshef/go-tdlib/client"
	"github.com/satyshef/go-tdlib/tdlib"
	"github.com/satyshef/tdbot/events/event"
	"github.com/satyshef/tdbot/profile"
)

// обработчик событий Телеграм клиента
func (bot *Bot) eventCatcher(tdEvent *tdc.SystemEvent) *tdlib.Error {
	//bot.Logger.Infof("NEW EVENT : %#v\n", tdEvent)
	if tdEvent == nil {
		return tdlib.NewError(tdc.ErrorCodeSystem, "CLIENT_EMPTY_UPDATE", "Received an empty update to client")
	}

	if bot.Profile == nil {
		go bot.Restart()
		return tdlib.NewError(profile.ErrorCodeDirNotExists, "PROFILE_NOT_INIT", "Profile not init")
	}

	if bot.Client == nil {
		return tdlib.NewError(tdc.ErrorCodeNotInit, "CLIENT_NOT_INIT", "Client not init")
	}
	// Проверяем существование директории профиля. Решить что делать при отсутcтвии директории
	if !bot.Profile.CheckDir() {
		//bot.Profile.User.Status = user.StatusNotExist
		//go bot.Stop()
		return tdlib.NewError(profile.ErrorCodeDirNotExists, "PROFILE_DIR_NOT_EXISTS", "Profile Dir Not Exists")
	}

	ev := event.New(string(tdEvent.Type), tdEvent.Name, tdEvent.Time, tdEvent.DataJSON())

	switch tdEvent.Type {
	case tdc.EventTypeRequest:
		// если запрос то сначала проверяем лимиты затем пишим событие
		if err := bot.CheckEventLimits(ev); err != nil {
			return err
		}
		if err := bot.Profile.Event.Write(ev); err != nil && !strings.Contains(err.Error(), "Event not observed") {
			bot.Logger.Errorln(err)
			bot.Stop()
			return tdlib.NewError(profile.ErrorCodeLimitExceeded, "PROFILE_EVENT_DONT_WRITE", err.Error())
		}
	case tdc.EventTypeResponse,
		tdc.EventTypeError:
		// если ответ или ошибка то сначала пишим событие а затем проверяем лимиты
		err := bot.Profile.Event.Write(ev)
		if err := bot.CheckEventLimits(ev); err != nil {
			bot.Logger.Errorf("LIMIT %#v\n", err)
			return nil
		}
		if err != nil && !strings.Contains(err.Error(), "Event not observed") {
			bot.Logger.Errorln(err)
			bot.Stop()
			return tdlib.NewError(profile.ErrorCodeLimitExceeded, "PROFILE_EVENT_DONT_WRITE", err.Error())
		}
	}

	return nil
}
